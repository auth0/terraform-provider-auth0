resource "auth0_encryption_key" "my_encryption_keys_dont_rekey" {
}

resource "auth0_encryption_key" "my_encryption_keys_rekey" {
  rekey = true
}

