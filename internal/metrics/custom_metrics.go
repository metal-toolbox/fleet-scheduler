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
			Name:      "errors_total",
			Help:      "a count of all errors attempting to reach conditionorc",
		}, []string{
			"errors",
		},
	)

	FleetdbErrorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "fleet-scheduler",
			Subsystem: "fleetdb",
			Name:      "errors_total",
			Help:      "a count of all errors attempting to reach fleetdb",
		}, []string{
			"errors",
		},
	)

	InventoryCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "fleet-scheduler",
			Subsystem: "core",
			Name:      "errors_total",
			Help:      "a count of all errors attempting to reach fleet-scheduler dependencies",
		}, []string{},
	)
}

func AddCustomMetrics(pusher *Pusher) {
	pusher.AddCollector(ConditionorcErrorCounter)
	pusher.AddCollector(FleetdbErrorCounter)
	pusher.AddCollector(InventoryCounter)
}
