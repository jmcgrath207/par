package dns

import (
	"fmt"
	"github.com/miekg/dns"
	"net"
)

func Start() {
	// Define the DNS server address and port
	addr := ":53"

	// Create a new DNS server
	server := &dns.Server{Addr: addr, Net: "udp"}

	// Set the DNS server handler function
	server.Handler = dns.HandlerFunc(func(w dns.ResponseWriter, r *dns.Msg) {
		// Handle the DNS request
		fmt.Println(r.Question[0].Name)
		msg := new(dns.Msg)
		msg.SetReply(r)
		msg.Answer = []dns.RR{
			&dns.A{
				Hdr: dns.RR_Header{Name: r.Question[0].Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
				A:   net.ParseIP("127.0.0.1"),
			},
		}
		w.WriteMsg(msg)
	})

	// Start the DNS server
	fmt.Println("DNS server listening on", addr)
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("Error starting DNS server:", err)
	}
}
