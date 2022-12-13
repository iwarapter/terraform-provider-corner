package framework

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-corner/internal/backend"
)

var (
	_ resource.Resource                = &resourceUser{}
	_ resource.ResourceWithConfigure   = &resourceUser{}
	_ resource.ResourceWithImportState = &resourceUser{}
)

func NewUserResource() resource.Resource {
	return &resourceUser{}
}

type resourceUser struct {
	client *backend.Client
}

func (r *resourceUser) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *resourceUser) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"email": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"age": schema.NumberAttribute{
				Required: true,
			},
			// included only for compatibility with SDKv2 test framework
			"id": schema.StringAttribute{
				Optional: true,
			},
			"date_joined": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"language": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"oidc_policy": schema.SingleNestedAttribute{
				Description: `Open ID Connect Policy settings.  This is included in the message only when OIDC is enabled.`,
				Optional:    true,
				Computed:    true,
				Attributes:  singleClientOIDCPolicy(),
				PlanModifiers: []planmodifier.Object{
					DefaultObject(map[string]attr.Type{
						"grant_access_session_revocation_api":         types.BoolType,
						"grant_access_session_session_management_api": types.BoolType,
						"pairwise_identifier_user_type":               types.BoolType,
						"ping_access_logout_capable":                  types.BoolType,
						"id_token_content_encryption_algorithm":       types.StringType,
						"id_token_encryption_algorithm":               types.StringType,
						"id_token_signing_algorithm":                  types.StringType,
						"policy_group":                                types.StringType,
						"sector_identifier_uri":                       types.StringType,
						"logout_uris":                                 types.ListType{ElemType: types.StringType},
					}, map[string]attr.Value{
						"grant_access_session_revocation_api":         types.BoolValue(false),
						"grant_access_session_session_management_api": types.BoolValue(false),
						"pairwise_identifier_user_type":               types.BoolValue(false),
						"ping_access_logout_capable":                  types.BoolValue(false),
						"id_token_content_encryption_algorithm":       types.StringNull(),
						"id_token_encryption_algorithm":               types.StringNull(),
						"id_token_signing_algorithm":                  types.StringNull(),
						"policy_group":                                types.StringNull(),
						"sector_identifier_uri":                       types.StringNull(),
						"logout_uris":                                 types.ListNull(types.StringType),
					}),
				},
			},
		},
	}
}

func singleClientOIDCPolicy() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"grant_access_session_revocation_api": schema.BoolAttribute{
			Optional: true,
			PlanModifiers: []planmodifier.Bool{
				DefaultBool(false),
			},
		},
		"grant_access_session_session_management_api": schema.BoolAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.Bool{
				DefaultBool(false),
			},
		},
		"id_token_content_encryption_algorithm": schema.StringAttribute{Optional: true},
		"id_token_encryption_algorithm":         schema.StringAttribute{Optional: true},
		"id_token_signing_algorithm": schema.StringAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.String{
				DefaultString("RS256"),
			},
		},
		"logout_uris": schema.ListAttribute{Optional: true, ElementType: types.StringType},
		"pairwise_identifier_user_type": schema.BoolAttribute{
			Optional: true,
			Computed: true,
			PlanModifiers: []planmodifier.Bool{
				DefaultBool(false),
			},
		},
		"ping_access_logout_capable": schema.BoolAttribute{Optional: true},
		"policy_group":               schema.StringAttribute{Optional: true},
		"sector_identifier_uri":      schema.StringAttribute{Optional: true},
	}
}

func (r *resourceUser) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*backend.Client)
}

type user struct {
	Email      string                `tfsdk:"email"`
	Name       string                `tfsdk:"name"`
	Age        int                   `tfsdk:"age"`
	Id         string                `tfsdk:"id"`
	DateJoined types.String          `tfsdk:"date_joined"`
	Language   types.String          `tfsdk:"language"`
	OidcPolicy *ClientOIDCPolicyData `tfsdk:"oidc_policy"`
}

type ClientOIDCPolicyData struct {
	GrantAccessSessionRevocationApi        types.Bool     `tfsdk:"grant_access_session_revocation_api"`
	GrantAccessSessionSessionManagementApi types.Bool     `tfsdk:"grant_access_session_session_management_api"`
	IdTokenContentEncryptionAlgorithm      types.String   `tfsdk:"id_token_content_encryption_algorithm"`
	IdTokenEncryptionAlgorithm             types.String   `tfsdk:"id_token_encryption_algorithm"`
	IdTokenSigningAlgorithm                types.String   `tfsdk:"id_token_signing_algorithm"`
	LogoutUris                             []types.String `tfsdk:"logout_uris"`
	PairwiseIdentifierUserType             types.Bool     `tfsdk:"pairwise_identifier_user_type"`
	PingAccessLogoutCapable                types.Bool     `tfsdk:"ping_access_logout_capable"`
	PolicyGroup                            types.String   `tfsdk:"policy_group"`
	SectorIdentifierUri                    types.String   `tfsdk:"sector_identifier_uri"`
}

func (r resourceUser) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan user
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	newUser := &backend.User{
		Email: plan.Email,
		Name:  plan.Name,
		Age:   plan.Age,
	}
	if !plan.Language.IsUnknown() {
		newUser.Language = plan.Language.ValueString()
	}

	err := r.client.CreateUser(newUser)
	if err != nil {
		resp.Diagnostics.AddError("Error creating user", err.Error())
		return
	}

	p, err := r.client.ReadUser(newUser.Email)
	if err != nil {
		resp.Diagnostics.AddError("Error reading user", err.Error())
		return
	}

	if p == nil {
		resp.Diagnostics.AddError("Error reading user", "could not find user after it was created")
		return
	}
	plan.DateJoined = types.StringValue(p.DateJoined)
	plan.Language = types.StringValue(p.Language)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceUser) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state user
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	p, err := r.client.ReadUser(state.Email)
	if err != nil {
		resp.Diagnostics.AddError("Error reading user", err.Error())
		return
	}

	if p == nil {
		return
	}

	state.Name = p.Name
	state.Age = p.Age
	state.DateJoined = types.StringValue(p.DateJoined)
	state.Language = types.StringValue(p.Language)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceUser) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan user
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	newUser := &backend.User{
		Email: plan.Email,
		Name:  plan.Name,
		Age:   plan.Age,
	}
	if !plan.Language.IsUnknown() {
		newUser.Language = plan.Language.ValueString()
	}

	err := r.client.UpdateUser(newUser)
	if err != nil {
		resp.Diagnostics.AddError("Error updating user", err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceUser) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state user
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	userToDelete := &backend.User{
		Email: state.Email,
		Name:  state.Name,
		Age:   state.Age,
	}

	err := r.client.DeleteUser(userToDelete)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting user", err.Error())
		return
	}
}

func (r resourceUser) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
