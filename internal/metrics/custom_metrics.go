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
			Namespace: "fleet_scheduler",
			Subsystem: "conditionorc",
			Name:      "client_errors",
			Help:      "a count of all errors attempting to reach conditionorc",
		}, []string{},
	)

	FleetdbErrorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "fleet_scheduler",
			Subsystem: "fleetdb",
			Name:      "client_errors",
			Help:      "a count of all errors attempting to reach fleetdb",
		}, []string{},
	)

	InventoryCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "fleet_scheduler",
			Subsystem: "core",
			Name:      "inventory_count",
			Help:      "a count of all errors attempting to reach fleet-scheduler dependencies",
		}, []string{},
	)
}

func AddCustomMetrics(pusher *Pusher) {
	pusher.AddCollector(ConditionorcErrorCounter)
	pusher.AddCollector(FleetdbErrorCounter)
	pusher.AddCollector(InventoryCounter)
}
