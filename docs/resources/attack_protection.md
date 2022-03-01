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

* `enabled` - (Optional) Boolean if feature enabled
* `shields` - (Optional)
* `allowlist` - (Optional) 
* `mode` - (Optional)
* `max_attempts` - (Optional) Number of 

### suspicious_ip_throttling

The following arguments are supported for `suspicious_ip_throttling`:

* `enabled` - (Optional) Boolean if feature enabled
* `shields` - (Optional)
* `allowlist` - (Optional) 
* `pre_login` - (Optional)
* `pre_user_registration` - (Optional) 

### brute_force_protection

* `enabled` - (Optional) Boolean if feature enabled
* `shields` - (Optional)
* `allowlist` - (Optional)
* `mode` - (Optional)
* `max_attempts` - (Optional)


## Import

As this is not a resource identifiable by an ID within the Auth0 Management API, guardian can be imported using a random
string. We recommend [Version 4 UUID](https://www.uuidgenerator.net/version4) e.g.

```shell
$ terraform import auth0_guardian.default 24940d4b-4bd4-44e7-894e-f92e4de36a40
```