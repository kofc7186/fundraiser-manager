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
