package sweep

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func init() {
	Actions()
	Clients()
	Connections()
	CustomDomains()
	Email()
	EmailTemplates()
	LogStreams()
	Organizations()
	ResourceServers()
	Roles()
	RuleConfigs()
	Users()
}

// This is needed so that the test
// sweepers get registered.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}
