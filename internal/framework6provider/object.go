package framework

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ planmodifier.Object = objectDefaultModifier{}

func DefaultObject(elementType map[string]attr.Type, elements map[string]attr.Value) objectDefaultModifier {
	return objectDefaultModifier{ElementType: elementType, Elements: elements}
}

type objectDefaultModifier struct {
	ElementType map[string]attr.Type
	Elements    map[string]attr.Value
}

// Description returns a plain text description of the validator's behavior, suitable for a practitioner to understand its impact.
func (m objectDefaultModifier) Description(_ context.Context) string {
	return fmt.Sprintf("If value is not configured, defaults to %s", m.Elements)
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior, suitable for a practitioner to understand its impact.
func (m objectDefaultModifier) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("If value is not configured, defaults to `%s`", m.Elements)
}

// PlanModifyObject updates the planned value with the default if its not null
func (m objectDefaultModifier) PlanModifyObject(_ context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	// If the attribute configuration is not null, we are done here
	if !req.ConfigValue.IsNull() {
		return
	}

	// If the attribute plan is "known" and "not null", then a previous plan modifier in the sequence
	// has already been applied, and we don't want to interfere.
	if !req.PlanValue.IsUnknown() && !req.PlanValue.IsNull() {
		return
	}
	resp.PlanValue, resp.Diagnostics = types.ObjectValue(m.ElementType, m.Elements)
}
