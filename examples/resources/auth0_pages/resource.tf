resource "auth0_pages" "my_pages" {
  login {
    enabled = true
    html    = "<html><body>My Custom Login Page</body></html>"
  }

  change_password {
    enabled = true
    html    = "<html><body>My Custom Reset Password Page</body></html>"
  }

  guardian_mfa {
    enabled = true
    html    = "<html><body>My Custom MFA Page</body></html>"
  }

  error {
    show_log_link = true
    html          = "<html><body>My Custom Error Page</body></html>"
    url           = "https://example.com"
  }
}
