package jks_trust_store

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"jks_trust_store": resourceTrustStore(),
		},
		DataSourcesMap: map[string]*schema.Resource{},
	}
}
