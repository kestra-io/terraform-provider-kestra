---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "kestra_worker_group Data Source - terraform-provider-kestra"
subcategory: ""
description: |-
  Use this data source to access information about an existing Kestra Worker Group.
---

# kestra_worker_group (Data Source)

Use this data source to access information about an existing Kestra Worker Group.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `key` (String) The worker group key.

### Read-Only

- `allowed_tenants` (String) The list of tenants allowed to use the worker group.
- `description` (String) The worker group description.
- `id` (String) The worker group id.
