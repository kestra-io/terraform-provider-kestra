---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "kestra_user_api_token Resource - terraform-provider-kestra"
subcategory: ""
description: |-
  Manages a Kestra User Api Token.
---

# kestra_user_api_token (Resource)

Manages a Kestra User Api Token.

## Example Usage

```terraform
resource "kestra_user_api_token" "example" {
  user_id = "4by6NvSLcPXFhCj8nwbZOM"

  name        = "test-token"
  description = "Test token"
  max_age     = "PT1H"
  extended    = false
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `description` (String) The API token description.
- `max_age` (String) The time the token remains valid since creation (ISO 8601 duration format).
- `name` (String) The API token display name.
- `user_id` (String) The ID of the user owning the API Token.

### Optional

- `extended` (Boolean) Specify whether the expiry date is automatically moved forward by max age whenever the token is used. Defaults to `false`.

### Read-Only

- `full_token` (String, Sensitive) The full API token.
- `id` (String) The ID of this resource.

## Import

Import is supported using the following syntax:

```shell
terraform import kestra_user_api_token.example {{user_id}}
```