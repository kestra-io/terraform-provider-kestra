data "kestra_namespace_file" "example" {
  namespace = "io.kestra.mynamespace"
  filename  = "myscript.py"
  content   = file("myscript.py")
}
