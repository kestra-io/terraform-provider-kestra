resource "kestra_namespace_file" "example" {
  namespace        = "io.kestra.mynamespace"
  destination_path = "/path/my-file.sh"
  content          = <<EOT
#!/bin/bash
echo "Hello World"
EOT
}

resource "kestra_namespace_file" "withsource" {
  namespace        = "io.kestra.mynamespace"
  destination_path = "/path/my-file.sh"
  local_path       = "./kestra/file.sh"
}