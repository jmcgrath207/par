package dns

import (
	"context"
	"fmt"
	dnsv1alpha1 "github.com/jmcgrath207/par/apis/dns/v1alpha1"
	"github.com/jmcgrath207/par/metrics"
	"github.com/jmcgrath207/par/storage"
	"github.com/miekg/dns"
	"net"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
)

func Start() {
	server := &dns.Server{Addr: ":9000", Net: "udp"}
	log.FromContext(context.Background()).Info("Starting DNS server", "port", "9000")
	server.Handler = dns.HandlerFunc(handleDNSRequest)
	<-storage.DNSReady
	err := server.ListenAndServe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server: %s\n", err.Error())
		panic(err)
	}
	log.FromContext(context.Background()).Info("DNS server running", "port", "9000")
}

func handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	var ips []net.IP
	m := new(dns.Msg)
	m.SetReply(r)
	if len(r.Question) == 0 {
		m.SetRcode(r, dns.RcodeServerFailure)
		w.WriteMsg(m)
		return
	}

	ipString := w.RemoteAddr().String()
	host, _, _ := net.SplitHostPort(ipString)
	clientIP := net.ParseIP(host)

	q := r.Question[0]
	if q.Qtype == dns.TypeA {
		ips, err := lookupIP(q.Name, clientIP.String())
		if err == nil {
			for _, ip := range ips {
				if ip.To4() == nil {
					continue
				}
				a := &dns.A{
					Hdr: dns.RR_Header{
						Name:   q.Name,
						Rrtype: dns.TypeA,
						Class:  dns.ClassINET,
						Ttl:    60,
					},
					A: ip,
				}
				m.Answer = append(m.Answer, a)
			}
		}
	}
	m.SetRcode(r, dns.RcodeSuccess)
	err := w.WriteMsg(m)
	if err != nil {
		panic(err)
	}
	// TODO: Report Slice not working
	reportSlice := []string{}
	for _, ip := range ips {
		reportSlice = append(reportSlice, ip.String())
	}
	metrics.DNSQueryCount.WithLabelValues(
		q.Name,
		strings.Join(reportSlice, " "),
		clientIP.String()).Inc()
	return
}

func lookupIP(domainName string, clientIP string) ([]net.IP, error) {
	var ips []net.IP

	// force traffic to go through proxy by return proxy address
	proxyIP, ok := storage.ToProxySourceHostMap[clientIP]
	if ok {
		return append(ips, proxyIP), nil
	}
	id, okId := storage.ClientId[clientIP]
	if okId {
		val, okRecord := storage.GetRecord("A", domainName+id)
		if okRecord {
			aRecord := val.(dnsv1alpha1.ARecordsSpec)
			for _, ip := range aRecord.IPAddresses {
				ips = append(ips, net.ParseIP(ip))
			}
			return ips, nil
		}

	}

	ips, err := net.LookupIP(domainName)
	if err != nil {
		return nil, err
	}

	return ips, nil
}
