resource "auth0_global_client" "global" {
  // Auth0 Universal Login - Custom Login Page
  custom_login_page_on = true
  custom_login_page    = <<PAGE
<html>
    <head><title>My Custom Login Page</title></head>
    <body>
        I should probably have a login form here
    </body>
</html>
PAGE
  callbacks            = ["http://somehostname.com/a/callback"]
}
