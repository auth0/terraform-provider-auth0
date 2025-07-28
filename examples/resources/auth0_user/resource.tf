resource "auth0_user" "user" {
  connection_name = "Username-Password-Authentication"
  user_id         = "12345"
  username        = "unique_username"
  name            = "Firstname Lastname"
  nickname        = "some.nickname"
  email           = "test@test.com"
  email_verified  = true
  password        = "passpass$12$12"
  picture         = "https://www.example.com/a-valid-picture-url.jpg"
}


# Create a user with custom_domain_header
resource "auth0_user" "auth0_user_with_custom_domain" {
  connection_name      = "Username-Password-Authentication"
  username             = "your_new_user_"
  email                = "change.username@acceptance.test.com"
  email_verified       = true
  password             = "MyPass123$"
  custom_domain_header = "my-custom.domain.org"
}
