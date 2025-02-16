// Provider.go will have the resource server function calls.
package dofunc

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"golang.org/x/oauth2"
)

func tokenSource(token string) oauth2.TokenSource {
	return oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
}

func Provider() *schema.Provider {
    return &schema.Provider{
        Schema: map[string]*schema.Schema{
            "api_token": {
                Type:        schema.TypeString,
                Required:    true,
                Sensitive:   true,
                Description: "DigitalOcean API Token",
            },
        },
        ResourcesMap: map[string]*schema.Resource{
            "dofunc_function": resourceFunction(),
            "dofunc_namespace": resourceNamespace(),
        },
		ConfigureFunc: func(d *schema.ResourceData) (interface{}, error) {  
			apiToken := d.Get("api_token").(string)  
			return apiToken, nil  // ✅ Correct return type (error)
		},
    }
}
