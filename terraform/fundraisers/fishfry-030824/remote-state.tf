terraform {
  backend "gcs" {
    bucket = "kofc7186-bcallaway-test-tfstate"
    prefix = "fishfry-030824"
  }
}
