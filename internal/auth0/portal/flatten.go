package portal

import (
	"fmt"

	"github.com/auth0/go-auth0/v3/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

// flattenPortal fails before it sets any attribute when the portal has a component that the schema
// cannot represent, so that no component is dropped from the state without notice.
func flattenPortal(data *schema.ResourceData, portal *management.Portal) diag.Diagnostics {
	navigation, err := flattenNavigation(portal.Navigation)
	if err != nil {
		return diag.FromErr(err)
	}

	pages, err := flattenPages(portal.Pages)
	if err != nil {
		return diag.FromErr(err)
	}

	result := multierror.Append(
		data.Set("name", portal.GetName()),
		data.Set("slug", portal.GetSlug()),
		data.Set("client", flattenClient(data, portal.Client)),
		data.Set("custom_code", flattenCustomCode(portal.CustomCode)),
		data.Set("navigation", navigation),
		data.Set("pages", pages),
		data.Set("created_at", value.FormatTime(portal.GetCreatedAt())),
		data.Set("updated_at", value.FormatTime(portal.GetUpdatedAt())),
	)

	return diag.FromErr(result.ErrorOrNil())
}

// flattenClient keeps the secret attributes from the configuration, because the API never returns the secret.
func flattenClient(data *schema.ResourceData, client *management.PortalClientSecretPost) []interface{} {
	if client == nil {
		return nil
	}

	secretPost := map[string]interface{}{
		"client_id": client.GetClientID(),
	}

	// The secret is never returned by the API, so it is kept from the prior state. The write-only
	// secret is never kept, as Terraform requires it to be null in state.
	if v, ok := data.GetOk("client.0.client_secret_post.0.client_secret"); ok {
		secretPost["client_secret"] = v
	}
	if v, ok := data.GetOk("client.0.client_secret_post.0.client_secret_wo_version"); ok {
		secretPost["client_secret_wo_version"] = v
	}

	return []interface{}{map[string]interface{}{
		"client_secret_post": []interface{}{secretPost},
	}}
}

func flattenCustomCode(customCode *management.PortalCustomCode) []interface{} {
	if customCode == nil {
		return nil
	}

	return []interface{}{map[string]interface{}{
		"head": customCode.GetHead(),
	}}
}

func flattenNavigation(navigation *management.PortalNavigation) ([]interface{}, error) {
	if navigation == nil {
		return nil, nil
	}

	// A navigation without a sidebar is kept as an empty block, as the API returns it as `{}`.
	if navigation.Sidebar == nil {
		return []interface{}{map[string]interface{}{}}, nil
	}

	components := make([]interface{}, 0, len(navigation.Sidebar.GetComponents()))
	for i, component := range navigation.Sidebar.GetComponents() {
		path := fmt.Sprintf("navigation.sidebar.components[%d]", i)

		switch {
		case component.SidebarComponentAuth0ExternalLinkV2 != nil:
			config := component.SidebarComponentAuth0ExternalLinkV2.Config
			components = append(components, flattenComponentType("external_link_v2", map[string]interface{}{
				"label": config.GetLabel(),
				"href":  config.GetHref(),
				"icon":  config.GetIcon(),
			}))
		case component.SidebarComponentAuth0InternalLinkV2 != nil:
			config := component.SidebarComponentAuth0InternalLinkV2.Config
			components = append(components, flattenComponentType("internal_link_v2", map[string]interface{}{
				"label":     config.GetLabel(),
				"page_slug": config.GetPageSlug(),
				"icon":      config.GetIcon(),
			}))
		default:
			return nil, unsupportedComponentError(path, component.Type)
		}
	}

	return []interface{}{map[string]interface{}{
		"sidebar": []interface{}{map[string]interface{}{
			"component": components,
		}},
	}}, nil
}

func flattenPages(pages *management.PortalPagesByType) ([]interface{}, error) {
	if pages == nil {
		return nil, nil
	}

	standard := make([]interface{}, 0, len(pages.GetStandard()))
	for i, page := range pages.GetStandard() {
		components := make([]interface{}, 0, len(page.GetComponents()))
		for j, component := range page.GetComponents() {
			flattened, err := flattenPageComponent(component, fmt.Sprintf("pages.standard[%d].components[%d]", i, j))
			if err != nil {
				return nil, err
			}

			components = append(components, flattened)
		}

		standard = append(standard, map[string]interface{}{
			"title":     page.GetTitle(),
			"slug":      page.GetSlug(),
			"component": components,
		})
	}

	return []interface{}{map[string]interface{}{
		"default":  pages.GetDefault(),
		"standard": standard,
	}}, nil
}

// flattenPageComponent flattens the section component type, and reuses flattenSectionChild for the component types
// that a page component shares with a section child.
func flattenPageComponent(component *management.PageComponent, path string) (map[string]interface{}, error) {
	if section := component.PageComponentAuth0StructureSection; section != nil {
		config := section.Config

		children := make([]interface{}, 0, len(config.GetChildren()))
		for i, child := range config.GetChildren() {
			flattened, err := flattenSectionChild(child, fmt.Sprintf("%s.config.children[%d]", path, i))
			if err != nil {
				return nil, err
			}

			children = append(children, flattened)
		}

		return flattenComponentType("structure_section", map[string]interface{}{
			"variant":     string(config.GetVariant()),
			"title":       config.GetTitle(),
			"description": config.GetDescription(),
			"child":       children,
		}), nil
	}

	return flattenSectionChild(&management.PageComponentNoChildren{
		Type:                                             component.Type,
		PageComponentAuth0Form:                           component.PageComponentAuth0Form,
		PageComponentAuth0TypographyHeading:              component.PageComponentAuth0TypographyHeading,
		PageComponentAuth0MediaImage:                     component.PageComponentAuth0MediaImage,
		PageComponentAuth0TypographyRichText:             component.PageComponentAuth0TypographyRichText,
		PageComponentAuth0StructureSeparator:             component.PageComponentAuth0StructureSeparator,
		PageComponentAuth0MyOrganizationDomainTable:      component.PageComponentAuth0MyOrganizationDomainTable,
		PageComponentAuth0MyOrganizationMemberManagement: component.PageComponentAuth0MyOrganizationMemberManagement,
		PageComponentAuth0MyOrganizationDetailsEdit:      component.PageComponentAuth0MyOrganizationDetailsEdit,
		PageComponentAuth0MyAccountPasskeyManagement:     component.PageComponentAuth0MyAccountPasskeyManagement,
		PageComponentAuth0MyAccountMfaManagement:         component.PageComponentAuth0MyAccountMfaManagement,
	}, path)
}

func flattenSectionChild(component *management.PageComponentNoChildren, path string) (map[string]interface{}, error) {
	switch {
	case component.PageComponentAuth0Form != nil:
		config := component.PageComponentAuth0Form.Config
		return flattenComponentType("form", map[string]interface{}{
			"form_id":            config.GetFormID(),
			"completion_message": config.GetCompletionMessage(),
		}), nil
	case component.PageComponentAuth0TypographyHeading != nil:
		config := component.PageComponentAuth0TypographyHeading.Config
		return flattenComponentType("typography_heading", map[string]interface{}{
			"title":       config.GetTitle(),
			"description": config.GetDescription(),
		}), nil
	case component.PageComponentAuth0TypographyRichText != nil:
		config := component.PageComponentAuth0TypographyRichText.Config
		return flattenComponentType("typography_rich_text", map[string]interface{}{
			"content": config.GetContent(),
		}), nil
	case component.PageComponentAuth0MediaImage != nil:
		config := component.PageComponentAuth0MediaImage.Config
		return flattenComponentType("media_image", map[string]interface{}{
			"src":      config.GetSrc(),
			"position": string(config.GetPosition()),
			"height":   config.GetHeight(),
		}), nil
	case component.PageComponentAuth0StructureSeparator != nil:
		config := component.PageComponentAuth0StructureSeparator.Config
		return flattenComponentType("structure_separator", map[string]interface{}{
			"variant": string(config.GetVariant()),
			"text":    config.GetText(),
		}), nil
	case component.PageComponentAuth0MyOrganizationDomainTable != nil:
		return flattenComponentType("my_organization_domain_table", map[string]interface{}{}), nil
	case component.PageComponentAuth0MyOrganizationMemberManagement != nil:
		return flattenComponentType("my_organization_member_management", map[string]interface{}{}), nil
	case component.PageComponentAuth0MyOrganizationDetailsEdit != nil:
		return flattenComponentType("my_organization_details_edit", map[string]interface{}{}), nil
	case component.PageComponentAuth0MyAccountPasskeyManagement != nil:
		return flattenComponentType("my_account_passkey_management", map[string]interface{}{}), nil
	case component.PageComponentAuth0MyAccountMfaManagement != nil:
		return flattenComponentType("my_account_mfa_management", map[string]interface{}{}), nil
	}

	return nil, unsupportedComponentError(path, component.Type)
}

// flattenComponentType returns a component entry that sets only the given component type block.
func flattenComponentType(name string, fields map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		name: []interface{}{fields},
	}
}

// unsupportedComponentError is returned for a component type that the SDK or the schema does not know yet.
func unsupportedComponentError(path, componentType string) error {
	return fmt.Errorf("a component at `%s` of type `%s` that this provider version does not support", path, componentType)
}
