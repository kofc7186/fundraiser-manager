locals {
  function_group = "customer-controller"
  function_exclude_list = setunion(
    ["go.sum"],
    fileset("${path.module}", "cmd/**"),         # main harness
    fileset("${path.module}", "**_test.go"),     # go test files
    fileset("${path.module}", "**.tf*"),         # terraform files
    fileset("${path.module}", "**terraform*"),   # terraform files
    fileset("${path.module}", "**terraform/**"), # terraform files
    fileset("${path.module}", "*source-*.zip")   # other source zips
  )
}

resource "google_project_service" "service" {
  for_each = toset([
    "run.googleapis.com",
    "eventarc.googleapis.com",
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
  name   = "${local.function_group}/${var.fundraiser_id}-${data.archive_file.function_source_zip.output_md5}-source.zip"
  bucket = var.gcs_function_source_bucket
  source = data.archive_file.function_source_zip.output_path
}

resource "google_cloudfunctions2_function" "customer-controller-square-webhook" {
  name     = "${local.function_group}-${var.fundraiser_id}-square-webhook"
  location = var.gcp_region

  build_config {
    runtime     = "go121"
    entry_point = "ProcessSquareCustomerWebhookEvent"
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
      GCP_PROJECT                   = var.gcp_project_id
      EXPIRATION_TIME               = var.expiration_time
      FUNDRAISER_ID                 = var.fundraiser_id
      CUSTOMER_EVENTS_TOPIC         = var.customer_events_topic
      SQUARE_CUSTOMER_WEBHOOK_TOPIC = var.square_customer_webhook_topic
      SQUARE_CUSTOMER_REQUEST_TOPIC = var.square_customer_request_topic
    }
  }

  event_trigger {
    trigger_region = var.gcp_region
    event_type     = "google.cloud.pubsub.topic.v1.messagePublished"
    pubsub_topic   = "projects/${var.gcp_project_id}/topics/${var.square_customer_webhook_topic}"
    retry_policy   = "RETRY_POLICY_RETRY"
  }
}

resource "google_cloudfunctions2_function" "customer-controller-cdc" {
  name     = "${local.function_group}-${var.fundraiser_id}-cdc"
  location = var.gcp_region

  build_config {
    runtime     = "go121"
    entry_point = "ProcessCDCEvent"
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
      GCP_PROJECT                   = var.gcp_project_id
      EXPIRATION_TIME               = var.expiration_time
      FUNDRAISER_ID                 = var.fundraiser_id
      CUSTOMER_EVENTS_TOPIC         = var.customer_events_topic
      SQUARE_CUSTOMER_WEBHOOK_TOPIC = var.square_customer_webhook_topic
      SQUARE_CUSTOMER_REQUEST_TOPIC = var.square_customer_request_topic
    }
  }

  event_trigger {
    trigger_region = var.gcp_region
    event_type     = "google.cloud.firestore.document.v1.written"
    event_filters {
      attribute = "database"
      value     = "(default)"
    }
    event_filters {
      operator  = "match-path-pattern"
      attribute = "document"
      value     = "fundraisers/${var.fundraiser_id}/customers/{customer}"
    }
    retry_policy   = "RETRY_POLICY_RETRY"
  }
}

resource "google_cloudfunctions2_function" "customer-controller-order-watcher" {
  name     = "${local.function_group}-${var.fundraiser_id}-order-watcher"
  location = var.gcp_region

  build_config {
    runtime     = "go121"
    entry_point = "OrderWatcher"
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
      GCP_PROJECT                   = var.gcp_project_id
      EXPIRATION_TIME               = var.expiration_time
      FUNDRAISER_ID                 = var.fundraiser_id
      CUSTOMER_EVENTS_TOPIC         = var.customer_events_topic
      SQUARE_CUSTOMER_WEBHOOK_TOPIC = var.square_customer_webhook_topic
      SQUARE_CUSTOMER_REQUEST_TOPIC = var.square_customer_request_topic
    }
  }

  event_trigger {
    trigger_region = var.gcp_region
    event_type     = "google.cloud.pubsub.topic.v1.messagePublished"
    pubsub_topic   = "projects/${var.gcp_project_id}/topics/${var.order_events_topic}"
    retry_policy   = "RETRY_POLICY_RETRY"
  }
}

resource "google_cloudfunctions2_function" "customer-controller-payment-watcher" {
  name     = "${local.function_group}-${var.fundraiser_id}-payment-watcher"
  location = var.gcp_region

  build_config {
    runtime     = "go121"
    entry_point = "PaymentWatcher"
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
      GCP_PROJECT                   = var.gcp_project_id
      EXPIRATION_TIME               = var.expiration_time
      FUNDRAISER_ID                 = var.fundraiser_id
      CUSTOMER_EVENTS_TOPIC         = var.customer_events_topic
      SQUARE_CUSTOMER_WEBHOOK_TOPIC = var.square_customer_webhook_topic
      SQUARE_CUSTOMER_REQUEST_TOPIC = var.square_customer_request_topic
    }
  }

  event_trigger {
    trigger_region = var.gcp_region
    event_type     = "google.cloud.pubsub.topic.v1.messagePublished"
    pubsub_topic   = "projects/${var.gcp_project_id}/topics/${var.payment_events_topic}"
    retry_policy   = "RETRY_POLICY_RETRY"
  }
}

resource "google_cloudfunctions2_function" "customer-controller-square-customer-response" {
  name     = "${local.function_group}-${var.fundraiser_id}-square-customer-response"
  location = var.gcp_region

  build_config {
    runtime     = "go121"
    entry_point = "ProcessSquareCustomerResponse"
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
      GCP_PROJECT                   = var.gcp_project_id
      EXPIRATION_TIME               = var.expiration_time
      FUNDRAISER_ID                 = var.fundraiser_id
      CUSTOMER_EVENTS_TOPIC         = var.customer_events_topic
      SQUARE_CUSTOMER_WEBHOOK_TOPIC = var.square_customer_webhook_topic
      SQUARE_CUSTOMER_REQUEST_TOPIC = var.square_customer_request_topic
    }
  }

  event_trigger {
    trigger_region = var.gcp_region
    event_type     = "google.cloud.pubsub.topic.v1.messagePublished"
    pubsub_topic   = "projects/${var.gcp_project_id}/topics/${var.square_customer_response_topic}"
    retry_policy   = "RETRY_POLICY_RETRY"
  }
}
