resource "google_project_service" "function_service" {
  for_each = toset([
    "run.googleapis.com",
    "eventarc.googleapis.com",
  ])

  project = var.gcp_project_id
  service = each.key

  disable_on_destroy = false
}

# source code for functions must be uploaded to GCS before deployed
resource "google_storage_bucket" "function-source-bucket" {
  name          = format("%s-functions-source", var.fundraiser_id)
  location      = "US"
  force_destroy = true # will delete contents and bucket on 'terraform destroy'

  uniform_bucket_level_access = true

  versioning {
    enabled = true
  }
}

module "square-webhook-ingress" {
  source = "../../../external-service-gateways/ingress/square-webhook-ingress"

  gcp_project_id = var.gcp_project_id
  gcp_region     = var.gcp_region

  gcs_function_source_bucket = google_storage_bucket.function-source-bucket.name

  fundraiser_id      = var.fundraiser_id
  expiration_time    = var.expiration_time
  min_instance_count = var.min_instance_count

  square_order_request_topic    = google_pubsub_topic.topic["${var.fundraiser_id}-square-order-request"].name
  square_customer_webhook_topic = google_pubsub_topic.topic["${var.fundraiser_id}-square-customer-webhook"].name
  square_payment_webhook_topic  = google_pubsub_topic.topic["${var.fundraiser_id}-square-payment-webhook"].name
  square_refund_webhook_topic   = google_pubsub_topic.topic["${var.fundraiser_id}-square-refund-webhook"].name
}

module "egress-square-gateway" {
  source = "../../../external-service-gateways/egress/egress-square-gateway"

  gcp_project_id = var.gcp_project_id
  gcp_region     = var.gcp_region

  gcs_function_source_bucket = google_storage_bucket.function-source-bucket.name

  fundraiser_id      = var.fundraiser_id
  expiration_time    = var.expiration_time
  min_instance_count = var.min_instance_count

  square_environment = "production"

  square_payment_events_request_topic   = google_pubsub_topic.topic["${var.fundraiser_id}-square-payment-request"].name
  square_payment_events_response_topic  = google_pubsub_topic.topic["${var.fundraiser_id}-square-payment-response"].name
  square_order_events_request_topic     = google_pubsub_topic.topic["${var.fundraiser_id}-square-order-request"].name
  square_order_events_response_topic    = google_pubsub_topic.topic["${var.fundraiser_id}-square-order-response"].name
  square_customer_events_request_topic  = google_pubsub_topic.topic["${var.fundraiser_id}-square-customer-request"].name
  square_customer_events_response_topic = google_pubsub_topic.topic["${var.fundraiser_id}-square-customer-response"].name
}

module "event-lake-controller" {
  source = "../../../controllers/event-lake-controller"

  gcp_project_id = var.gcp_project_id
  gcp_region     = var.gcp_region

  gcs_function_source_bucket = google_storage_bucket.function-source-bucket.name

  fundraiser_id      = var.fundraiser_id
  expiration_time    = var.expiration_time
  min_instance_count = var.min_instance_count

  topics_to_monitor = local.topic_list
}

module "payment-controller" {
  source = "../../../controllers/payment-controller"

  gcp_project_id = var.gcp_project_id
  gcp_region     = var.gcp_region

  gcs_function_source_bucket = google_storage_bucket.function-source-bucket.name

  fundraiser_id      = var.fundraiser_id
  expiration_time    = var.expiration_time
  min_instance_count = var.min_instance_count

  square_payment_webhook_topic  = google_pubsub_topic.topic["${var.fundraiser_id}-square-payment-webhook"].name
  payment_events_topic          = google_pubsub_topic.topic["${var.fundraiser_id}-payment-events"].name
  square_payment_request_topic  = google_pubsub_topic.topic["${var.fundraiser_id}-square-payment-request"].name
  square_payment_response_topic = google_pubsub_topic.topic["${var.fundraiser_id}-square-payment-response"].name

  pull_payments_enabled    = var.pull_payments_enabled
  pull_payments_schedule   = var.pull_payments_schedule
  pull_payments_begin_time = var.pull_payments_begin_time
  pull_payments_end_time   = var.pull_payments_end_time
}

module "refund-controller" {
  source = "../../../controllers/refund-controller"

  gcp_project_id = var.gcp_project_id
  gcp_region     = var.gcp_region

  gcs_function_source_bucket = google_storage_bucket.function-source-bucket.name

  fundraiser_id      = var.fundraiser_id
  expiration_time    = var.expiration_time
  min_instance_count = var.min_instance_count

  square_refund_webhook_topic = google_pubsub_topic.topic["${var.fundraiser_id}-square-refund-webhook"].name
  refund_events_topic         = google_pubsub_topic.topic["${var.fundraiser_id}-refund-events"].name
}

module "customer-controller" {
  source = "../../../controllers/customer-controller"

  gcp_project_id = var.gcp_project_id
  gcp_region     = var.gcp_region

  gcs_function_source_bucket = google_storage_bucket.function-source-bucket.name

  fundraiser_id      = var.fundraiser_id
  expiration_time    = var.expiration_time
  min_instance_count = var.min_instance_count

  square_customer_webhook_topic  = google_pubsub_topic.topic["${var.fundraiser_id}-square-customer-webhook"].name
  customer_events_topic          = google_pubsub_topic.topic["${var.fundraiser_id}-customer-events"].name
  order_events_topic             = google_pubsub_topic.topic["${var.fundraiser_id}-order-events"].name
  payment_events_topic           = google_pubsub_topic.topic["${var.fundraiser_id}-payment-events"].name
  square_customer_request_topic  = google_pubsub_topic.topic["${var.fundraiser_id}-square-customer-request"].name
  square_customer_response_topic = google_pubsub_topic.topic["${var.fundraiser_id}-square-customer-response"].name
}

module "order-controller" {
  source = "../../../controllers/order-controller"

  gcp_project_id = var.gcp_project_id
  gcp_region     = var.gcp_region

  gcs_function_source_bucket = google_storage_bucket.function-source-bucket.name

  fundraiser_id      = var.fundraiser_id
  expiration_time    = var.expiration_time
  min_instance_count = var.min_instance_count

  customer_events_topic       = google_pubsub_topic.topic["${var.fundraiser_id}-customer-events"].name
  order_events_topic          = google_pubsub_topic.topic["${var.fundraiser_id}-order-events"].name
  payment_events_topic        = google_pubsub_topic.topic["${var.fundraiser_id}-payment-events"].name
  square_order_request_topic  = google_pubsub_topic.topic["${var.fundraiser_id}-square-order-request"].name
  square_order_response_topic = google_pubsub_topic.topic["${var.fundraiser_id}-square-order-response"].name
}
