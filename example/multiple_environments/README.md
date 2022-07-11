# Multiple Environments

This example demonstrates how you can share tenant configurations between different environments by leveraging Terraform modules.

It has configurations for three tenants: [prod](prod), [stage](stage), and a [custom](custom) tenant where the Auth0
domain is set by a Terraform variable or environment variable. Each of these tenants is set up identically with the
included [admin_console](modules/admin_console) module.
