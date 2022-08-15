package jks_trust_store

import (
	"bufio"
	"bytes"
	"context"
	"crypto"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/pavel-v-chernykh/keystore-go/v4"
	"strings"
	"time"
)

func resourceTrustStore() *schema.Resource {
	return &schema.Resource{
		Description:   "JKS trust store generated from one or more PEM encoded certificates.",
		CreateContext: resourceTrustStoreCreate,
		ReadContext:   resourceTrustStoreRead,
		DeleteContext: resourceTrustStoreDelete,
		Schema: map[string]*schema.Schema{
			"certificates": {
				Description: "CA certificates or chains to include in generated trust store; in PEM format.",
				Type:        schema.TypeList,
				Required:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				MinItems: 1,
				ForceNew: true,
			},
			"password": {
				Description: "Password to secure trust store. Defaults to empty string.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				ForceNew:    true,
			},
			"jks": {
				Description: "JKS trust store data; base64 encoded.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceTrustStoreCreate(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	ks := keystore.New()

	chains := d.Get("certificates").([]interface{})
	if len(chains) == 0 {
		return diag.Errorf("Empty certificates")
	}

	for chainIdx, chain := range chains {
		rest := []byte(strings.TrimSpace(chain.(string)))
		certDerData := make([]byte, 0)
		certIdx := -1
		for {
			certIdx += 1

			var block *pem.Block = nil
			block, rest = pem.Decode(rest)
			if block == nil && len(rest) == 0 {
				// Done iterating PEM blocks
				break
			} else if block == nil {
				diags = append(diags, diag.Errorf("chain %d, certificate %d: failed to load PEM", chainIdx, certIdx)...)
			} else if (block.Type != "CERTIFICATE") && (block.Type != "EC PRIVATE KEY") {
				diags = append(diags, diag.Errorf("chain %d, certificate %d: expected CERTIFICATE but found %q", chainIdx, certIdx, block.Type)...)
			}

			certDerData = append(certDerData, block.Bytes...)
		}

		err := ks.SetTrustedCertificateEntry(
			fmt.Sprintf("%d", chainIdx),
			keystore.TrustedCertificateEntry{
				CreationTime: time.Now(),
				Certificate: keystore.Certificate{
					Type:    "X.509",
					Content: certDerData,
				},
			},
		)
		if err != nil {
			diags = append(diags, diag.Errorf("chain %d: %v", chainIdx, err)...)
		}
	}

	var jksBuffer bytes.Buffer
	jksWriter := bufio.NewWriter(&jksBuffer)

	err := ks.Store(jksWriter, []byte(d.Get("password").(string)))
	if err != nil {
		diags = append(diags, diag.Errorf("Failed to generate JKS: %v", err)...)
	}

	err = jksWriter.Flush()
	if err != nil {
		diags = append(diags, diag.Errorf("Failed to flush JKS: %v", err)...)
	}

	jksData := base64.StdEncoding.EncodeToString(jksBuffer.Bytes())

	idHash := crypto.SHA1.New()
	idHash.Write([]byte(jksData))

	id := hex.EncodeToString(idHash.Sum([]byte{}))
	d.SetId(id)

	if err = d.Set("jks", jksData); err != nil {
		diags = append(diags, diag.Errorf("Failed to save JKS: %v", err)...)
	}

	return diags
}

func resourceTrustStoreRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceTrustStoreCreate(ctx, d, m)
}

func resourceTrustStoreDelete(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	d.SetId("")

	return diags
}
