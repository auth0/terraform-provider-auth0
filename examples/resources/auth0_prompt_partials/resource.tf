resource "auth0_prompt_partials" "my_login_prompt_partials" {
  prompt = "login"

  form_content_start      = "<div>Updated Form Content Start</div>"
  form_content_end        = "<div>Updated Form Content End</div>"
  form_footer_start       = "<div>Updated Footer Start</div>"
  form_footer_end         = "<div>Updated Footer End</div>"
  secondary_actions_start = "<div>Updated Secondary Actions Start</div>"
  secondary_actions_end   = "<div>Updated Secondary Actions End</div>"
}
