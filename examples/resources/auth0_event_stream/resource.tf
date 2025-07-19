# Creates a event stream of type event bridge.
resource "auth0_event_stream" "my_event_stream_event_bridge" {
  name = "my-eventbridge"
  subscriptions = [
    "user.created",
    "user.updated"
  ]
  destination {
    type = "eventbridge"
    configuration = jsonencode(
      {
        aws_account_id = "242849305777"
        aws_region     = "us-east-1"
      }
    )
  }
}

# Creates a event stream of type webhook
resource "auth0_event_stream" "my_event_stream_webhook" {
  name = "my-webhook"
  subscriptions = [
    "user.created",
    "user.updated"
  ]
  destination {
    type = "webhook"
    configuration = jsonencode(
      {
        webhook_endpoint = "https://eof28wtn4v4506o.m.pipedream.net"
        webhook_authorization = {
          method = "bearer"
          token  = "123456789"
        }
      }
    )
  }
}
