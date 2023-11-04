resource "kestra_namespace_file" "example" {
  namespace = "io.kestra.mynamespace"
  filename  = "/path/my-file.sh"
  content   = <<EOT
#!/bin/bash
echo "Hello World"
EOT
}

resource "kestra_namespace_file" "withsource" {
  namespace = "io.kestra.mynamespace"
  filename  = "/path/my-file.sh"
  content   = file("./kestra/file.sh")
}
