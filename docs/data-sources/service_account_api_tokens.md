---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "kestra_service_account_api_tokens Data Source - terraform-provider-kestra"
subcategory: ""
description: |-
  Use this data source to access information about the API tokens of a Kestra Service Account.
  -> This resource is only available on the Enterprise Edition https://kestra.io/enterprise
---

# kestra_service_account_api_tokens (Data Source)

Use this data source to access information about the API tokens of a Kestra Service Account.

-> This resource is only available on the [Enterprise Edition](https://kestra.io/enterprise)



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `service_account_id` (String) The ID of the Service Account owning the API Tokens.

### Read-Only

- `api_tokens` (Set of Object) The API tokens of the Service Account. (see [below for nested schema](#nestedatt--api_tokens))
- `id` (String) The ID of this resource.

<a id="nestedatt--api_tokens"></a>
### Nested Schema for `api_tokens`

Read-Only:

- `description` (String)
- `exp` (String)
- `expired` (Boolean)
- `extended` (Boolean)
- `iat` (String)
- `last_used` (String)
- `name` (String)
- `token_id` (String)
- `token_prefix` (String)
