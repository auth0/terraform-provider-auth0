package provider

import (
	"context"
	"net/http"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func newEmail() *schema.Resource {
	return &schema.Resource{
		CreateContext: createEmail,
		ReadContext:   readEmail,
		UpdateContext: updateEmail,
		DeleteContext: deleteEmail,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"default_from_address": {
				Type:     schema.TypeString,
				Required: true,
			},
			"credentials": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"api_user": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"api_key": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
							ForceNew:  true,
						},
						"access_key_id": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
							ForceNew:  true,
						},
						"secret_access_key": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
							ForceNew:  true,
						},
						"region": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"domain": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"smtp_host": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"smtp_port": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"smtp_user": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"smtp_pass": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
							ForceNew:  true,
						},
					},
				},
			},
		},
	}
}

func createEmail(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	email := buildEmail(d)
	api := m.(*management.Management)
	if err := api.Email.Create(email); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(auth0.StringValue(email.Name))

	return readEmail(ctx, d, m)
}

func readEmail(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)
	email, err := api.Email.Read()
	if err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return diag.FromErr(err)
	}

	d.SetId(auth0.StringValue(email.Name))

	result := multierror.Append(
		d.Set("name", email.Name),
		d.Set("enabled", email.Enabled),
		d.Set("default_from_address", email.DefaultFromAddress),
	)

	if credentials := email.Credentials; credentials != nil {
		credentialsMap := make(map[string]interface{})
		credentialsMap["api_user"] = credentials.APIUser
		credentialsMap["api_key"] = d.Get("credentials.0.api_key")
		credentialsMap["access_key_id"] = d.Get("credentials.0.access_key_id")
		credentialsMap["secret_access_key"] = d.Get("credentials.0.secret_access_key")
		credentialsMap["region"] = credentials.Region
		credentialsMap["domain"] = credentials.Domain
		credentialsMap["smtp_host"] = credentials.SMTPHost
		credentialsMap["smtp_port"] = credentials.SMTPPort
		credentialsMap["smtp_user"] = credentials.SMTPUser
		credentialsMap["smtp_pass"] = d.Get("credentials.0.smtp_pass")
		result = multierror.Append(result, d.Set("credentials", []map[string]interface{}{credentialsMap}))
	}

	return diag.FromErr(result.ErrorOrNil())
}

func updateEmail(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	email := buildEmail(d)
	api := m.(*management.Management)
	if err := api.Email.Update(email); err != nil {
		return diag.FromErr(err)
	}

	return readEmail(ctx, d, m)
}

func deleteEmail(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)
	if err := api.Email.Delete(); err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
	}

	return nil
}

func buildEmail(d *schema.ResourceData) *management.Email {
	email := &management.Email{
		Name:               String(d, "name"),
		Enabled:            Bool(d, "enabled"),
		DefaultFromAddress: String(d, "default_from_address"),
	}

	List(d, "credentials").Elem(func(d ResourceData) {
		email.Credentials = &management.EmailCredentials{
			APIUser:         String(d, "api_user"),
			APIKey:          String(d, "api_key"),
			AccessKeyID:     String(d, "access_key_id"),
			SecretAccessKey: String(d, "secret_access_key"),
			Region:          String(d, "region"),
			Domain:          String(d, "domain"),
			SMTPHost:        String(d, "smtp_host"),
			SMTPPort:        Int(d, "smtp_port"),
			SMTPUser:        String(d, "smtp_user"),
			SMTPPass:        String(d, "smtp_pass"),
		}
	})

	return email
}
