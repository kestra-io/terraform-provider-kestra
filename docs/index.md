---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "kestra Provider"
subcategory: ""
description: |-
  
---

# kestra Provider

## Kestra 0.23.x compatibility

!> **Warning:** Kestra Terraform provider 0.23.x is only compatible with Kestra 0.23.x and above.

Additionally, if you want to terraform Kestra 0.23.x you need to use Kestra Terraform provider 0.23.x

### Breaking changes from 0.23.x

`kestra_service_account` resource field `username` was renamed to `name` (see [service_account.md](resources/service_account.md)) 


## Example Usage

```terraform
provider "kestra" {
  # mandatory, the Kestra webserver/standalone URL
  url = "http://localhost:8080"

  # optional basic auth username
  username = "john"

  # optional basic auth password
  password = "my-password"

  # optional jwt token (EE)
  jwt = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6Iktlc3RyYS5pbyIsImlhdCI6MTUxNjIzOTAyMn0.hm2VKztDJP7CUsI69Th6Y5NLEQrXx7OErLXay55GD5U"

  # optional tenant id (EE)
  tenant_id = "the-tenant-id"

  # optional extra headers
  extra_headers = {
    x-pipeline    = "*****"
    authorization = "Bearer *****"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `api_token` (String, Sensitive) The API token (EE)
- `extra_headers` (Map of String) Extra headers to add to every request
- `jwt` (String, Sensitive) The JWT token (EE)
- `keep_original_source` (Boolean) Keep original source code, keeping comment and indentation. Setting to false is now deprecated and will be removed in the future.
- `password` (String, Sensitive) The BasicAuth password
- `tenant_id` (String) The tenant id (EE)
- `timeout` (Number) The timeout (in seconds) for http requests
- `url` (String) The endpoint url
- `username` (String) The BasicAuth username
