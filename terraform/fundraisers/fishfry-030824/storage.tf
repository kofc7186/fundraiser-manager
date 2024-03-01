resource "google_storage_bucket" "label_bucket" {
  name          = format("%s-labels", var.fundraiser_id)
  location      = "US"
  force_destroy = true # will delete contents and bucket on 'terraform destroy'

  versioning {
    enabled = true
  }

  # this will cause all objects in the bucket to be deleted within 30 days
  lifecycle_rule {
    condition {
      age = 30
    }
    action {
      type = "Delete"
    }
  }
}

# allows anyone in world with URL to read contents
resource "google_storage_bucket_access_control" "public_rule" {
  bucket = google_storage_bucket.label_bucket.name
  role   = "READER"
  entity = "allUsers"
}
