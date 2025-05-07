package pingdom

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"pingdom": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProviderConfigure(t *testing.T) {
	var expectedToken string
	var isAccTestEnabled bool

	if v := os.Getenv("TF_ACC"); v != "" {
		isAccTestEnabled = true
	}

	if v := os.Getenv("PINGDOM_API_TOKEN"); v != "" {
		expectedToken = v
	} else {
		expectedToken = "foo"
	}

	raw := map[string]interface{}{
		"api_token": expectedToken,
	}

	// Previously there is only one client, which is the Pingdom client. It does not require obtaining any kind of
	// token during its initialization process, thus it will not verify whether the token provided is valid or not.
	// However, the case is different for the Solarwinds client because it will not initialize successfully unless
	// there are real user credentials provided. In this case, we need to skip the init process to avoid any test
	// errors if the credentials are not provided.
	if isAccTestEnabled {
		rp := Provider()
		err := rp.Configure(context.Background(), terraform.NewResourceConfigRaw(raw))
		if err != nil {
			t.Fatal(err)
		}

		pingdomClient := rp.Meta().(*Clients).Pingdom
		if pingdomClient.APIToken != expectedToken {
			t.Fatalf("bad: %#v", pingdomClient)
		}
	}
}

// New test function for API Token Only authentication
func TestProviderConfigureWithAPITokenOnly(t *testing.T) {
	var expectedTokenOnly string
	var isAccTestEnabled bool

	if v := os.Getenv("TF_ACC"); v != "" {
		isAccTestEnabled = true
	}

	if v := os.Getenv("PINGDOM_API_TOKEN_ONLY"); v != "" {
		expectedTokenOnly = v
	} else {
		expectedTokenOnly = "bar"
	}

	raw := map[string]interface{}{
		"api_token_only": expectedTokenOnly,
	}

	if isAccTestEnabled {
		rp := Provider()
		err := rp.Configure(context.Background(), terraform.NewResourceConfigRaw(raw))
		if err != nil {
			t.Fatal(err)
		}

		pingdomClient := rp.Meta().(*Clients).Pingdom
		if pingdomClient.APITokenOnly != expectedTokenOnly {
			t.Fatalf("bad APITokenOnly value: %#v", pingdomClient)
		}
	}
}

// Test both authentication methods together
func TestProviderConfigureWithBothTokens(t *testing.T) {
	var expectedToken, expectedTokenOnly string
	var isAccTestEnabled bool

	if v := os.Getenv("TF_ACC"); v != "" {
		isAccTestEnabled = true
	}

	if v := os.Getenv("PINGDOM_API_TOKEN"); v != "" {
		expectedToken = v
	} else {
		expectedToken = "foo"
	}

	if v := os.Getenv("PINGDOM_API_TOKEN_ONLY"); v != "" {
		expectedTokenOnly = v
	} else {
		expectedTokenOnly = "bar"
	}

	raw := map[string]interface{}{
		"api_token":      expectedToken,
		"api_token_only": expectedTokenOnly,
	}

	if isAccTestEnabled {
		rp := Provider()
		err := rp.Configure(context.Background(), terraform.NewResourceConfigRaw(raw))
		if err != nil {
			t.Fatal(err)
		}

		pingdomClient := rp.Meta().(*Clients).Pingdom
		if pingdomClient.APIToken != expectedToken {
			t.Fatalf("bad APIToken value: %#v", pingdomClient)
		}
		if pingdomClient.APITokenOnly != expectedTokenOnly {
			t.Fatalf("bad APITokenOnly value: %#v", pingdomClient)
		}
	}
}

func testAccPreCheck(t *testing.T) {
	// API token is required for all tests
	if v := os.Getenv("PINGDOM_API_TOKEN"); v == "" {
		t.Fatal("PINGDOM_API_TOKEN environment variable must be set for acceptance tests")
	}
	
	// API token only is optional but recommended for testing both authentication methods
	if v := os.Getenv("PINGDOM_API_TOKEN_ONLY"); v == "" {
		t.Log("PINGDOM_API_TOKEN_ONLY environment variable is not set. Some tests may be skipped.")
	}

	if v := os.Getenv("SOLARWINDS_USER"); v == "" {
		t.Fatal("SOLARWINDS_USER environment variable must be set for acceptance tests")
	}

	if v := os.Getenv("SOLARWINDS_PASSWD"); v == "" {
		t.Fatal("SOLARWINDS_PASSWD environment variable must be set for acceptance tests")
	}

	if v := os.Getenv("SOLARWINDS_ORG_ID"); v == "" {
		t.Fatal("SOLARWINDS_ORG_ID environment variable must be set for acceptance tests")
	}
}

func testAccCheckPingdomResourceID(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Can't find: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("ID not set: %s", name)
		}

		return nil
	}
}
