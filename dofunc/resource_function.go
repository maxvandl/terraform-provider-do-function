package dofunc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"



	"archive/zip"




	"io"


	"os"
	"os/exec"
	"path/filepath"


	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Define the resource schema
func resourceFunction() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFunctionCreate,
		ReadContext:   resourceFunctionRead,
		UpdateContext: resourceFunctionUpdate, // ✅ Added Update function
		DeleteContext: resourceFunctionDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Function name",
			},
			"code": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Base64-encoded function code",
			},
			"runtime": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "nodejs",
				Description: "Runtime environment",
			},
			"url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Function URL",
			},
		},
	}
}

// Create function
func resourceFunctionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiToken := m.(string)
	name := d.Get("name").(string)
	runtime := d.Get("runtime").(string)
	gitZipURL := d.Get("git_zip_url").(string)

	// Folder for serverless function
	functionDir := "/tmp/example-project"

	// Step 1: Download ZIP from GitHub
	fmt.Println("Downloading function ZIP from:", gitZipURL)
	zipPath := "/tmp/function.zip"
	err := downloadFile(zipPath, gitZipURL)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to download ZIP: %w", err))
	}

	// Step 2: Extract ZIP
	fmt.Println("Extracting ZIP to:", functionDir)
	err = unzip(zipPath, functionDir)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to extract ZIP: %w", err))
	}

	// Step 3: Deploy using `doctl`
	fmt.Println("Deploying function with doctl...")
	err = deployFunction(functionDir)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to deploy function: %w", err))
	}

	// Step 4: Register function with DigitalOcean API
	fmt.Println("Registering function in DigitalOcean...")
	apiURL := "https://api.digitalocean.com/v2/functions/namespaces"
	reqBody, err := json.Marshal(map[string]string{
		"name":    name,
		"runtime": runtime,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return diag.FromErr(err)
	}
	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	// Check response for debugging
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return diag.FromErr(err)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error response: %s\n", string(body))
		return diag.Errorf("Failed to create function: %s", string(body))
	}

	fmt.Println("Function deployed successfully!")
	d.Set("url", "https://your-function-url.com")
	d.SetId(name)
	return nil
}

// downloadFile downloads a file from a given URL
func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// unzip extracts a ZIP file to a destination folder
func unzip(src string, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	err = os.MkdirAll(dest, 0755)
	if err != nil {
		return err
	}

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			err = os.MkdirAll(fpath, os.ModePerm)
			if err != nil {
				return err
			}
			continue
		}

		err = extractFile(f, fpath)
		if err != nil {
			return err
		}
	}
	return nil
}

// extractFile extracts a single file from ZIP
func extractFile(f *zip.File, dest string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, rc)
	return err
}

// deployFunction runs `doctl serverless deploy`
func deployFunction(dir string) error {
	cmd := exec.Command("doctl", "serverless", "deploy", dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
// Read function
func resourceFunctionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiToken := m.(string)
	name := d.Get("name").(string)

	apiURL := fmt.Sprintf("https://api.digitalocean.com/v2/functions/%s", name)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return diag.FromErr(err)
	}
	req.Header.Set("Authorization", "Bearer "+apiToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	}

	body, _ := ioutil.ReadAll(resp.Body)

	var responseData struct {
		URL string `json:"url"`
	}
	json.Unmarshal(body, &responseData)

	d.Set("url", responseData.URL)
	return nil
}

// Update function
func resourceFunctionUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiToken := m.(string)
	name := d.Get("name").(string)
	code := d.Get("code").(string)
	runtime := d.Get("runtime").(string)

	apiURL := fmt.Sprintf("https://api.digitalocean.com/v2/functions/%s", name)

	reqBody, _ := json.Marshal(map[string]string{
		"runtime": runtime,
		"code":    code,
	})

	req, err := http.NewRequest("PUT", apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return diag.FromErr(err)
	}
	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return diag.Errorf("Failed to update function: %s", resp.Status)
	}

	return resourceFunctionRead(ctx, d, m)
}

// Delete function
func resourceFunctionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	apiToken := m.(string)
	name := d.Get("name").(string)

	apiURL := fmt.Sprintf("https://api.digitalocean.com/v2/functions/%s", name)

	req, err := http.NewRequest("DELETE", apiURL, nil)
	if err != nil {
		return diag.FromErr(err)
	}
	req.Header.Set("Authorization", "Bearer "+apiToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return diag.Errorf("Failed to delete function: %s", resp.Status)
	}

	d.SetId("")
	return nil
}
