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

variable "topics_to_monitor" {
  description = "The list of topic IDs that the event-lake-controller should subscribe to, in order to record events"
  type        = set(string)
}
