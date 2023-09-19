package user

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandUser(data *schema.ResourceData) (*management.User, error) {
	cfg := data.GetRawConfig()

	user := &management.User{}

	if data.IsNewResource() {
		user.ID = value.String(cfg.GetAttr("user_id"))
	}
	if data.HasChange("email") {
		user.Email = value.String(cfg.GetAttr("email"))
	}
	if data.HasChange("username") {
		user.Username = value.String(cfg.GetAttr("username"))
	}
	if data.HasChange("password") {
		user.Password = value.String(cfg.GetAttr("password"))
	}
	if data.HasChange("phone_number") {
		user.PhoneNumber = value.String(cfg.GetAttr("phone_number"))
	}
	if data.HasChange("email_verified") {
		user.EmailVerified = value.Bool(cfg.GetAttr("email_verified"))
	}
	if data.HasChange("verify_email") {
		user.VerifyEmail = value.Bool(cfg.GetAttr("verify_email"))
	}
	if data.HasChange("phone_verified") {
		user.PhoneVerified = value.Bool(cfg.GetAttr("phone_verified"))
	}
	if data.HasChange("given_name") {
		user.GivenName = value.String(cfg.GetAttr("given_name"))
	}
	if data.HasChange("family_name") {
		user.FamilyName = value.String(cfg.GetAttr("family_name"))
	}
	if data.HasChange("nickname") {
		user.Nickname = value.String(cfg.GetAttr("nickname"))
	}
	if data.HasChange("name") {
		user.Name = value.String(cfg.GetAttr("name"))
	}
	if data.HasChange("picture") {
		user.Picture = value.String(cfg.GetAttr("picture"))
	}
	if data.HasChange("blocked") {
		user.Blocked = value.Bool(cfg.GetAttr("blocked"))
	}
	if data.HasChange("connection_name") {
		user.Connection = value.String(cfg.GetAttr("connection_name"))
	}
	if data.HasChange("user_metadata") {
		userMetadata, err := expandMetadata(data, "user")
		if err != nil {
			return nil, err
		}
		user.UserMetadata = &userMetadata
	}
	if data.HasChange("app_metadata") {
		appMetadata, err := expandMetadata(data, "app")
		if err != nil {
			return nil, err
		}
		user.AppMetadata = &appMetadata
	}

	return user, nil
}

func expandMetadata(data *schema.ResourceData, metadataType string) (map[string]interface{}, error) {
	oldMetadata, newMetadata := data.GetChange(metadataType + "_metadata")
	if oldMetadata == "" {
		return value.MapFromJSON(data.GetRawConfig().GetAttr(metadataType + "_metadata"))
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
