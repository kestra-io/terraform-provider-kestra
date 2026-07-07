data "kestra_user" "example" {
  user_id = "68xAawPfiJPkTkZJIPX6jQ"
}

data "kestra_user" "by_email" {
  email = "john@example.com"
}
