provider "google" {
  project = var.gcp_project_id
  region  = var.gcp_region
}

# data.external.project_workspace.result.path will be the top level directory of the git workspace
# data "external" "project_workspace" {
#   program = ["jq", "-n", "--arg", "path", "$(git rev-parse --show-toplevel)", "'{path:$path}'"]
# }
