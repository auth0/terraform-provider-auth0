resource "auth0_attack_protection" "my_protection" {
  suspicious_ip_throttling {
    enabled   = true
    shields   = ["admin_notification", "block"]
    allowlist = ["192.168.1.1"]

    pre_login {
      max_attempts = 100
      rate         = 864000
    }

    pre_user_registration {
      max_attempts = 50
      rate         = 1200
    }
  }

  brute_force_protection {
    allowlist    = ["127.0.0.1"]
    enabled      = true
    max_attempts = 5
    mode         = "count_per_identifier_and_ip"
    shields      = ["block", "user_notification"]
  }

  breached_password_detection {
    admin_notification_frequency = ["daily"]
    enabled                      = true
    method                       = "standard"
    shields                      = ["admin_notification", "block"]

    pre_user_registration {
      shields = ["block"]
    }

    pre_change_password {
      shields = ["block", "admin_notification"]
    }
  }
}
