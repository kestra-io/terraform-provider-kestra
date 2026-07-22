# A tenant-scope policy narrowed to a namespace subtree
resource "kestra_policy" "tenant" {
  scope     = "TENANT"
  policy_id = "deny-shell-commands"

  content = <<EOT
id: deny-shell-commands
displayName: Deny shell commands
description: Disallow the shell Commands plugin on the data teams
enforcement: ACTIVE
target:
  namespaces:
    - company.team
rules:
  - type: io.kestra.plugin.ee.rules.Deny
    on: PLUGIN
    action: BLOCK
    errorMessage: Shell commands are not allowed
    where:
      - field: type
        operator: EQUAL_TO
        value: io.kestra.plugin.scripts.shell.Commands
EOT
}

# A namespace-scope policy requiring an owner label on every flow
resource "kestra_policy" "namespace" {
  scope     = "NAMESPACE"
  policy_id = "require-owner-label"
  namespace = "company.team"

  content = <<EOT
id: require-owner-label
rules:
  - type: io.kestra.plugin.ee.rules.Require
    on: FLOW
    properties:
      - labels.owner
    errorMessage: Flows must carry an owner label
EOT
}

# An instance-scope policy (super-admin only), narrowed to some tenants
resource "kestra_policy" "instance" {
  scope     = "INSTANCE"
  policy_id = "inject-http-timeout"

  content = <<EOT
id: inject-http-timeout
target:
  tenants:
    - production
rules:
  - type: io.kestra.plugin.ee.rules.Add
    on: PLUGIN
    where:
      - field: type
        operator: EQUAL_TO
        value: io.kestra.plugin.core.http.Request
    values:
      options:
        readTimeout: PT30S
EOT
}
