package dofunc

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Define the resource schema
func resourceNamespace() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNamespaceCreate,
		ReadContext:   resourceNamespaceRead,
		UpdateContext: resourceNamespaceUpdate,
		DeleteContext: resourceNamespaceDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Namespace name",
			},
			"label": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Label for the namespace",
			},
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "nyc1",
				Description: "Region for the namespace",
			},

		},
	}
}

// Create function to handle namespace creation
func resourceNamespaceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	label := d.Get("label").(string)
	region := d.Get("region").(string)

	// Construct the doctl command to create the namespace
	cmd := exec.Command("doctl", "serverless", "namespaces", "create", "--label", label, "--region", region)

	// Capture the output of the command for debugging
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	// Run the command and handle errors
	err := cmd.Run()
	if err != nil {
		return diag.Errorf("Failed to create serverless namespace: %s", out.String())
	}

	// Print the output for debugging (or use it to capture namespace details)
	fmt.Println("doctl output:", out.String())


	d.SetId(label)  // Set the ID to the namespace label
	return nil
}

// Read function to retrieve namespace details (optional)
func resourceNamespaceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Implementation for reading namespace details (if needed)
	return nil
}

// Update function (if applicable)
func resourceNamespaceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Get the new label and region from the resource data
	newLabel := d.Get("label").(string)
	region := d.Get("region").(string)

	// Get the previous label (before the update) and the new label (after the update)
	oldLabel, _ := d.GetChange("label")

	// If the label is the same, return without doing anything
	if oldLabel == newLabel {
		return nil
	}

	// Step 1: Mark the resource as destroyed in Terraform state (remove old ID)
	d.SetId("") // This marks the resource as destroyed

	// Step 2: Delete the existing namespace using the old label
	cmdDelete := exec.Command("doctl", "serverless", "namespaces", "delete", oldLabel.(string), "--force")

	// Capture the output of the delete command for debugging
	var outDelete bytes.Buffer
	cmdDelete.Stdout = &outDelete
	cmdDelete.Stderr = &outDelete

	// Run the delete command and handle errors
	err := cmdDelete.Run()
	if err != nil {
		return diag.Errorf("Failed to delete serverless namespace with label '%s': %s", oldLabel, outDelete.String())
	}

	// Print the output for debugging
	fmt.Println("doctl delete output:", outDelete.String())

	// Step 3: Create the new namespace with the updated label and region
	cmdCreate := exec.Command("doctl", "serverless", "namespaces", "create", "--label", newLabel, "--region", region)

	// Capture the output of the create command for debugging
	var outCreate bytes.Buffer
	cmdCreate.Stdout = &outCreate
	cmdCreate.Stderr = &outCreate

	// Run the create command and handle errors
	err = cmdCreate.Run()
	if err != nil {
		return diag.Errorf("Failed to create serverless namespace with label '%s': %s", newLabel, outCreate.String())
	}

	// Print the output for debugging
	fmt.Println("doctl create output:", outCreate.String())

	// Step 4: Set the new ID after creation (mark as created)
	d.SetId(newLabel)

	// Return nil to indicate the update was successful
	return nil
}





// Delete function to remove the namespace (optional)
func resourceNamespaceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	label := d.Get("label").(string)

	// Construct the doctl command to delete the namespace
	cmd := exec.Command("doctl", "serverless", "namespaces", "delete", label, "--force")

	// Capture the output of the command for debugging
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	// Run the command and handle errors
	err := cmd.Run()
	if err != nil {
		return diag.Errorf("Failed to delete serverless namespace: %s", out.String())
	}

	// Print the output for debugging
	fmt.Println("doctl output:", out.String())

	// Remove the resource ID, as the namespace is deleted
	d.SetId("")  // Remove the ID of the resource
	return nil
}
