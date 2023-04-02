package dns

import (
	"fmt"
	"github.com/jmcgrath207/par/storage"
	"github.com/miekg/dns"
	"net"
	"os"
)

func Start() {
	server := &dns.Server{Addr: ":53", Net: "udp"}
	<-storage.AcquiredProxyServiceIP
	server.Handler = dns.HandlerFunc(handleDNSRequest)
	err := server.ListenAndServe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server: %s\n", err.Error())
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
	senderIP := net.ParseIP(host)

	q := r.Question[0]
	if q.Qtype == dns.TypeA {
		ip, err := lookupIP(q.Name, senderIP)
		if err == nil {
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
			m.SetRcode(r, dns.RcodeSuccess)
			w.WriteMsg(m)
			return
		}
	}
	m.SetRcode(r, dns.RcodeNameError)
	w.WriteMsg(m)
}

func lookupIP(domainName string, senderIP net.IP) (net.IP, error) {

	// force traffic to work to proxy
	proxyIP, ok := storage.SourceHostMap[senderIP.String()]
	if ok {
		return proxyIP, nil
	}

	val, ok := storage.GetRecord("A", domainName)
	if ok {
		return net.ParseIP(val), nil
	}

	ips, err := net.LookupIP(domainName)
	if err != nil {
		return nil, err
	}

	return ips[0], nil
}
