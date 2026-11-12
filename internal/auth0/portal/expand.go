package portal

import (
	"github.com/auth0/go-auth0/v3/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandPortalCreate(data *schema.ResourceData) *management.CreatePortalRequestContent {
	raw := data.GetRawConfig()

	return &management.CreatePortalRequestContent{
		Name:       data.Get("name").(string),
		Slug:       data.Get("slug").(string),
		Client:     expandClient(raw),
		Navigation: expandNavigation(raw),
		Pages:      expandPages(raw),
		CustomCode: expandCustomCode(raw),
	}
}

// expandPortalUpdate sends only the changed top-level fields, because the API merges the top level.
// A changed object is sent complete, because the API replaces each sent object whole, including the
// client, which must include the secret. A block removed from the configuration is sent as `null`, which clears it.
func expandPortalUpdate(data *schema.ResourceData) *management.UpdatePortalRequestContent {
	raw := data.GetRawConfig()
	request := &management.UpdatePortalRequestContent{}

	if data.HasChange("name") {
		name := data.Get("name").(string)
		request.SetName(&name)
	}

	if data.HasChange("slug") {
		slug := data.Get("slug").(string)
		request.SetSlug(&slug)
	}

	if data.HasChange("client") {
		request.SetClient(expandClient(raw))
	}

	if data.HasChange("navigation") {
		navigation := expandNavigation(raw)
		request.SetNavigation(&navigation)
	}

	if data.HasChange("pages") {
		pages := expandPages(raw)
		request.SetPages(&pages)
	}

	if data.HasChange("custom_code") {
		customCode := expandCustomCode(raw)
		request.SetCustomCode(&customCode)
	}

	return request
}

func expandClient(raw cty.Value) *management.PortalClientSecretPostRequest {
	client := value.FirstBlock(value.Attr(raw, "client"))
	secretPost := value.FirstBlock(value.Attr(client, "client_secret_post"))
	if !value.IsSet(secretPost) {
		return nil
	}

	secretName := "client_secret_wo"
	if !value.IsSet(value.Attr(secretPost, secretName)) {
		secretName = "client_secret"
	}

	return &management.PortalClientSecretPostRequest{
		TokenEndpointAuthMethod: management.PortalClientSecretPostRequestTokenEndpointAuthMethodClientSecretPost,
		ClientID:                stringAttr(secretPost, "client_id"),
		ClientSecret:            stringAttr(secretPost, secretName),
	}
}

func expandCustomCode(raw cty.Value) *management.PortalCustomCode {
	customCode := value.FirstBlock(value.Attr(raw, "custom_code"))
	if !value.IsSet(customCode) {
		return nil
	}

	return &management.PortalCustomCode{
		Head: value.String(value.Attr(customCode, "head")),
	}
}

func expandNavigation(raw cty.Value) *management.PortalNavigation {
	navigation := value.FirstBlock(value.Attr(raw, "navigation"))
	if !value.IsSet(navigation) {
		return nil
	}

	sidebar := value.FirstBlock(value.Attr(navigation, "sidebar"))
	if !value.IsSet(sidebar) {
		return &management.PortalNavigation{}
	}

	var components []*management.SidebarComponent
	for _, component := range blocks(value.Attr(sidebar, "component")) {
		components = append(components, expandSidebarComponent(component))
	}

	return &management.PortalNavigation{
		Sidebar: &management.Sidebar{Components: components},
	}
}

func expandSidebarComponent(component cty.Value) *management.SidebarComponent {
	sidebarComponent := &management.SidebarComponent{}

	if link := value.FirstBlock(value.Attr(component, "external_link_v2")); value.IsSet(link) {
		sidebarComponent.SidebarComponentAuth0ExternalLinkV2 = &management.SidebarComponentAuth0ExternalLinkV2{
			Config: &management.SidebarComponentAuth0ExternalLinkV2Config{
				Label: stringAttr(link, "label"),
				Href:  stringAttr(link, "href"),
				Icon:  value.String(value.Attr(link, "icon")),
			},
		}
	}

	if link := value.FirstBlock(value.Attr(component, "internal_link_v2")); value.IsSet(link) {
		sidebarComponent.SidebarComponentAuth0InternalLinkV2 = &management.SidebarComponentAuth0InternalLinkV2{
			Config: &management.SidebarComponentAuth0InternalLinkV2Config{
				Label:    stringAttr(link, "label"),
				PageSlug: stringAttr(link, "page_slug"),
				Icon:     value.String(value.Attr(link, "icon")),
			},
		}
	}

	return sidebarComponent
}

func expandPages(raw cty.Value) *management.PortalPagesByType {
	pages := value.FirstBlock(value.Attr(raw, "pages"))
	if !value.IsSet(pages) {
		return nil
	}

	portalPages := &management.PortalPagesByType{
		Default: value.String(value.Attr(pages, "default")),
	}

	for _, page := range blocks(value.Attr(pages, "standard")) {
		portalPage := &management.PortalPage{
			Title: stringAttr(page, "title"),
			Slug:  stringAttr(page, "slug"),
		}

		for _, component := range blocks(value.Attr(page, "component")) {
			portalPage.Components = append(portalPage.Components, expandPageComponent(component))
		}

		portalPages.Standard = append(portalPages.Standard, portalPage)
	}

	return portalPages
}

