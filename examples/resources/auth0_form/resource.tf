# Example:
resource "auth0_form" "my_form" {
  name = "My KYC Form"

  start = jsonencode({
    coordinates = {
      x = 0
      y = 0
    }
    next_node = "step_ggeX"
  })
  nodes = jsonencode([{
    alias = "New step"
    config = {
      components = [{
        category = "FIELD"
        config = {
          max_length = 50
          min_length = 1
          multiline  = false
        }
        id        = "full_name"
        label     = "Your Name"
        required  = true
        sensitive = false
        type      = "TEXT"
        }, {
        category = "BLOCK"
        config = {
          text = "Continue"
        }
        id   = "next_button_3FbA"
        type = "NEXT_BUTTON"
      }]
      next_node = "$ending"
    }
    coordinates = {
      x = 500
      y = 0
    }
    id   = "step_ggeX"
    type = "STEP"
  }])
  ending = jsonencode({
    after_submit = {
      flow_id = "<my_flow_id>" # Altenative ways: (flow_id = auth0_flow.my_flow.id) or using terraform variables
    }
    coordinates = {
      x = 1250
      y = 0
    }
    resume_flow = true
  })

  style = jsonencode({
    css = "h1 {\n  color: white;\n  text-align: center;\n}"
  })

  translations = jsonencode({
    es = {
      components = {
        rich_text_uctu = {
          config = {
            content = "<h2>Help us verify your personal information</h2><p>We want to learn more about you so that we can validate and protect your account...</p>"
          }
        }
      }
      messages = {
        custom = {}
        errors = {
          ERR_ACCEPTANCE_REQUIRED = "Por favor, marca este campo para continuar."
        }
      }
    }
  })

  messages {
    errors = jsonencode({
      ERR_REQUIRED_PROPERTY = "This field is required for user kyc."
    })
  }
  languages {
    default = "en"
    primary = "en"
  }
}
