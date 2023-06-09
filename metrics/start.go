package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var DNSQueryCount *prometheus.CounterVec

func Start() {
	DNSQueryCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dns_query_count",
			Help: "Returns the number of DNS queries made to the DNS server",
		},
		[]string{
			"domainName",
			"ips",
			"clientIP",
		},
	)

	metrics.Registry.MustRegister(DNSQueryCount)
}
