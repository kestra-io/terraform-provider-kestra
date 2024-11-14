---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "kestra_kv Resource - terraform-provider-kestra"
subcategory: ""
description: |-
  Manages a Kestra Namespace File.
---

# kestra_kv (Resource)

Manages a Kestra Namespace Key Value Store.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `key` (String) The key of the pair.
- `namespace` (String) The namespace of the Key-Value pair.
- `value` (String) The fetched value.

### Optional

- `type` (String) The type of the value. If not provided, we will try to deduce the type based on the value. Useful in case you provide numbers, booleans, dates or json that you want to be stored as string. Accepted values are: STRING, NUMBER, BOOLEAN, DATETIME, DATE, DURATION, JSON.

### Read-Only

- `id` (String) The ID of this resource.
- `tenant_id` (String) The tenant id.