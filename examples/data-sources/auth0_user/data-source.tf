# An Auth0 User loaded using its ID.
data "auth0_user" "my_user" {
  user_id = "auth0|34fdr23fdsfdfsf"
}

# An Auth0 User loaded through Lucene query.
data "auth0_user" "my_user" {
  query = "email:testemail@gmail.com"
}

# An Auth0 User loaded through Lucene query.
data "auth0_user" "my_user" {
  query = "username:johndoe"
}
