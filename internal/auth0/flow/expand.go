package flow

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func expandFlow(data *schema.ResourceData) (*management.Flow, error) {
	flow := &management.Flow{}

	return flow, nil
}
