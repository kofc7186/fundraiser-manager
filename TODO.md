* message examples for real-time overrides

* generalize boilerplate code (many functions (controllers) look *very similar*)

* double check retry, message ordering, and dead-letter settings on pubsub

* log context labels (customer/refund/payment/order/label ID, order Number)

* passing names of topics down does not introduce terraform dependency on objects, such that topic creation happens before function initialization

* note max of 3 webhook URLs from Square to endpoints (per square 'app')
* terraform configuration of square webhook config

* GCS expiration policy for labels
* GCS expiration policy for GCF source code?

* read .env.yaml from combination of insecure values and secure values (stored as GitHub Secrets)
* naming restrictions on fundraiser-id variable (enforced in terraform) as it is used by various GCP objects and firestore paths

* dead-letter-topic config

* polling Square payments for anything missed

* pin terraform and provider versions, add to dependabot
* group dependabot PRs
* terraform fmt check
* tfsec check
* golangci-lint check

* (manually) create tf state bucket for terraform under correct GCP account

* order's customer_id doesn't seem to get set, but is reliably set on payment object.

- codify field mapping between objects
  * derived from (e.g. order copies from payment, customer; payment copies from refund, label copies from order)

* order-controller watches Payment(Created|Updated) events:
  - if it sees an payment.squareOrderID that it does not have a firestore entry for:
    - it creates a new firestore Order with the known information from the payment in an 'UNKNOWN' state
    - it creates an async Square RetrieveOrder event and exits
    - Square response comes in, populates other fields in order; state may or may not be sufficient at this point
  - if it does know about the order, it computes updates from all derived fields from that object (struct tags?)

* order.created/order.updated webhooks fire, triggering async square response to flow to order-controller:
  - check firestore for matching entry for order.ID
    - if exists, update fields (what logic?)
    - if doesn't exist:
      - create a new firestore Order with known information from Square response
      - if customer_id known AND insufficient quality of Customer-derived information OR if customer_id unknown:
        - emit OrderCreated event

& customer-controller and payment-controller should listen for OrderIncomplete events, and respond with sending the latest known-good CustomerUpdated / PaymentUpdated events

* double check idempotencyKey of "" - shouldn't we set this to something?

CreateOrderFromPayment
Need a consistent function for validing order transition to go from UNKNOWN to (ONLINE | PRESENT)

? payment-controller watches orders; for each order it sees, it computes a list of updates against the latest known payment for that order;
  - if no updates, just quietly exit
  - if order appears to be 'stale', then 

* terraform import artifact-repository for gcf-functions, set expiration policy

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
* unlinked and linked refunds
* out-of-order .created/.updated webhook events

Instructions:
* rotate Square access token, webhook signature keys

Documentation:
* overall system diagram
* higher-order events VS lower-order events
