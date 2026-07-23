data "kestra_policy" "tenant" {
  scope     = "TENANT"
  policy_id = "deny-shell-commands"
}

data "kestra_policy" "namespace" {
  scope     = "NAMESPACE"
  namespace = "company.team"
  policy_id = "require-owner-label"
}
