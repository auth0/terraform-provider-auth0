# This resource can be imported using the pre-defined template name.
#
# These names are `verify_email`, `verify_email_by_code`, `reset_email`, `reset_email_by_code`,
# `welcome_email`, `blocked_account`, `stolen_credentials`, `enrollment_email`,
# `mfa_oob_code`, `user_invitation`, and `async_approval`.
#
# The names `change_password`, and `password_reset` are also supported
# for legacy scenarios.
#
# Example:
terraform import auth0_email_template.my_email_template "welcome_email"
