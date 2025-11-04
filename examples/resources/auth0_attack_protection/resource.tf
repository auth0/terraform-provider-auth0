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

    bot_detection {
        bot_detection_level             = "medium"
        challenge_password_policy       = "when_risky"
        challenge_passwordless_policy   = "when_risky"
        challenge_password_reset_policy = "always"
        allowlist                       = ["192.168.1.0", "10.0.0.0"]
        monitoring_mode_enabled         = true
    }
}

# ============================================================================
# CAPTCHA PROVIDER EXAMPLES - One per Provider
# ============================================================================

# Example 1: reCAPTCHA v2
resource "auth0_attack_protection" "captcha_recaptcha_v2" {
    captcha {
        active_provider_id = "recaptcha_v2"
        recaptcha_v2 {
            site_key = var.recaptcha_v2_site_key
            secret   = var.recaptcha_v2_secret
        }
    }
}

# Example 2: reCAPTCHA Enterprise
resource "auth0_attack_protection" "captcha_recaptcha_enterprise" {
    captcha {
        active_provider_id = "recaptcha_enterprise"
        recaptcha_enterprise {
            site_key   = var.recaptcha_enterprise_site_key
            api_key    = var.recaptcha_enterprise_api_key
            project_id = var.recaptcha_enterprise_project_id
        }
    }
}

# Example 3: hCaptcha
resource "auth0_attack_protection" "captcha_hcaptcha" {
    captcha {
        active_provider_id = "hcaptcha"
        hcaptcha {
            site_key = var.hcaptcha_site_key
            secret   = var.hcaptcha_secret
        }
    }
}

# Example 4: Friendly Captcha
resource "auth0_attack_protection" "captcha_friendly_captcha" {
    captcha {
        active_provider_id = "friendly_captcha"
        friendly_captcha {
            site_key = var.friendly_captcha_site_key
            secret   = var.friendly_captcha_secret
        }
    }
}

# Example 5: Arkose Labs
resource "auth0_attack_protection" "captcha_arkose" {
    captcha {
        active_provider_id = "arkose"
        arkose {
            site_key           = var.arkose_site_key
            secret             = var.arkose_secret
            client_subdomain   = "client.example.com"
            verify_subdomain   = "verify.example.com"
            fail_open          = false
        }
    }
}

# ============================================================================
# VARIABLES FOR SENSITIVE DATA
# ============================================================================

# reCAPTCHA v2
variable "recaptcha_v2_site_key" {
    type        = string
    description = "Google reCAPTCHA v2 site key"
    sensitive   = true
}

variable "recaptcha_v2_secret" {
    type        = string
    description = "Google reCAPTCHA v2 secret key"
    sensitive   = true
}

# reCAPTCHA Enterprise
variable "recaptcha_enterprise_site_key" {
    type        = string
    description = "Google reCAPTCHA Enterprise site key"
    sensitive   = true
}

variable "recaptcha_enterprise_api_key" {
    type        = string
    description = "Google reCAPTCHA Enterprise API key"
    sensitive   = true
}

variable "recaptcha_enterprise_project_id" {
    type        = string
    description = "Google reCAPTCHA Enterprise project ID"
}

# hCaptcha
variable "hcaptcha_site_key" {
    type        = string
    description = "hCaptcha site key"
    sensitive   = true
}

variable "hcaptcha_secret" {
    type        = string
    description = "hCaptcha secret key"
    sensitive   = true
}

# Friendly Captcha
variable "friendly_captcha_site_key" {
    type        = string
    description = "Friendly Captcha site key"
    sensitive   = true
}

variable "friendly_captcha_secret" {
    type        = string
    description = "Friendly Captcha secret key"
    sensitive   = true
}

# Arkose Labs
variable "arkose_site_key" {
    type        = string
    description = "Arkose Labs site key"
    sensitive   = true
}

variable "arkose_secret" {
    type        = string
    description = "Arkose Labs secret key"
    sensitive   = true
}
