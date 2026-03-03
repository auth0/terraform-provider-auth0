package supplementalsignals

import (
	"github.com/auth0/go-auth0/v2/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// expandSupplementalSignals converts Terraform ResourceData to SDK struct.
func expandSupplementalSignals(data *schema.ResourceData) *management.UpdateSupplementalSignalsRequestContent {
	config := &management.UpdateSupplementalSignalsRequestContent{}

	if data.HasChange("akamai_enabled") {
		akamaiEnabled := data.Get("akamai_enabled").(bool)
		config.SetAkamaiEnabled(akamaiEnabled)
	}

	return config
}

// expandSupplementalSignalsForDelete resets configuration to default values.
func expandSupplementalSignalsForDelete() *management.UpdateSupplementalSignalsRequestContent {
	config := &management.UpdateSupplementalSignalsRequestContent{}

	config.SetAkamaiEnabled(false)

	return config
}
