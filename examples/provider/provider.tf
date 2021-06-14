provider "kestra" {
  # mandatory, the url to kestra
  url = "http://localhost:8080"

  # optional basic auth username
  username = "john"

  # optional basic auth password
  password = "my-password"

  # optional jwt token
  jwt = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6Iktlc3RyYS5pbyIsImlhdCI6MTUxNjIzOTAyMn0.hm2VKztDJP7CUsI69Th6Y5NLEQrXx7OErLXay55GD5U"
}
