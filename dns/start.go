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
	q := r.Question[0]
	if q.Qtype == dns.TypeA {
		ip, err := lookupIP(q.Name)
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

func lookupIP(host string) (net.IP, error) {
	storage.GetRecord("A", host)
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, err
	}
	return ips[0], nil
}
