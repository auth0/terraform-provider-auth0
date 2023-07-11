package user

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandUser(d *schema.ResourceData) (*management.User, error) {
	cfg := d.GetRawConfig()

	user := &management.User{}

	if d.IsNewResource() {
		user.ID = value.String(cfg.GetAttr("user_id"))
	}
	if d.HasChange("email") {
		user.Email = value.String(cfg.GetAttr("email"))
	}
	if d.HasChange("username") {
		user.Username = value.String(cfg.GetAttr("username"))
	}
	if d.HasChange("password") {
		user.Password = value.String(cfg.GetAttr("password"))
	}
	if d.HasChange("phone_number") {
		user.PhoneNumber = value.String(cfg.GetAttr("phone_number"))
	}
	if d.HasChange("email_verified") {
		user.EmailVerified = value.Bool(cfg.GetAttr("email_verified"))
	}
	if d.HasChange("verify_email") {
		user.VerifyEmail = value.Bool(cfg.GetAttr("verify_email"))
	}
	if d.HasChange("phone_verified") {
		user.PhoneVerified = value.Bool(cfg.GetAttr("phone_verified"))
	}
	if d.HasChange("given_name") {
		user.GivenName = value.String(cfg.GetAttr("given_name"))
	}
	if d.HasChange("family_name") {
		user.FamilyName = value.String(cfg.GetAttr("family_name"))
	}
	if d.HasChange("nickname") {
		user.Nickname = value.String(cfg.GetAttr("nickname"))
	}
	if d.HasChange("name") {
		user.Name = value.String(cfg.GetAttr("name"))
	}
	if d.HasChange("picture") {
		user.Picture = value.String(cfg.GetAttr("picture"))
	}
	if d.HasChange("blocked") {
		user.Blocked = value.Bool(cfg.GetAttr("blocked"))
	}
	if d.HasChange("connection_name") {
		user.Connection = value.String(cfg.GetAttr("connection_name"))
	}
	if d.HasChange("user_metadata") {
		userMetadata, err := expandMetadata(d, "user")
		if err != nil {
			return nil, err
		}
		user.UserMetadata = &userMetadata
	}
	if d.HasChange("app_metadata") {
		appMetadata, err := expandMetadata(d, "app")
		if err != nil {
			return nil, err
		}
		user.AppMetadata = &appMetadata
	}

	return user, nil
}

func expandMetadata(d *schema.ResourceData, metadataType string) (map[string]interface{}, error) {
	oldMetadata, newMetadata := d.GetChange(metadataType + "_metadata")
	if oldMetadata == "" {
		return value.MapFromJSON(d.GetRawConfig().GetAttr(metadataType + "_metadata"))
	}

	if newMetadata == "" {
		return map[string]interface{}{}, nil
	}

	oldMap, err := structure.ExpandJsonFromString(oldMetadata.(string))
	if err != nil {
		return map[string]interface{}{}, err
	}

	newMap, err := structure.ExpandJsonFromString(newMetadata.(string))
	if err != nil {
		return map[string]interface{}{}, err
	}

	for key := range oldMap {
		if _, ok := newMap[key]; !ok {
			newMap[key] = nil
		}
	}

	return newMap, nil
}
