package dns

import (
	"context"
	"fmt"
	"github.com/jmcgrath207/par/storage"
	"github.com/miekg/dns"
	"net"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func Start() {
	server := &dns.Server{Addr: ":9000", Net: "udp"}
	<-storage.ProxyReady
	log.FromContext(context.Background()).Info("Starting DNS server", "port", "9000")
	server.Handler = dns.HandlerFunc(handleDNSRequest)
	err := server.ListenAndServe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server: %s\n", err.Error())
		panic(err)
	}
}

func handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
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
		ips, err := lookupIP(q.Name, clientIP)
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
	return
}

func lookupIP(domainName string, clientIP net.IP) ([]net.IP, error) {
	var ipSlice []net.IP

	// force traffic to go through proxy
	proxyIP, ok := storage.SourceHostMap[clientIP.String()]
	if ok {
		log.FromContext(context.Background()).Info("Found client IP in storage, returning proxy IP",
			"domainName", domainName, "ips", proxyIP, "clientIP", clientIP)
		return append(ipSlice, proxyIP), nil
	}

	val, ok := storage.GetRecord("A", domainName)
	if ok {
		log.FromContext(context.Background()).Info("Found A record in storage, returning ip", "domainName", domainName, "ips", val, "clientIP", clientIP)
		return append(ipSlice, net.ParseIP(val.IPAddress)), nil
	}

	ips, err := net.LookupIP(domainName)
	if err != nil {
		return nil, err
	}
	log.FromContext(context.Background()).Info("Return A record in found in Cluster DNS, returning ip", "domainName", domainName, "ips", ips, "clientIP", clientIP)

	return ips, nil
}
