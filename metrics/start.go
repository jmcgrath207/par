package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	promMetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

var DNSQueryCount *prometheus.CounterVec

func Start() {
	DNSQueryCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dns_queries",
			Help: "Returns the number of DNS queries made to the DNS server",
		},
		[]string{
			"type",
			"domainName",
			"ips",
			"clientIP",
		},
	)

	promMetrics.Registry.MustRegister(DNSQueryCount)
}
