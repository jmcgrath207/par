package storage

import (
	"net"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	recordMap     map[string]map[string]interface{}
	SourceHostMap = map[string]net.IP{}
	ProxyReady    chan bool
	ClientK8s     client.Client
	Mgr           ctrl.Manager
)

func Start(mgr ctrl.Manager) {
	recordMap = make(map[string]map[string]interface{})
	ProxyReady = make(chan bool)
	ClientK8s = mgr.GetClient()
	Mgr = mgr
}

func SetRecord(recordType string, hostname string, record interface{}) {
	hostname = hostname + "."
	if recordMap[hostname] == nil {
		recordMap[hostname] = make(map[string]interface{})
	}
	recordMap[hostname][recordType] = record
}

func GetRecord(recordType string, hostname string) (interface{}, bool) {
	_, ok := recordMap[hostname]
	if ok {
		val, ok := recordMap[hostname][recordType]
		if ok {
			return val, true
		}
	}
	return nil, false
}
