data "kestra_namespace_file" "example" {
  namespace = "company.team"
  filename  = "myscript.py"
  content   = file("myscript.py")
}
