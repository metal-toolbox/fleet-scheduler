package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	ConditionorcErrorCounter *prometheus.CounterVec
	FleetdbErrorCounter      *prometheus.CounterVec

	InventoryCounter *prometheus.CounterVec
)

func init() {
	ConditionorcErrorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "fleet-scheduler",
			Subsystem: "conditionorc",
			Name:      "fleet_scheduler_conditionorc_errors",
			Help:      "a count of all errors attempting to reach conditionorc",
		}, []string{
			"errors",
		},
	)

	FleetdbErrorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "fleet-scheduler",
			Subsystem: "fleetdb",
			Name:      "fleet_scheduler_fleetdb_errors",
			Help:      "a count of all errors attempting to reach fleetdb",
		}, []string{
			"errors",
		},
	)

	InventoryCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "fleet-scheduler",
			Subsystem: "core",
			Name:      "fleet_scheduler_inventory_count",
			Help:      "a count of all errors attempting to reach fleet-scheduler dependencies",
		}, []string{},
	)
}

func AddCustomMetrics(pusher *Pusher) {
	pusher.AddCollector(ConditionorcErrorCounter)
	pusher.AddCollector(FleetdbErrorCounter)
	pusher.AddCollector(InventoryCounter)
}
