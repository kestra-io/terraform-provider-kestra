package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceDashboard(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDashboard("new"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("kestra_dashboard.new", "id"),
				),
			},
		},
	})
}

func testAccResourceDashboard(resourceId string) string {
	return fmt.Sprintf(
		`
        resource "kestra_dashboard" "%s" {
            source_code = <<-EOF
id: %s
title: Overview_test
charts:
  - id: executions_timeseries
    type: io.kestra.plugin.core.dashboard.chart.TimeSeries
    chartOptions:
      displayName: Executions
      description: Executions duration and count per date
      legend:
        enabled: true
      column: date
      colorByColumn: state
    data:
      type: io.kestra.plugin.core.dashboard.data.Executions
      columns:
        date:
          field: START_DATE
          displayName: Date
        state:
          field: STATE
        total:
          displayName: Executions
          agg: COUNT
          graphStyle: BARS
EOF
        }`,
		resourceId,
		resourceId)
}
