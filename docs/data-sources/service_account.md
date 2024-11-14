---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "kestra_service_account Data Source - terraform-provider-kestra"
subcategory: ""
description: |-
  Use this data source to access information about an existing Kestra Service Account.
---

# kestra_service_account (Data Source)

Use this data source to access information about an existing Kestra Service Account.

## Example Usage

```terraform
data "kestra_user_service_account" "example" {
  id = "68xAawPfiJPkTkZJIPX6jQ"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `id` (String) The service account id.

### Optional

- `group` (Block Set) The service account group. (see [below for nested schema](#nestedblock--group))

### Read-Only

- `description` (String) The service account description.
- `username` (String) The service account name.

<a id="nestedblock--group"></a>
### Nested Schema for `group`

Optional:

- `tenant_id` (String) The tenant id for this group.

Read-Only:

- `group_id` (String) The group id.