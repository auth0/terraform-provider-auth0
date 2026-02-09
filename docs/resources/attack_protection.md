---
page_title: "Resource: auth0_attack_protection"
description: |-
  Auth0 can detect attacks and stop malicious attempts to access your application such as blocking traffic from certain IPs and displaying CAPTCHAs.
---

# Resource: auth0_attack_protection

Auth0 can detect attacks and stop malicious attempts to access your application such as blocking traffic from certain IPs and displaying CAPTCHAs.

## Example Usage

```terraform
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
      shields = ["admin_notification", "block"]
    }

    pre_change_password {
      shields = ["admin_notification", "block"]
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
      site_key         = var.arkose_site_key
      secret           = var.arkose_secret
      client_subdomain = "client.example.com"
      verify_subdomain = "verify.example.com"
      fail_open        = false
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
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `bot_detection` (Block List, Max: 1) Bot detection configuration to identify and prevent automated threats. (see [below for nested schema](#nestedblock--bot_detection))
- `breached_password_detection` (Block List, Max: 1) Breached password detection protects your applications from bad actors logging in with stolen credentials. (see [below for nested schema](#nestedblock--breached_password_detection))
- `brute_force_protection` (Block List, Max: 1) Brute-force protection safeguards against a single IP address attacking a single user account. (see [below for nested schema](#nestedblock--brute_force_protection))
- `captcha` (Block List, Max: 1) CAPTCHA configuration for attack protection. (see [below for nested schema](#nestedblock--captcha))
- `suspicious_ip_throttling` (Block List, Max: 1) Suspicious IP throttling blocks traffic from any IP address that rapidly attempts too many logins or signups. (see [below for nested schema](#nestedblock--suspicious_ip_throttling))

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--bot_detection"></a>
### Nested Schema for `bot_detection`

Optional:

- `allowlist` (Set of String) List of IP addresses or ranges that will not trigger bot detection.
- `bot_detection_level` (String) Bot detection level. Possible values: `low`, `medium`, `high`. Set to empty string to disable.
- `challenge_password_policy` (String) Challenge policy for password flow. Possible values: `never`, `when_risky`, `always`.
- `challenge_password_reset_policy` (String) Challenge policy for password reset flow. Possible values: `never`, `when_risky`, `always`.
- `challenge_passwordless_policy` (String) Challenge policy for passwordless flow. Possible values: `never`, `when_risky`, `always`.
- `monitoring_mode_enabled` (Boolean) Whether monitoring mode is enabled for bot detection.


<a id="nestedblock--breached_password_detection"></a>
### Nested Schema for `breached_password_detection`

Required:

- `enabled` (Boolean) Whether breached password detection is active.

Optional:

- `admin_notification_frequency` (Set of String) When `admin_notification` is enabled within the `shields` property, determines how often email notifications are sent. Possible values: `immediately`, `daily`, `weekly`, `monthly`.
- `method` (String) The subscription level for breached password detection methods. Use "enhanced" to enable Credential Guard. Possible values: `standard`, `enhanced`.
- `pre_change_password` (Block List, Max: 1) Configuration options that apply before every password change attempt. (see [below for nested schema](#nestedblock--breached_password_detection--pre_change_password))
- `pre_user_registration` (Block List, Max: 1) Configuration options that apply before every user registration attempt. Only available on public tenants. (see [below for nested schema](#nestedblock--breached_password_detection--pre_user_registration))
- `shields` (Set of String) Action to take when a breached password is detected. Options include: `block` (block compromised user accounts), `user_notification` (send an email to user when we detect that they are using compromised credentials) and `admin_notification` (send an email with a summary of the number of accounts logging in with compromised credentials).

<a id="nestedblock--breached_password_detection--pre_change_password"></a>
### Nested Schema for `breached_password_detection.pre_change_password`

Optional:

- `shields` (Set of String) Action to take when a breached password is detected before the password is changed. Possible values: `block` (block compromised credentials for new accounts), `admin_notification` (send an email notification with a summary of compromised credentials in new accounts).


<a id="nestedblock--breached_password_detection--pre_user_registration"></a>
### Nested Schema for `breached_password_detection.pre_user_registration`

Optional:

- `shields` (Set of String) Action to take when a breached password is detected during a signup. Possible values: `block` (block compromised credentials for new accounts), `admin_notification` (send an email notification with a summary of compromised credentials in new accounts).



<a id="nestedblock--brute_force_protection"></a>
### Nested Schema for `brute_force_protection`

Required:

- `enabled` (Boolean) Whether brute force attack protections are active.

Optional:

- `allowlist` (Set of String) List of trusted IP addresses that will not have attack protection enforced against them. This field allows you to specify multiple IP addresses, or ranges. You can use IPv4 or IPv6 addresses and CIDR notation.
- `max_attempts` (Number) Maximum number of consecutive failed login attempts from a single user before blocking is triggered. Only available on public tenants.
- `mode` (String) Determines whether the IP address is used when counting failed attempts. Possible values: `count_per_identifier_and_ip` (lockout an account from a given IP Address) or `count_per_identifier` (lockout an account regardless of IP Address).
- `shields` (Set of String) Action to take when a brute force protection threshold is violated. Possible values: `block` (block login attempts for a flagged user account), `user_notification` (send an email to user when their account has been blocked).


<a id="nestedblock--captcha"></a>
### Nested Schema for `captcha`

Optional:

- `active_provider_id` (String) Active CAPTCHA provider ID. Set to empty string to disable CAPTCHA. Possible values: `recaptcha_v2`, `recaptcha_enterprise`, `hcaptcha`, `friendly_captcha`, `arkose`, `auth_challenge`, `simple_captcha`.
- `arkose` (Block List, Max: 1) Configuration for Arkose Labs. (see [below for nested schema](#nestedblock--captcha--arkose))
- `auth_challenge` (Block List, Max: 1) Configuration for Auth0's Auth Challenge. (see [below for nested schema](#nestedblock--captcha--auth_challenge))
- `friendly_captcha` (Block List, Max: 1) Configuration for Friendly Captcha. (see [below for nested schema](#nestedblock--captcha--friendly_captcha))
- `hcaptcha` (Block List, Max: 1) Configuration for hCaptcha. (see [below for nested schema](#nestedblock--captcha--hcaptcha))
- `recaptcha_enterprise` (Block List, Max: 1) Configuration for Google reCAPTCHA Enterprise. (see [below for nested schema](#nestedblock--captcha--recaptcha_enterprise))
- `recaptcha_v2` (Block List, Max: 1) Configuration for Google reCAPTCHA v2. (see [below for nested schema](#nestedblock--captcha--recaptcha_v2))

<a id="nestedblock--captcha--arkose"></a>
### Nested Schema for `captcha.arkose`

Required:

- `site_key` (String) Site key for Arkose Labs.

Optional:

- `client_subdomain` (String) Client subdomain for Arkose Labs.
- `fail_open` (Boolean) Whether the captcha should fail open.
- `secret` (String, Sensitive) Secret for Arkose Labs. Required when configuring Arkose Labs.
- `verify_subdomain` (String) Verify subdomain for Arkose Labs.


<a id="nestedblock--captcha--auth_challenge"></a>
### Nested Schema for `captcha.auth_challenge`

Optional:

- `fail_open` (Boolean) Whether the auth challenge should fail open.


<a id="nestedblock--captcha--friendly_captcha"></a>
### Nested Schema for `captcha.friendly_captcha`

Required:

- `site_key` (String) Site key for Friendly Captcha.

Optional:

- `secret` (String, Sensitive) Secret for Friendly Captcha. Required when configuring Friendly Captcha.


<a id="nestedblock--captcha--hcaptcha"></a>
### Nested Schema for `captcha.hcaptcha`

Required:

- `site_key` (String) Site key for hCaptcha.

Optional:

- `secret` (String, Sensitive) Secret for hCaptcha. Required when configuring hCaptcha.


<a id="nestedblock--captcha--recaptcha_enterprise"></a>
### Nested Schema for `captcha.recaptcha_enterprise`

Required:

- `project_id` (String) Project ID for reCAPTCHA Enterprise.
- `site_key` (String) Site key for reCAPTCHA Enterprise.

Optional:

- `api_key` (String, Sensitive) API key for reCAPTCHA Enterprise. Required when configuring reCAPTCHA Enterprise.


<a id="nestedblock--captcha--recaptcha_v2"></a>
### Nested Schema for `captcha.recaptcha_v2`

Required:

- `site_key` (String) Site key for reCAPTCHA v2.

Optional:

- `secret` (String, Sensitive) Secret for reCAPTCHA v2. Required when configuring reCAPTCHA v2.



<a id="nestedblock--suspicious_ip_throttling"></a>
### Nested Schema for `suspicious_ip_throttling`

Required:

- `enabled` (Boolean) Whether suspicious IP throttling attack protections are active.

Optional:

- `allowlist` (Set of String) List of trusted IP addresses that will not have attack protection enforced against them. This field allows you to specify multiple IP addresses, or ranges. You can use IPv4 or IPv6 addresses and CIDR notation.
- `pre_login` (Block List, Max: 1) Configuration options that apply before every login attempt. Only available on public tenants. (see [below for nested schema](#nestedblock--suspicious_ip_throttling--pre_login))
- `pre_user_registration` (Block List, Max: 1) Configuration options that apply before every user registration attempt. Only available on public tenants. (see [below for nested schema](#nestedblock--suspicious_ip_throttling--pre_user_registration))
- `shields` (Set of String) Action to take when a suspicious IP throttling threshold is violated. Possible values: `block` (throttle traffic from an IP address when there is a high number of login attempts targeting too many different accounts), `admin_notification` (send an email notification when traffic is throttled on one or more IP addresses due to high-velocity traffic).

<a id="nestedblock--suspicious_ip_throttling--pre_login"></a>
### Nested Schema for `suspicious_ip_throttling.pre_login`

Optional:

- `max_attempts` (Number) The maximum number of failed login attempts allowed from a single IP address.
- `rate` (Number) Interval of time, given in milliseconds at which new login tokens will become available after they have been used by an IP address. Each login attempt will be added on the defined throttling rate.


<a id="nestedblock--suspicious_ip_throttling--pre_user_registration"></a>
### Nested Schema for `suspicious_ip_throttling.pre_user_registration`

Optional:

- `max_attempts` (Number) The maximum number of sign up attempts allowed from a single IP address.
- `rate` (Number) Interval of time, given in milliseconds at which new sign up tokens will become available after they have been used by an IP address. Each sign up attempt will be added on the defined throttling rate.

## Import

Import is supported using the following syntax:

```shell
# As this is not a resource identifiable by an ID within the Auth0 Management API,
# attack_protection can be imported using a random string.
#
# We recommend [Version 4 UUID](https://www.uuidgenerator.net/version4)
#
# Example:
terraform import auth0_attack_protection.my_protection "24940d4b-4bd4-44e7-894e-f92e4de36a40"
```
