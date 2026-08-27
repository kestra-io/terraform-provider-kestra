terraform import kestra_policy.instance INSTANCE/{{policy_id}}
terraform import kestra_policy.tenant TENANT/{{tenant_id}}/{{policy_id}}
terraform import kestra_policy.namespace NAMESPACE/{{tenant_id}}/{{namespace}}/{{policy_id}}
