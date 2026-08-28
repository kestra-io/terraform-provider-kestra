locals {
  sanitycheck_files = fileset("${path.module}/example-sanity-checks", "*.yaml")
}