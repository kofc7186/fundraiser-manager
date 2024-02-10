resource "google_project_service" "service" {
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

  expiration_time = var.expiration_time

  square_order_request_topic = google_pubsub_topic.topic["${var.fundraiser_id}-square-order-request"].name
  customer_events_topic      = google_pubsub_topic.topic["${var.fundraiser_id}-customer-events"].name
  payment_events_topic       = google_pubsub_topic.topic["${var.fundraiser_id}-payment-events"].name
  refund_events_topic        = google_pubsub_topic.topic["${var.fundraiser_id}-refund-events"].name
}

module "egress-square-gateway" {
  source = "../../../external-service-gateways/egress/egress-square-gateway"

  gcp_project_id = var.gcp_project_id
  gcp_region     = var.gcp_region

  gcs_function_source_bucket = google_storage_bucket.function-source-bucket.name

  fundraiser_id   = var.fundraiser_id
  expiration_time = var.expiration_time

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

  fundraiser_id   = var.fundraiser_id
  expiration_time = var.expiration_time

  topics_to_monitor = local.topic_list
}

module "payment-controller" {
  source = "../../../controllers/payment-controller"

  gcp_project_id = var.gcp_project_id
  gcp_region     = var.gcp_region

  gcs_function_source_bucket = google_storage_bucket.function-source-bucket.name

  fundraiser_id   = var.fundraiser_id
  expiration_time = var.expiration_time

  payment_events_topic = google_pubsub_topic.topic["${var.fundraiser_id}-payment-events"].name
}

module "refund-controller" {
  source = "../../../controllers/refund-controller"

  gcp_project_id = var.gcp_project_id
  gcp_region     = var.gcp_region

  gcs_function_source_bucket = google_storage_bucket.function-source-bucket.name

  fundraiser_id   = var.fundraiser_id
  expiration_time = var.expiration_time

  refund_events_topic = google_pubsub_topic.topic["${var.fundraiser_id}-refund-events"].name
}

module "customer-controller" {
  source = "../../../controllers/customer-controller"

  gcp_project_id = var.gcp_project_id
  gcp_region     = var.gcp_region

  gcs_function_source_bucket = google_storage_bucket.function-source-bucket.name

  fundraiser_id   = var.fundraiser_id
  expiration_time = var.expiration_time

  customer_events_topic = google_pubsub_topic.topic["${var.fundraiser_id}-customer-events"].name
  square_customer_request_topic = google_pubsub_topic.topic["${var.fundraiser_id}-square-customer-request"].name
  square_customer_response_topic = google_pubsub_topic.topic["${var.fundraiser_id}-square-customer-response"].name
}

module "order-controller" {
  source = "../../../controllers/order-controller"

  gcp_project_id = var.gcp_project_id
  gcp_region     = var.gcp_region

  gcs_function_source_bucket = google_storage_bucket.function-source-bucket.name

  fundraiser_id   = var.fundraiser_id
  expiration_time = var.expiration_time

  order_events_topic = google_pubsub_topic.topic["${var.fundraiser_id}-order-events"].name
  square_order_response_topic = google_pubsub_topic.topic["${var.fundraiser_id}-square-order-response"].name
}
