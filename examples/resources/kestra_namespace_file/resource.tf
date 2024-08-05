resource "kestra_namespace_file" "example" {
  namespace = "company.team"
  filename  = "/path/my-file.sh"
  content   = <<EOT
#!/bin/bash
echo "Hello World"
EOT
}

resource "kestra_namespace_file" "withsource" {
  namespace = "company.team"
  filename  = "/path/my-file.sh"
  content   = file("./kestra/file.sh")
}
