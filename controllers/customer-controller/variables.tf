variable "gcp_project_id" {
  type = string
  validation {
    condition     = length(var.gcp_project_id) > 0
    error_message = "Must specify gcp_project_id variable."
  }
}

variable "gcp_region" {
  description = "The GCP region in which to create the various cloud resources"
  type        = string
}

variable "gcs_function_source_bucket" {
  description = "The name of the GCS bucket where this function's source code should be uploaded"
  type        = string
  validation {
    condition     = length(var.gcs_function_source_bucket) > 0
    error_message = "Must specify gcs_function_source_bucket variable."
  }
}

variable "fundraiser_id" {
  description = "The unique ID for the specific fundraiser; example is 'fishfry-022324'"
  type        = string
}

variable "expiration_time" {
  description = "The expiration value for all Firestore documents created for the fundraiser, as expressed in a UTC timestamp string in RFC 3339 format; example is '2024-02-25T00:00:00Z'"
  type        = string
}

variable "min_instance_count" {
  description = "The limit on the minimum number of function instances that may coexist at a given time"
  type        = number
  default     = 0
}

variable "customer_events_topic" {
  description = "The pubsub topic where internal customer events are published"
  type        = string
}

variable "order_events_topic" {
  description = "The pubsub topic where internal order events are published"
  type        = string
}

variable "payment_events_topic" {
  description = "The pubsub topic where internal payment events are published"
  type        = string
}

variable "square_customer_request_topic" {
  description = "The pubsub topic where Square customer requests are published"
  type        = string
}

variable "square_customer_response_topic" {
  description = "The pubsub topic where Square customer responses are published"
  type        = string
}

variable "square_customer_webhook_topic" {
  description = "The pubsub topic where Square customer webhook events are published"
  type        = string
}
