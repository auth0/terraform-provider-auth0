provider "auth0" {}

resource "auth0_attack_protection" "attack_protection" {
  suspicious_ip_throttling {
    enabled   = true
    shields   = ["block", "admin_notification"]
    allowlist = ["127.0.0.1"]
    pre_user_registration {
      max_attempts = 1
      rate         = 3600
    }
    pre_login {
      max_attempts = 1
      rate         = 3600
    }
  }
}
