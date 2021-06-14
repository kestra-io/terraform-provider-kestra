resource "kestra_group" "example" {
  name        = "Friendly name"
  description = "Friendly description"

  global_roles = [
    "4by6NvSLcPXFhCj8nwbZOM",
    "UetX7LZLQBFlNHGHbhElO",
  ]

  namespace_roles {
    namespace = "io.kestra.n1"
    roles     = "UetX7LZLQBFlNHGHbhElO"
  }

  namespace_roles {
    namespace = "io.kestra.n2"
    roles     = "UetX7LZLQBFlNHGHbhElO"
  }
}
