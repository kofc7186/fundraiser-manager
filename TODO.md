* message examples for real-time overrides

* generalize boilerplate code (many functions (controllers) look *very similar*)

* double check retry, message ordering, and dead-letter settings on pubsub

* log context labels (customer/refund/payment/order/label ID, order Number)

* passing names of topics down does not introduce terraform dependency on objects, such that topic creation happens before function initialization

* square-webhook-ingress URL does not contain fundraiser ID in it, so if we ever wanted to run more than one fundraiser at a time this might be a problem
* note max of 3 webhook URLs from Square to endpoints (per square 'app')
* terraform configuration of square webhook config

* GCS expiration policy for labels
* GCS expiration policy for GCF source code?

* read .env.yaml from combination of insecure values and secure values (stored as GitHub Secrets)
* naming restrictions on fundraiser-id variable (enforced in terraform) as it is used by various GCP objects and firestore paths

* pin terraform and provider versions, add to dependabot
* group dependabot PRs
* terraform fmt check
* tfsec check
* golangci-lint check

* (manually) create tf state bucket for terraform under correct GCP account

* variables for minimum instance count on functions (during event)

* ORDER WEBHOOKS ARE DIFFERENT
* they don't pass the order object

* doc is at https://docs.google.com/document/d/1uF_sTjbruagMH0XyMbdFCKh6TkWndbQ4FfE0yfAs7_U/edit#heading=h.haf2nx8ywfn0

SAs:
* square-webhook-ingress needs pubsub.Publisher, run.Invoker (for allUsers)
* eventarc-triggered functions need iam.serviceAccountTokenCreator

Testing:
* unit tests
* local integration tests
* square sandbox tests
* idempotency tests
* timeout tests (transient vs fatal errors)
* customer ID is missing or object doesn't exist
* does a square update over-write internally-computed fields incorrectly?

Instructions:
* rotate Square access token, webhook signature keys

Documentation:
* overall system diagram
