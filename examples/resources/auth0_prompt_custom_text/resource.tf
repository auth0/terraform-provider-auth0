resource "auth0_prompt_custom_text" "example" {
  prompt   = "login"
  language = "en"
  body = jsonencode(
    {
      "login" : {
        "alertListTitle" : "Alerts",
        "buttonText" : "Continue",
        "description" : "Login to",
        "editEmailText" : "Edit",
        "emailPlaceholder" : "Email address",
        "federatedConnectionButtonText" : "Continue with ${connectionName}",
        "footerLinkText" : "Sign up",
        "footerText" : "Don't have an account?",
        "forgotPasswordText" : "Forgot password?",
        "invitationDescription" : "Log in to accept ${inviterName}'s invitation to join ${companyName} on ${clientName}.",
        "invitationTitle" : "You've Been Invited!",
        "logoAltText" : "${companyName}",
        "pageTitle" : "Log in | ${clientName}",
        "passwordPlaceholder" : "Password",
        "separatorText" : "Or",
        "signupActionLinkText" : "${footerLinkText}",
        "signupActionText" : "${footerText}",
        "title" : "Welcome",
        "usernamePlaceholder" : "Username or email address"
      }
    }
  )
}
