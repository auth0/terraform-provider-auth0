package prompt_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/auth0/terraform-provider-auth0/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
)

var (
	domain       = os.Getenv("AUTH0_DOMAIN")
	clientID     = os.Getenv("AUTH0_CLIENT_ID")
	clientSecret = os.Getenv("AUTH0_CLIENT_SECRET")
	manager, _   = management.New(domain, management.WithClientCredentials(context.Background(), clientID, clientSecret))
)

func TestAccPromptPartials(t *testing.T) {
	_ = givenACustomDomain(t)
	_ = givenAUniversalLogin(t)

	t.Cleanup(func() {
		cleanupPartialsPrompt(t, management.PartialsPromptSegment("login"))
	})

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccPromptPartialsCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0prompt_partials.prompt_partials", "prompt", "login"),
					resource.TestCheckResourceAttr("auth0prompt_partials.prompt_partials", "form_content_start", "<div>Test Header</div>"),
				),
			},
			{
				Config: testAccPromptPartialsUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0prompt_partials.prompt_partials", "prompt", "login"),
					resource.TestCheckResourceAttr("auth0prompt_partials.prompt_partials", "form_content_start", "<div>Updated Test Header</div>"),
				),
			},
		},
	})
}

func givenACustomDomain(t *testing.T) *management.CustomDomain {
	t.Helper()

	customDomain := &management.CustomDomain{
		Domain:    auth0.Stringf("%d.auth.uat.auth0.com", time.Now().UTC().Unix()),
		Type:      auth0.String("auth0_managed_certs"),
		TLSPolicy: auth0.String("recommended"),
	}

	err := manager.CustomDomain.Create(context.Background(), customDomain)
	assert.NoError(t, err)

	t.Cleanup(func() {
		cleanupCustomDomain(t, customDomain.GetID())
	})

	return customDomain
}

func givenAUniversalLogin(t *testing.T) *management.BrandingUniversalLogin {
	t.Helper()

	body := `<!DOCTYPE html><html><head>{%- auth0:head -%}</head><body>{%- auth0:widget -%}</body></html>`
	ul := &management.BrandingUniversalLogin{
		Body: auth0.String(body),
	}

	err := manager.Branding.SetUniversalLogin(context.Background(), ul)
	assert.NoError(t, err)

	t.Cleanup(func() {
		cleanupUniversalLogin(t)
	})

	return ul
}

func cleanupCustomDomain(t *testing.T, customDomainID string) {
	t.Helper()

	err := manager.CustomDomain.Delete(context.Background(), customDomainID)
	assert.NoError(t, err)
}

func cleanupUniversalLogin(t *testing.T) {
	t.Helper()

	err := manager.Branding.DeleteUniversalLogin(context.Background())
	assert.NoError(t, err)
}

func cleanupPartialsPrompt(t *testing.T, prompt management.PartialsPromptSegment) {
	t.Helper()

	err := manager.Prompt.DeletePartials(context.Background(), &management.PartialsPrompt{Segment: prompt})
	assert.NoError(t, err)
}

const testAccPromptPartialsCreate = `
resource "auth0_prompt_partials" "prompt_partials" {
  prompt = "login"
  form_content_start = "<div>Test Header</div>"
}
`

const testAccPromptPartialsUpdate = `
resource "auth0_prompt_partials" "prompt_partials" {
  prompt = "login"
  form_content_start = "<div>Updated Test Header</div>"
}
`
