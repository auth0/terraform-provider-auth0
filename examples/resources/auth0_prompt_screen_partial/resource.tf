resource "auth0_prompt_screen_partial" "login" {
  prompt_type = "login"
  screen_name = "login"
  insertion_points {
    form_content       = "<div>Form Content</div>"
    form_content_start = "<div>Form Content Start</div>"
    form_content_end   = "<div>Form Content End</div>"
  }
}

resource "auth0_prompt_screen_partial" "customized-consent" {
  prompt_type = "customized-consent"
  screen_name = "customized-consent"
  insertion_points {
    form_content = "<div>Form Content</div>"
  }
}
