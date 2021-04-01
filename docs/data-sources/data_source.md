---
page_title: "kestra_data_source Data Source - terraform-provider-kestra"
subcategory: ""
description: |-
  Sample data source in the Terraform provider kestra.
---

# Data Source `kestra_data_source`

Sample data source in the Terraform provider kestra.

## Example Usage

```terraform
data "kestra_data_source" "example" {
  namespace = "foo"
}
```

## Schema

### Required

- **namespace** (String, Required) Sample attribute.

### Optional

- **id** (String, Optional) The ID of this resource.


