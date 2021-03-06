---
layout: "auth0"
page_title: "Auth0: auth0_rule_config"
description: |-
  With this resource, you can create and manage variables for rules, which are custom Javascript snippets that run in a 
  secure, isolate sandbox as part of your authentication pipeline.
---

# auth0_rule_config

With Auth0, you can create custom Javascript snippets that run in a secure, isolated sandbox as part of your 
authentication pipeline, which are otherwise known as rules. This resource allows you to create and manage variables 
that are available to all rules via Auth0's global configuration object. Used in conjunction with configured rules.

## Example Usage

```hcl
resource "auth0_rule" "my_rule" {
  name = "empty-rule"
  script = <<EOF
function (user, context, callback) {
  callback(null, user, context);
}
EOF
  enabled = true
}

resource "auth0_rule_config" "my_rule_config" {
  key = "foo"
  value = "bar"
}
```

## Argument Reference

Arguments accepted by this resource include:

* `key` - (Required) String. Key for a rules configuration variable.
* `value` - (Required) String, Case-sensitive. Value for a rules configuration variable.

## Attributes Reference

No additional attributes are exported by this resource.

## Import

Existing rule configs can be imported using their key name, e.g.

```shell
$ terraform import auth0_rule_config.my_rule_config foo
```
