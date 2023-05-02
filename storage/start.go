package storage

import (
	dnsv1 "github.com/jmcgrath207/par/apis/dns/v1"
	"net"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	RecordMap     = map[string]map[string]string{}
	ArecordMap    = map[string]dnsv1.Arecord{}
	SourceHostMap = map[string]net.IP{}
	ProxyReady    chan bool
	ClientK8s     client.Client
	Mgr           ctrl.Manager
)

func Start(mgr ctrl.Manager) {
	ProxyReady = make(chan bool)
	ClientK8s = mgr.GetClient()
	Mgr = mgr
}

func SetRecord(recordType string, hostname string, data string) {
	hostname = hostname + "."

	// Initialize the inner map if it does not exist.
	if RecordMap[hostname] == nil {
		RecordMap[hostname] = map[string]string{}
	}

	// Set the key-value pair in the inner map.
	RecordMap[hostname][recordType] = data
}

func GetRecord(recordType string, hostname string) (string, bool) {
	_, ok := RecordMap[hostname]
	if ok {
		val, ok := RecordMap[hostname][recordType]
		if ok {
			return val, true
		}

	}
	return "", false
}
