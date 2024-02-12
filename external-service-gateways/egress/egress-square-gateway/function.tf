locals {
  function_group = "egress-square-gateway"
  function_exclude_list = setunion(
    ["go.sum"],
    fileset("${path.module}", "cmd/**"),         # main harness
    fileset("${path.module}", "**_test.go"),     # go test files
    fileset("${path.module}", "**.tf*"),         # terraform files
    fileset("${path.module}", "**terraform*"),   # terraform files
    fileset("${path.module}", "**terraform/**"), # terraform files
    fileset("${path.module}", "*source-*.zip")   # other source zips
  )

  apis = tomap({
    payment = {
      api   = "EgressSquarePaymentGateway",
      topic = var.square_payment_events_request_topic
    },
    order = {
      api   = "EgressSquareOrderGateway",
      topic = var.square_order_events_request_topic
    },
    customer = {
      api   = "EgressSquareCustomerGateway",
      topic = var.square_customer_events_request_topic
    }
  })
}

resource "google_project_service" "service" {
  for_each = toset([
    "run.googleapis.com",
    "eventarc.googleapis.com",
    "secretmanager.googleapis.com",
    "storage.googleapis.com",
  ])

  project = var.gcp_project_id
  service = each.key

  disable_on_destroy = false
}

# this runs 'go mod vendor' from ${path.module} for inclusion in the source zip file
data "external" "go_mod_vendor" {
  program = ["bash", "../../../go-mod-vendor.sh", "${path.module}"]
}

# random_string.r is to ensure that the zip file gets recreated upon
# running "terraform apply"
resource "random_string" "r" {
  length  = 16
  special = false
}

data "archive_file" "function_source_zip" {
  type = "zip"

  # *source.zip is in .gitignore so should never be committed
  output_path = "${path.module}/${local.function_group}-source.zip"
  source_dir  = path.module

  # this is computed once for all functions
  excludes = local.function_exclude_list

  # this ensures that 'go mod vendor' is run before creating the zip file
  # the dependency on random_string.r is to ensure that the zip file gets
  # recreated upon running "terraform apply"
  depends_on = [data.external.go_mod_vendor, random_string.r]
}

resource "google_storage_bucket_object" "function_source_object" {
  name   = "${local.function_group}/${data.archive_file.function_source_zip.output_md5}-source.zip"
  bucket = var.gcs_function_source_bucket
  source = data.archive_file.function_source_zip.output_path
}

resource "google_cloudfunctions2_function" "egress_square_gateway" {
  for_each = local.apis

  name     = "${local.function_group}-${var.fundraiser_id}-${each.key}"
  location = var.gcp_region

  build_config {
    runtime     = "go121"
    entry_point = each.value.api
    source {
      storage_source {
        bucket = var.gcs_function_source_bucket
        object = google_storage_bucket_object.function_source_object.name
      }
    }
  }

  service_config {
    available_memory   = "128Mi"
    timeout_seconds    = 60
    min_instance_count = var.min_instance_count

    environment_variables = {
      GCP_PROJECT        = var.gcp_project_id
      EXPIRATION_TIME    = var.expiration_time
      SQUARE_ENVIRONMENT = var.square_environment
      SQUARE_VERSION     = var.square_version

      SQUARE_PAYMENT_RESPONSE_TOPIC_PATH  = var.square_payment_events_response_topic
      SQUARE_ORDER_RESPONSE_TOPIC_PATH    = var.square_order_events_response_topic
      SQUARE_CUSTOMER_RESPONSE_TOPIC_PATH = var.square_customer_events_response_topic
    }

    secret_environment_variables {
      key        = "SQUARE_ACCESS_TOKEN"
      project_id = var.gcp_project_id
      secret     = google_secret_manager_secret.square_access_token.secret_id
      version    = "latest"
    }
  }

  event_trigger {
    trigger_region = var.gcp_region
    event_type     = "google.cloud.pubsub.topic.v1.messagePublished"
    pubsub_topic   = "projects/${var.gcp_project_id}/topics/${each.value.topic}"
    retry_policy   = "RETRY_POLICY_RETRY"
  }
}

# contains Square Access Token used to authenticate Square API calls
# this is manually entered into GCP Secret Manager after creating the object
resource "google_secret_manager_secret" "square_access_token" {
  secret_id = "square_access_token"

  replication {
    auto {}
  }
}

# allows the SA that the function runs as to access the secret value
resource "google_secret_manager_secret_iam_member" "member" {
  project   = google_secret_manager_secret.square_access_token.project
  secret_id = google_secret_manager_secret.square_access_token.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:194415472833-compute@developer.gserviceaccount.com" # TODO: fix this to map to an explicitly declared SA
}
