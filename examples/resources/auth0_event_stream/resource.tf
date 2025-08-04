# Creates an event stream of type eventbridge
resource "auth0_event_stream" "my_event_stream_event_bridge" {
  name             = "my-eventbridge"
  destination_type = "eventbridge"
  subscriptions = [
    "user.created",
    "user.updated"
  ]

  eventbridge_configuration {
    aws_account_id = "242849305777"
    aws_region     = "us-east-1"
  }
}

# Creates an event stream of type webhook
resource "auth0_event_stream" "my_event_stream_webhook" {
  name             = "my-webhook"
  destination_type = "webhook"
  subscriptions = [
    "user.created",
    "user.updated"
  ]

  webhook_configuration {
    webhook_endpoint = "https://eof28wtn4v4506o.m.pipedream.net"

    webhook_authorization {
      method = "bearer"
      token  = "123456789"
    }
  }
}
