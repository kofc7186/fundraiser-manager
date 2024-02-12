locals {
  topic_list = toset([for topic in [
    "payment-events",
    "order-events",
    "label-events",
    "customer-events",
    "refund-events",
    "square-payment-webhook",
    "square-payment-request",
    "square-payment-response",
    "square-order-request",
    "square-order-response",
    "square-customer-webhook",
    "square-customer-request",
    "square-customer-response",
    "square-refund-webhook",
    ] :
    format("%s-%s", var.fundraiser_id, topic)
  ])
}

resource "google_project_service" "pubsub_service" {
  for_each = toset([
    "pubsub.googleapis.com",
  ])

  project = var.gcp_project_id
  service = each.key

  disable_on_destroy = false
}

resource "google_pubsub_topic" "topic" {
  for_each = local.topic_list

  name = each.key
}

# we need to give the SA access to create service account tokens
# https://cloud.google.com/pubsub/docs/authenticate-push-subscriptions#configure_for_push_authentication
resource "google_project_iam_member" "viewer" {
  project = var.gcp_project_id
  role    = "roles/iam.serviceAccountTokenCreator"
  member  = "serviceAccount:service-194415472833@gcp-sa-pubsub.iam.gserviceaccount.com" #TODO: get project number as variable
}
