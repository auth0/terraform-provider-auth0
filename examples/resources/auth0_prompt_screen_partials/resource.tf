resource "auth0_prompt_screen_partials" "prompt_screen_partials" {
  prompt_type = "login-passwordless"

  screen_partials {
    screen_name = "login-passwordless-email-code"
    insertion_points {
      form_content_start = "<div>Form Content Start</div>"
      form_content_end   = "<div>Form Content End</div>"
    }
  }

  screen_partials {
    screen_name = "login-passwordless-sms-otp"
    insertion_points {
      form_content_start = "<div>Form Content Start</div>"
      form_content_end   = "<div>Form Content End</div>"
    }
  }
}
