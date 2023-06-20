package storage

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
)

var (
	recordMap    map[string]map[string]interface{}
	ClientK8s    client.Client
	Mgr          ctrl.Manager
	ClientId     = map[string]string{}
	DNSWaitGroup sync.WaitGroup
)

func Start(mgr ctrl.Manager) {
	recordMap = make(map[string]map[string]interface{})
	ClientK8s = mgr.GetClient()
	DNSWaitGroup.Add(1)
	Mgr = mgr
}

func SetRecord(recordType string, hostname string, record interface{}) {
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
