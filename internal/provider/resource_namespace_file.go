package provider

import (
	"bytes"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
)

func resourceNamespaceFile() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Kestra Namespace File.",

		CreateContext: resourceNamespaceFileCreate,
		ReadContext:   resourceNamespaceFileRead,
		DeleteContext: resourceNamespaceFileDelete,
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description: "The tenant id.",
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
			},
			"namespace": {
				Description: "The namespace of the namespace file resource.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"filename": {
				Description: "The path to the namespace file that will be created.\n" +
					"Missing parent directories will be created.\n" +
					"If the file already exists, it will be overridden with the given content.",
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"content": {
				Description: "Content to store in the file, expected to be a UTF-8 encoded string.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceNamespaceFileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	namespace := d.Get("namespace").(string)
	filename := d.Get("filename").(string)
	content := d.Get("content").(string)
	tenantId := c.TenantId

	url := c.Url + fmt.Sprintf("%s/namespaces/%s/files?path=%s", apiRoot(tenantId), namespace, filename)

	req, err := addFilePartRequest(ctx, url, content)
	if err != nil {
		return diag.FromErr(err)
	}

	_, reqErr := c.rawRequest("POST", url, req)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	d.SetId(fmt.Sprintf("%s/%s", namespace, filename))
	if *c.TenantId != "" {
		if err := d.Set("tenant_id", c.TenantId); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set("namespace", namespace); err != nil {
		return diag.FromErr(err)
	}

    filename = d.Get("filename").(string)

    // Check if filename starts with a "/"
    if !strings.HasPrefix(filename, "/") {
        filename = "/" + filename
    }

    if err := d.Set("filename", filename); err != nil {
     return diag.FromErr(err)
    }

	if err := d.Set("content", content); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceNamespaceFileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	namespace, filename := namespaceFileConvertId(d.Id())
	tenantId := c.TenantId

	url := c.Url + fmt.Sprintf("%s/namespaces/%s/files?path=%s", apiRoot(tenantId), namespace, filename)

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf(url), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	_, body, reqErr := c.rawResponseRequest("GET", req)
	if reqErr != nil {
		if reqErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return diags
		}
		return diag.FromErr(reqErr.Err)
	}

	d.SetId(fmt.Sprintf("%s/%s", namespace, filename))
	if *c.TenantId != "" {
		if err := d.Set("tenant_id", c.TenantId); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set("namespace", namespace); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("filename", filename); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("content", string(body)); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceNamespaceFileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	namespace, filename := namespaceFileConvertId(d.Id())
	tenantId := c.TenantId

	url := fmt.Sprintf("%s/namespaces/%s/files?path=%s", apiRoot(tenantId), namespace, filename)

	_, reqErr := c.request("DELETE", url, nil)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	d.SetId("")

	return diags
}

func addFilePartRequest(ctx context.Context, url, content string) (*http.Request, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	var r io.Reader = strings.NewReader(content)

	fw, err := w.CreateFormFile("fileContent", "file.txt")
	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(fw, r); err != nil {
		return nil, err
	}

	w.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf(url), &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req, nil
}

func namespaceFileConvertId(id string) (string, string) {
	splits := strings.Split(id, "/")

	return splits[0], strings.Join(splits[1:], "/")
}
