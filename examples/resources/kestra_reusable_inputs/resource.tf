# A reusable inputs block, referenced from flows via a REUSABLE_INPUTS input
resource "kestra_reusable_inputs" "environment" {
  reusable_inputs_id = "environment-selector"
  namespace          = "company.team"

  content = <<EOT
description: Shared environment selector for the data team's flows
inputs:
  - id: environment
    type: SELECT
    values:
      - dev
      - staging
      - prod
    defaults: dev
EOT
}
