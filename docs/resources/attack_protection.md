---
layout: "auth0"
page_title: "Auth0: auth0_attack_protection"
description: |-
  Auth0 can detect attacks and stop malicious attempts to access your application such as blocking traffic from certain IPs and displaying CAPTCHA.
---

# auth0_attack_protection

Auth0 can detect attacks and stop malicious attempts to access your application such as blocking traffic from certain IPs and displaying CAPTCHA

## Example Usage

```hcl
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
```

## Argument Reference

The following arguments are supported:

* `breached_password_detection` - (Optional) Breached password detection protects your applications from bad actors logging in with stolen credentials. 
* `suspicious_ip_throttling` - (Optional) Suspicious IP throttling blocks traffic from any IP address that rapidly attempts too many logins or signups.
* `brute_force_protection` - (Optional) Safeguards against a single IP address attacking a single user account.

### brute_force_protection

The following arguments are supported for `brute_force_protection`:

* `enabled` - (Optional) Whether or not brute force attack protections are active.
* `shields` - (Optional) Action to take when a brute force protection threshold is violated. Possible values: `block`, `user_notification`.
* `allowlist` - (Optional) List of trusted IP addresses that will not have attack protection enforced against them.
* `mode` - (Optional) Determines whether or not IP address is used when counting failed attempts. Possible values: `count_per_identifier_and_ip` or `count_per_identifier`.
* `max_attempts` - (Optional) Maximum number of unsuccessful attempts.

### suspicious_ip_throttling

The following arguments are supported for `suspicious_ip_throttling`:

* `enabled` - (Optional) Whether or not suspicious IP throttling attack protections are active.
* `shields` - (Optional) Action to take when a suspicious IP throttling threshold is violated. Possible values: `block`, `admin_notification`.
* `allowlist` - (Optional) List of trusted IP addresses that will not have attack protection enforced against them. 
* `pre_login` - (Optional) Configuration options that apply before every login attempt.
* `pre_user_registration` - (Optional) Configuration options that apply before every user registration attempt.

### breached_password_protection

* `enabled` - (Optional) Whether or not breached password detection is active.
* `shields` - (Optional) Action to take when a breached password is detected. Possible values: `block`, `user_notification`, `admin_notification`.
* `admin_notification_frequency` - (Optional) When "admin_notification" is enabled, determines how often email notifications are sent. Possible values: `immediately`, `daily`, `weekly`, `monthly`.
* `method` - (Optional) The subscription level for breached password detection methods. Use "enhanced" to enable Credential Guard. Possible values: `standard`, `enhanced`.


## Import

As this is not a resource identifiable by an ID within the Auth0 Management API, guardian can be imported using a random
string. We recommend [Version 4 UUID](https://www.uuidgenerator.net/version4) e.g.

```shell
$ terraform import auth0_guardian.default 24940d4b-4bd4-44e7-894e-f92e4de36a40
```