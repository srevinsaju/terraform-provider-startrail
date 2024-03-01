resource "startrail_service" "hello_world" {
  name        = "hello-world"
  description = "A service which tells hello world"
  environment = "development"

  access {
    auth     = true
    endpoint = "https://example.com/hello"
    internal = false
  }
}
