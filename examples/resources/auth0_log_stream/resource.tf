# This is an example of an http log stream.
resource "auth0_log_stream" "my_webhook" {
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

# This is an example of an Amazon EventBridge log stream.
resource "auth0_log_stream" "example_aws" {
  name   = "AWS Eventbridge"
  type   = "eventbridge"
  status = "active"

  sink {
    aws_account_id = "my_account_id"
    aws_region     = "us-east-2"
  }
}
