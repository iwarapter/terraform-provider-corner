package provider1

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceUser1(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"tf6muxprovider": func() (tfprotov6.ProviderServer, error) {
				return tfsdk.NewProtocol6Server(New()), nil
			},
		},
		Steps: []resource.TestStep{
			{
				Config: configResourceUserBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("tf6muxprovider_user1.example", "age", "123"),
					resource.TestCheckResourceAttr("tf6muxprovider_user1.example", "email", "example@example.com"),
					resource.TestCheckResourceAttr("tf6muxprovider_user1.example", "id", "h"),
					resource.TestCheckResourceAttr("tf6muxprovider_user1.example", "language", "en"),
					resource.TestCheckResourceAttr("tf6muxprovider_user1.example", "name", "Example Name"),
				),
			},
		},
	})
}

const configResourceUserBasic = `
resource "tf6muxprovider_user1" "example" {
  age   = 123
  email = "example@example.com"
  id    = "h"
  name  = "Example Name"
}
`
