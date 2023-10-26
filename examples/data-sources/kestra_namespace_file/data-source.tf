data "kestra_namespace_file" "example" {
  namespace_       = "io.kestra.mynamespace"
  destination_path = "myscript.py"
}
