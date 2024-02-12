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

variable "square_environment" {
  description = "The Square environment to use; must be either 'production' or 'sandbox'"
  type        = string
  validation {
    condition     = var.square_environment == "production" || var.square_environment == "sandbox"
    error_message = "Square environment must be either 'production' or 'sandbox'"
  }
}

variable "square_version" {
  description = "The version of the Square API to use; if unset, will use the latest version published"
  type        = string
  default     = null
}

variable "square_payment_events_request_topic" {
  description = "The pubsub topic where async Square payment requests are published"
  type        = string
}

variable "square_payment_events_response_topic" {
  description = "The pubsub topic where async Square payment responses are published"
  type        = string
}

variable "square_order_events_request_topic" {
  description = "The pubsub topic where async Square order requests are published"
  type        = string
}

variable "square_order_events_response_topic" {
  description = "The pubsub topic where async Square order responses are published"
  type        = string
}

variable "square_customer_events_request_topic" {
  description = "The pubsub topic where async Square customer requests are published"
  type        = string
}

variable "square_customer_events_response_topic" {
  description = "The pubsub topic where async Square customer responses are published"
  type        = string
}