// expandPageComponent expands the section component type, and reuses expandComponentType for the component types
// that a page component shares with a section child.
func expandPageComponent(component cty.Value) *management.PageComponent {
	if section := value.FirstBlock(value.Attr(component, "structure_section")); value.IsSet(section) {
		return &management.PageComponent{
			PageComponentAuth0StructureSection: expandStructureSection(section),
		}
	}

	child := expandComponentType(component)

	return &management.PageComponent{
		PageComponentAuth0Form:                           child.PageComponentAuth0Form,
		PageComponentAuth0TypographyHeading:              child.PageComponentAuth0TypographyHeading,
		PageComponentAuth0MediaImage:                     child.PageComponentAuth0MediaImage,
		PageComponentAuth0TypographyRichText:             child.PageComponentAuth0TypographyRichText,
		PageComponentAuth0StructureSeparator:             child.PageComponentAuth0StructureSeparator,
		PageComponentAuth0MyOrganizationDomainTable:      child.PageComponentAuth0MyOrganizationDomainTable,
		PageComponentAuth0MyOrganizationMemberManagement: child.PageComponentAuth0MyOrganizationMemberManagement,
		PageComponentAuth0MyOrganizationDetailsEdit:      child.PageComponentAuth0MyOrganizationDetailsEdit,
		PageComponentAuth0MyAccountPasskeyManagement:     child.PageComponentAuth0MyAccountPasskeyManagement,
		PageComponentAuth0MyAccountMfaManagement:         child.PageComponentAuth0MyAccountMfaManagement,
	}
}

func expandStructureSection(section cty.Value) *management.PageComponentAuth0StructureSection {
	config := &management.PageComponentAuth0StructureSectionConfig{
		Variant:     management.PageComponentAuth0StructureSectionVariantEnum(stringAttr(section, "variant")),
		Title:       value.String(value.Attr(section, "title")),
		Description: value.String(value.Attr(section, "description")),
		// The API requires `children` in the request, at least as `[]`. A nil slice would be sent as `null`.
		Children: []*management.PageComponentNoChildren{},
	}

	for _, child := range blocks(value.Attr(section, "child")) {
		config.Children = append(config.Children, expandComponentType(child))
	}

	return &management.PageComponentAuth0StructureSection{Config: config}
}

// expandComponentType expands the component types that are allowed both on a page and in a section.
// Component types without config are sent without a `config` key, which the API keeps omitted.
func expandComponentType(component cty.Value) *management.PageComponentNoChildren {
	child := &management.PageComponentNoChildren{}
	typeBlock := func(name string) cty.Value {
		return value.FirstBlock(value.Attr(component, name))
	}

	if form := typeBlock("form"); value.IsSet(form) {
		child.PageComponentAuth0Form = &management.PageComponentAuth0Form{
			Config: &management.PageComponentAuth0FormConfig{
				FormID:            stringAttr(form, "form_id"),
				CompletionMessage: stringAttr(form, "completion_message"),
			},
		}
	}

	if heading := typeBlock("typography_heading"); value.IsSet(heading) {
		child.PageComponentAuth0TypographyHeading = &management.PageComponentAuth0TypographyHeading{
			Config: &management.PageComponentAuth0TypographyHeadingConfig{
				Title:       stringAttr(heading, "title"),
				Description: value.String(value.Attr(heading, "description")),
			},
		}
	}

	if richText := typeBlock("typography_rich_text"); value.IsSet(richText) {
		child.PageComponentAuth0TypographyRichText = &management.PageComponentAuth0TypographyRichText{
			Config: &management.PageComponentAuth0TypographyRichTextConfig{
				Content: stringAttr(richText, "content"),
			},
		}
	}

	if image := typeBlock("media_image"); value.IsSet(image) {
		config := &management.PageComponentAuth0MediaImageConfig{
			Src:    stringAttr(image, "src"),
			Height: value.Int(value.Attr(image, "height")),
		}
		if position := value.String(value.Attr(image, "position")); position != nil {
			config.Position = (*management.Position)(position)
		}

		child.PageComponentAuth0MediaImage = &management.PageComponentAuth0MediaImage{Config: config}
	}

	if separator := typeBlock("structure_separator"); value.IsSet(separator) {
		config := &management.PageComponentAuth0StructureSeparatorConfig{
			Text: value.String(value.Attr(separator, "text")),
		}
		if separatorVariant := value.String(value.Attr(separator, "variant")); separatorVariant != nil {
			config.Variant = (*management.PageComponentAuth0StructureSeparatorVariantEnum)(separatorVariant)
		}

		child.PageComponentAuth0StructureSeparator = &management.PageComponentAuth0StructureSeparator{}
		if config.Text != nil || config.Variant != nil {
			child.PageComponentAuth0StructureSeparator.Config = config
		}
	}

	if value.IsSet(typeBlock("my_organization_domain_table")) {
		child.PageComponentAuth0MyOrganizationDomainTable = &management.PageComponentAuth0MyOrganizationDomainTable{}
	}

	if value.IsSet(typeBlock("my_organization_member_management")) {
		child.PageComponentAuth0MyOrganizationMemberManagement = &management.PageComponentAuth0MyOrganizationMemberManagement{}
	}

	if value.IsSet(typeBlock("my_organization_details_edit")) {
		child.PageComponentAuth0MyOrganizationDetailsEdit = &management.PageComponentAuth0MyOrganizationDetailsEdit{}
	}

	if value.IsSet(typeBlock("my_account_passkey_management")) {
		child.PageComponentAuth0MyAccountPasskeyManagement = &management.PageComponentAuth0MyAccountPasskeyManagement{}
	}

	if value.IsSet(typeBlock("my_account_mfa_management")) {
		child.PageComponentAuth0MyAccountMfaManagement = &management.PageComponentAuth0MyAccountMfaManagement{}
	}

	return child
}

// stringAttr returns the string attribute of the object, or an empty string when it is not set.
func stringAttr(object cty.Value, name string) string {
	if v := value.String(value.Attr(object, name)); v != nil {
		return *v
	}

	return ""
}
