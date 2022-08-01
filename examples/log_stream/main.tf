terraform {
  required_providers {
    auth0 = {
      source = "auth0/auth0"
    }
  }
}

provider "auth0" {}

resource "auth0_log_stream" "example_http" {
  name = "HTTP log stream"
  type = "http"
  filters = [
    {
      type = "category"
      name = "auth.login.fail"
    },
    {
      type = "category"
      name = "auth.signup.fail"
    }
  ]

  sink {
    http_endpoint       = "https://example.com/logs"
    http_content_type   = "application/json"
    http_content_format = "JSONOBJECT"
    http_authorization  = "AKIAXXXXXXXXXXXXXXXX"
    http_custom_headers = [
      {
        header = "foo"
        value  = "bar"
      }
    ]
  }
}

resource "auth0_log_stream" "example_aws" {
  name   = "AWS Eventbridge"
  type   = "eventbridge"
  status = "active"
  sink {
    aws_account_id = "my_account_id"
    aws_region     = "us-east-2"
  }
}
