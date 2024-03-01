terraform {
  required_providers {
    startrail = {
      source  = "srevin/startrail"
      version = "~> 1.0"
    }
  }
}

provider "startrail" {
  endpoint = "http://localhost:8000"
  debug    = true
}

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


