package types

import (
	dnsv1alpha1 "github.com/jmcgrath207/par/apis/dns/v1alpha1"
	"reflect"
)

type Records struct {
	dnsv1alpha1.Records
	namespaces  map[string]int
	RecordItems []Record
}
type RecordsList struct {
	dnsv1alpha1.RecordsList
}

func (r *Records) Set() {

	r.namespaces = make(map[string]int)
	r.RecordItems = make([]Record, 0)

	v := reflect.ValueOf(r)
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i).Name
		if field == "ManagerAddress" {
			continue
		}
		record := v.Field(i).Interface().(Record)
		record.RecordType = field
		r.RecordItems = append(r.RecordItems, record)
		a := reflect.ValueOf(record)
		b := a.Type()
		for x := 0; x < a.NumField(); x++ {
			field = b.Field(i).Name
			if field == "Namespaces" {
				r.namespaces[field] = 1
			}

		}
	}
}

func (r *Records) InNamespaces(namespace string) bool {
	_, ok := r.namespaces[namespace]
	if ok {
		return true
	}
	return false
}

func (r *Records) InLabels(namespace string) bool {
	return true
}
