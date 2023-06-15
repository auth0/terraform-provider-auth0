package connection_new //nolint:all

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var baseSchema = map[string]*schema.Schema{
	"name": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "Name of the connection.",
	},
	"display_name": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Name used in login screen.",
	},
	"is_domain_connection": {
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
		Description: "Indicates whether the connection is domain level.",
	},
	"strategy": {
		Type:     schema.TypeString,
		Required: true,
		ValidateFunc: validation.StringInSlice([]string{
			"ad", "adfs", "amazon", "apple", "dropbox", "bitbucket", "aol",
			"auth0-adldap", "auth0-oidc", "auth0", "baidu", "bitly",
			"box", "custom", "daccount", "dwolla", "email",
			"evernote-sandbox", "evernote", "exact", "facebook",
			"fitbit", "flickr", "github", "google-apps",
			"google-oauth2", "guardian", "instagram", "ip", "linkedin",
			"miicard", "oauth1", "oauth2", "office365", "oidc", "okta", "paypal",
			"paypal-sandbox", "pingfederate", "planningcenter",
			"renren", "salesforce-community", "salesforce-sandbox",
			"salesforce", "samlp", "sharepoint", "shopify", "sms",
			"soundcloud", "thecity-sandbox", "thecity",
			"thirtysevensignals", "twitter", "untappd", "vkontakte",
			"waad", "weibo", "windowslive", "wordpress", "yahoo",
			"yammer", "yandex", "line",
		}, true),
		ForceNew:    true,
		Description: "Type of the connection, which indicates the identity provider.",
	},
	"metadata": {
		Type:             schema.TypeMap,
		Elem:             &schema.Schema{Type: schema.TypeString},
		Optional:         true,
		ValidateDiagFunc: validation.MapKeyLenBetween(0, 10),
		Description: "Metadata associated with the connection, in the form of a map of string values " +
			"(max 255 chars). Maximum of 10 metadata properties allowed.",
	},
	"realms": {
		Type:     schema.TypeList,
		Elem:     &schema.Schema{Type: schema.TypeString},
		Optional: true,
		Computed: true,
		Description: "Defines the realms for which the connection will be used (e.g., email domains). " +
			"If not specified, the connection name is added as the realm.",
	},
	"show_as_button": {
		Type:        schema.TypeBool,
		Optional:    true,
		Description: "Display connection as a button. Only available on enterprise connections.",
	},
	"enabled_clients": {
		Type: schema.TypeSet,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		Computed:    true,
		Description: "IDs of the clients for which the connection is enabled.",
	},
	"upstream_params": {
		Type:         schema.TypeString,
		Optional:     true,
		ValidateFunc: validation.StringIsJSON,
		Description: "You can pass provider-specific parameters to an identity provider during " +
			"authentication. The values can either be static per connection or dynamic per user.",
	},
	"non_persistent_attrs": {
		Type:     schema.TypeSet,
		Elem:     &schema.Schema{Type: schema.TypeString},
		Optional: true,
		Computed: true,
		Description: "If there are user fields that should not be stored in Auth0 databases due to " +
			"privacy reasons, you can add them to the DenyList here.",
	},
}
