package dns

import (
	"log"
	"net"
	"time"

	"github.com/Dani-1004/Smart-DNS/dns-server/Modules/db"
	"github.com/miekg/dns"
)

// ForwardHandler berisi konfigurasi untuk meneruskan request
type ForwardHandler struct {
	UpstreamAddr string
	UDPClient    *dns.Client
	TCPClient    *dns.Client
	LandingIP    string
}

// NewForwardHandler membuat instance baru dari ForwardHandler
func NewForwardHandler(upstreamAddr string, landingIP string) *ForwardHandler {
	return &ForwardHandler{
		UpstreamAddr: upstreamAddr,
		LandingIP:    landingIP,
		// Inisialisasi client DNS dengan timeout yang lebih singkat untuk upstream
		UDPClient: &dns.Client{
			Timeout: 1 * time.Second,
			Net:     "udp",
		},
		TCPClient: &dns.Client{
			Timeout: 2 * time.Second,
			Net:     "tcp",
		},
	}
}

// ServeDNS mengimplementasikan interface dns.Handler.
// Fungsi ini menangani request DNS yang masuk, meneruskannya, dan mengirim kembali response.
func (h *ForwardHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	// 1. Tentukan jaringan ("udp" atau "tcp") berdasarkan koneksi client
	netType := w.LocalAddr().Network()

	var qName string
	var recordtype string

	if len(r.Question) > 0 {
		qName = r.Question[0].Name
		recordtype = dns.TypeToString[r.Question[0].Qtype]
	}

	// BLACKLIST CHECK: return blocked response immediately (fast path)
	if rec, blocked := db.CheckBlacklist(qName); blocked {
		b := new(dns.Msg)
		b.SetReply(r)

		ip := net.ParseIP(h.LandingIP) // Merefer ke IP Server yang lari ke Landing Page

		aRecord := &dns.A{
			Hdr: dns.RR_Header{
				Name:   qName,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    uint32(600),
			},
			A: ip,
		}
		b.Answer = append(b.Answer, aRecord)
		if err := w.WriteMsg(b); err != nil {
			log.Printf("Error sending blocked response: %v\n", err)
		}
		log.Printf("Blocked (blacklist) %s (action=%s)\n", qName, rec.Action)
		return
	}

	// Try to get from cache first (fastest path)
	if records, err := db.GetRecords(qName, recordtype); err == nil && len(records) > 0 {
		resp := new(dns.Msg)
		resp.SetReply(r)

		for _, record := range records {
			if record.RecordType == "A" {
				if ip := net.ParseIP(record.RecordValue); ip != nil {
					rr := &dns.A{
						Hdr: dns.RR_Header{
							Name:   qName,
							Rrtype: dns.TypeA,
							Class:  dns.ClassINET,
							Ttl:    uint32(record.TTL),
						},
						A: ip,
					}
					resp.Answer = append(resp.Answer, rr)
				}
			}
		}

		if len(resp.Answer) > 0 {
			if err := w.WriteMsg(resp); err != nil {
				log.Printf("Error sending cached response: %v\n", err)
			}
			return
		}
	}

	// No cache hit, query upstream
	var in *dns.Msg
	var err error

	// Select appropriate client based on network type
	client := h.UDPClient
	if netType == "tcp" {
		client = h.TCPClient
	}

	in, _, err = client.Exchange(r, h.UpstreamAddr)

	// Handle error from upstream query
	if err != nil {
		log.Printf("Error forwarding request to upstream %s: %v\n", h.UpstreamAddr, err)
		m := new(dns.Msg)
		m.SetRcode(r, dns.RcodeServerFailure)
		w.WriteMsg(m)
		return
	}

	// Validate response from upstream
	if in == nil || in.Id != r.Id {
		log.Printf("Invalid or missing response from upstream %s\n", h.UpstreamAddr)
		m := new(dns.Msg)
		m.SetRcode(r, dns.RcodeServerFailure)
		w.WriteMsg(m)
		return
	}

	// Send response back to client immediately (don't block on database write)
	err = w.WriteMsg(in)
	if err != nil {
		log.Printf("Error sending response back to client: %v\n", err)
		return
	}

	// Process and cache records asynchronously in background goroutine
	// This prevents blocking the response to the client
	go h.cacheResponseAsync(in, qName)
}

// cacheResponseAsync processes upstream response in background without blocking client response
func (h *ForwardHandler) cacheResponseAsync(resp *dns.Msg, qName string) {
	if len(resp.Answer) == 0 {
		return
	}

	for _, rr := range resp.Answer {
		switch record := rr.(type) {
		case *dns.A:
			domain := record.Header().Name
			ip := record.A.String()
			ttl := record.Header().Ttl
			// Save asynchronously - don't block
			db.SaveRecord(domain, "A", ip, int(ttl), "", "")

		case *dns.AAAA:
			domain := record.Header().Name
			ip := record.AAAA.String()
			ttl := record.Header().Ttl
			db.SaveRecord(domain, "AAAA", ip, int(ttl), "", "")

		case *dns.CNAME:
			domain := record.Header().Name
			target := record.Target
			ttl := record.Header().Ttl
			db.SaveRecord(domain, "CNAME", target, int(ttl), "", "")

		case *dns.NS:
			domain := record.Header().Name
			ns := record.Ns
			ttl := record.Header().Ttl
			db.SaveRecord(domain, "NS", ns, int(ttl), "", "")
		}
	}
}
