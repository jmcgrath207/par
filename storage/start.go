package storage

import (
	"github.com/jmcgrath207/par/dns/types"
	"net"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	recordMap     = map[string]map[string]types.Record{}
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

func SetRecord(recordType string, record types.Record) {
	recordMap[record.HostName+"."][recordType] = record
}

func GetRecord(recordType string, hostname string) (types.Record, bool) {
	_, ok := recordMap[hostname]
	if ok {
		val, ok := recordMap[hostname][recordType]
		if ok {
			return val, true
		}

	}
	return types.Record{}, false
}
