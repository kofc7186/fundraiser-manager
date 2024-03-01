fundraiser_id = "fishfry-030824"

gcp_project_id = "serverless-fish-fry"
gcp_region     = "us-east1"

firestore_expiration_field_name = "expiration"

# this is set for EST (UTC-5)
expiration_time = "2024-03-10T05:00:00Z"

# this can be increased during an event to provide lower latency
min_instance_count = 0

pull_payments_enabled    = false
pull_payments_schedule   = "*/2 * * * *"
pull_payments_begin_time = "2024-02-25T00:00:00Z"
pull_payments_end_time   = "2024-03-09T01:00:00Z"
