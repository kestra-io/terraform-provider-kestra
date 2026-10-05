package provider_v2

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kestra-io/client-sdk/go-sdk/v2/kestra_api_client"
)

// Some backends answer 500 "tenantId cannot be null" on a tenant delete that did
// remove the tenant (#215). Delete may only trust that when a re-read says it is gone.
func TestTenantDeleteAfterServerError(t *testing.T) {
	tests := []struct {
		name         string
		deleteStatus int
		getStatus    int
		wantError    bool
	}{
		{"204 is a success", http.StatusNoContent, 0, false},
		{"404 is a success", http.StatusNotFound, 0, false},
		{"500 and tenant gone is a success", http.StatusInternalServerError, http.StatusNotFound, false},
		{"500 and tenant still there is an error", http.StatusInternalServerError, http.StatusOK, true},
		{"403 is an error and is not re-read", http.StatusForbidden, http.StatusNotFound, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodDelete {
					w.WriteHeader(tt.deleteStatus)
					return
				}
				w.WriteHeader(tt.getStatus)
			}))
			defer srv.Close()

			cfg := kestra_api_client.NewConfiguration()
			cfg.Servers = kestra_api_client.ServerConfigurations{{URL: srv.URL}}
			r := &tenantResource{providerData: ProviderData{Client: kestra_api_client.NewAPIClient(cfg)}}

			ctx := context.Background()
			state := currentTenantState(ctx, t)
			if diags := state.Set(ctx, &tenantModel{
				TenantId:             types.StringValue("t1"),
				StorageConfiguration: types.MapNull(types.StringType),
				SecretConfiguration:  types.MapNull(types.StringType),
			}); diags.HasError() {
				t.Fatalf("unexpected state diagnostics: %v", diags)
			}

			resp := resource.DeleteResponse{}
			r.Delete(ctx, resource.DeleteRequest{State: state}, &resp)

			if got := resp.Diagnostics.HasError(); got != tt.wantError {
				t.Errorf("error = %v, want %v (%v)", got, tt.wantError, resp.Diagnostics)
			}
		})
	}
}
