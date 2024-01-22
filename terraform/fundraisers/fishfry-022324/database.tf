locals {
  collection_groups = toset([
    #"events",
    #"payments", 
    #"orders", 
    #"labels", 
  ])
}

# Set TTL policy on all collection groups
# note: takes ~4 min to create/delete
resource "google_firestore_field" "expiration_ttl_policy" {
  for_each = local.collection_groups

  project    = var.gcp_project_id
  database   = "(default)"
  collection = each.key
  field      = var.firestore_expiration_field_name

  ttl_config {}
}
