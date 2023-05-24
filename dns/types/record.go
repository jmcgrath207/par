package types

import (
	dnsv1alpha1 "github.com/jmcgrath207/par/apis/dns/v1alpha1"
)

type Record struct {
	dnsv1alpha1.ARecordsSpec
	RecordType string
}
