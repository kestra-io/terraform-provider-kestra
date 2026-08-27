package provider_v2_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
	"github.com/kestra-io/terraform-provider-kestra/internal/provider"
	"github.com/kestra-io/terraform-provider-kestra/internal/provider_v2"
)

// TestMuxServesWorkerGroupFromFrameworkProvider pins the provider that serves
// the worker group resources. `kestra_worker_group` moved from the SDK provider
// to the framework one: registering it on both fails the mux outright, and
// dropping it from the framework provider would leave the `worker_group` blocks
// of `kestra_namespace` and `kestra_tenant` referencing a resource that no
// longer exists.
func TestMuxServesWorkerGroupFromFrameworkProvider(t *testing.T) {
	ctx := context.Background()

	mux, err := tf5muxserver.NewMuxServer(ctx, []func() tfprotov5.ProviderServer{
		providerserver.NewProtocol5(provider_v2.New("test")()),
		provider.New("test", nil)().GRPCProvider,
	}...)
	if err != nil {
		t.Fatalf("unexpected mux server error: %v", err)
	}

	resp, err := mux.ProviderServer().GetProviderSchema(ctx, &tfprotov5.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatalf("unexpected provider schema error: %v", err)
	}
	for _, d := range resp.Diagnostics {
		if d.Severity == tfprotov5.DiagnosticSeverityError {
			t.Fatalf("unexpected provider schema diagnostic: %s: %s", d.Summary, d.Detail)
		}
	}

	for _, name := range []string{"kestra_worker_group", "kestra_worker_queue"} {
		if _, ok := resp.ResourceSchemas[name]; !ok {
			t.Errorf("expected the mux server to serve the %q resource", name)
		}
		if _, ok := resp.DataSourceSchemas[name]; !ok {
			t.Errorf("expected the mux server to serve the %q data source", name)
		}
	}

	// The framework implementation is the one being served: the SDK schema held
	// the worker group identifier in `key` and knew nothing of subscriptions.
	workerGroup, ok := resp.ResourceSchemas["kestra_worker_group"]
	if !ok {
		t.Fatal("expected a kestra_worker_group resource schema")
	}
	attributes := make(map[string]bool, len(workerGroup.Block.Attributes))
	for _, attribute := range workerGroup.Block.Attributes {
		attributes[attribute.Name] = true
	}
	if !attributes["group_id"] || attributes["key"] {
		t.Error("expected kestra_worker_group to expose group_id and no longer expose key")
	}
	var hasSubscriptions bool
	for _, block := range workerGroup.Block.BlockTypes {
		if block.TypeName == "subscriptions" {
			hasSubscriptions = true
		}
	}
	if !hasSubscriptions {
		t.Error("expected kestra_worker_group to expose a subscriptions block")
	}
	if workerGroup.Version != 1 {
		t.Errorf("expected kestra_worker_group schema version 1 for the state upgrader, got %d", workerGroup.Version)
	}
}
