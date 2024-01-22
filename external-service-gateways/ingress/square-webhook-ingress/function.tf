locals {
  function_group = "square-webhook-ingress"
  function_exclude_list = setunion(
    ["go.sum"],
    fileset("${path.module}", "cmd/**"),         # main harness
    fileset("${path.module}", "**_test.go"),     # go test files
    fileset("${path.module}", "**.tf*"),         # terraform files
    fileset("${path.module}", "**terraform*"),   # terraform files
    fileset("${path.module}", "**terraform/**"), # terraform files
    fileset("${path.module}", "*source-*.zip")   # other source zips
  )
  webhook_url = format("https://%s-%s.cloudfunctions.net/%s", var.gcp_region, var.gcp_project_id, local.function_group)
}

resource "google_project_service" "service" {
  for_each = toset([
    "run.googleapis.com",
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

resource "google_cloudfunctions2_function" "webhook" {
  name     = local.function_group
  location = var.gcp_region

  build_config {
    runtime     = "go121"
    entry_point = "WebhookRouter"
    source {
      storage_source {
        bucket = var.gcs_function_source_bucket
        object = google_storage_bucket_object.function_source_object.name
      }
    }
  }

  service_config {
    available_memory = "128Mi"
    timeout_seconds  = 60

    environment_variables = {
      GCP_PROJECT          = "${var.gcp_project_id}"
      PAYMENT_EVENTS_TOPIC = "${var.payment_events_topic}"
      WEBHOOK_URL          = "${local.webhook_url}"
    }

    secret_environment_variables {
      key        = "SQUARE_SIGNATURE_KEY"
      project_id = var.gcp_project_id
      secret     = google_secret_manager_secret.square_signature_key.secret_id
      version    = "latest"
    }
  }
}

# this is a public-facing HTTP endpoint, so we need to allow all users to invoke this
resource "google_cloud_run_service_iam_member" "webhook" {
  location = google_cloudfunctions2_function.webhook.location
  service  = lower(google_cloudfunctions2_function.webhook.name)
  role     = "roles/run.invoker"
  member   = "allUsers"
}

# contains Square signature key to validate that webhook came from Square
resource "google_secret_manager_secret" "square_signature_key" {
  secret_id = "square_signature_key"

  replication {
    auto {}
  }
}

# allows the SA that the function runs as to access the secret value
resource "google_secret_manager_secret_iam_member" "member" {
  project   = google_secret_manager_secret.square_signature_key.project
  secret_id = google_secret_manager_secret.square_signature_key.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:194415472833-compute@developer.gserviceaccount.com" # TODO: fix this to map to an explicitly declared SA
}
