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

variable "fundraiser_id" {
  description = "The unique ID for the specific fundraiser; example is 'fishfry-022324'"
  type        = string
}

variable "expiration_time" {
  description = "The expiration value for all Firestore documents created for the fundraiser, as expressed in a UTC timestamp string in RFC 3339 format; example is '2024-02-25T00:00:00Z'"
  type        = string
}

variable "firestore_expiration_field_name" {
  description = "The field in all documents within fundraisers/{fid} which contains a timestamp, after which the document will be deleted"
  type        = string
}

variable "min_instance_count" {
  description = "The limit on the minimum number of function instances that may coexist at a given time"
  type        = number
  default     = 0
}

variable "pull_payments_enabled" {
  description = "Whether pull payments should be enabled as a Cloud Scheduler job"
  type        = bool
  default     = false
}

variable "pull_payments_schedule" {
  description = "The frequency that the pull payments job should run, as expressed in unix cron format"
  type        = string
}

variable "pull_payments_begin_time" {
  description = "The start time to search from for payments in Square, as expressed in a UTC timestamp string in RFC 3339 format; example is '2024-02-25T00:00:00Z'"
  type        = string
}

variable "pull_payments_end_time" {
  description = "The end time to search before for payments in Square, as expressed in a UTC timestamp string in RFC 3339 format; example is '2024-02-25T00:00:00Z'"
  type        = string
}
