resource "auth0_rule" "my_rule" {
  name    = "empty-rule"
  script  = <<EOF
    function (user, context, callback) {
      callback(null, user, context);
    }
  EOF
  enabled = true
}
