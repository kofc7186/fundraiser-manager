* message examples for real-time overrides

* GCS expiration policy for labels
* GCS expiration policy for GCF source code?

* read .env.yaml from combination of insecure values and secure values (stored as GitHub Secrets)
* naming restrictions on fundraiser-id variable (enforced in terraform) as it is used by various GCP objects and firestore paths

* pin terraform and provider versions, add to dependabot
* terraform fmt check
* golangci-lint check

* (manually) create tf state bucket for terraform under correct GCP account

* variables for minimum instance count on functions (during event)

* doc is at https://docs.google.com/document/d/1uF_sTjbruagMH0XyMbdFCKh6TkWndbQ4FfE0yfAs7_U/edit#heading=h.haf2nx8ywfn0

SAs:
* square-webhook-ingress needs pubsub.Publisher, run.Invoker (for allUsers)
* eventarc-triggered functions need iam.serviceAccountTokenCreator

Testing:
* unit tests
* local integration tests
* square sandbox tests
* idempotent tests
* timeout tests (transient vs fatal errors)
* customer ID is missing or object doesn't exist

Instructions:
* rotate Square access token, webhook signature keys

Documentation:
* overall system diagram
