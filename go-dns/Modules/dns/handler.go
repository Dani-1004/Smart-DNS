package dns

import (
	"log"
	"net"

	"github.com/Dani-1004/Smart-DNS/dns-server/Modules/api"
	"github.com/Dani-1004/Smart-DNS/dns-server/Modules/db"
	"github.com/miekg/dns"
)

// ServeDNS mengimplementasikan interface dns.Handler.
// Fungsi ini menangani request DNS yang masuk, meneruskannya, dan mengirim kembali response.
func (h *ForwardHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	// 1. Tentukan jaringan ("udp" atau "tcp") berdasarkan koneksi client
	netType := w.LocalAddr().Network()

	var qName string

	if len(r.Question) > 0 {
		qName = r.Question[0].Name
	}

	// BLACKLIST & WHITELIST CHECK: return blocked response immediately (fast path)
	category := db.GetCategory(qName)

	if category == "Blacklist" {
		b := new(dns.Msg)
		b.SetReply(r)
		ip := net.ParseIP(h.LandingIP) // Kirim IP Landing Page sebagai jawaban
		aRecord := &dns.A{
			Hdr: dns.RR_Header{
				Name:   qName,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    uint32(600),
			},
			A: ip}
		b.Answer = append(b.Answer, aRecord)
		if err := w.WriteMsg(b); err != nil {
			log.Printf("Error sending blocked response: %v\n", err)
		}
		log.Printf("Blocked (blacklist) %s (action=%s)\n", qName, category)
		return
	}
	// if rec, blocked := db.GetCategory(qName); blocked {

	// }

	if category == "Whitelist" {

		var err error
		// Try to get from cache first (fastest path)
		// 1. Ambil cache dari Redis
		cached, err := db.GetDNSRecords(qName)
		if err != nil {
			log.Printf("Redis cache error for %s: %v", qName, err)
			// lanjut ke upstream → JANGAN return
		} else {
			// 2. Convert structured cache → []dns.RR
			answers := db.ConvertToDNSResponse(cached, qName)
			if len(answers) > 0 {
				// 3. Build DNS reply secara efisien
				resp := new(dns.Msg)
				resp.SetReply(r)
				resp.Answer = answers

				// 4. Kirim ke client
				_ = w.WriteMsg(resp)
				return
			}

			// Jika cached ada tapi tidak bisa dikonversi → fallback upstream
			log.Printf("Cache format invalid for %s, ignoring cache", qName)
		}

		// No cache hit, query upstream
		var in *dns.Msg

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

		// Save to Redis Cache asynchronously
		// go db.SaveDNSRecord()

		// Push to Python for Checking and updating
		go h.cacheResponseAsync(in, qName)
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

	// Push to Python for Checking and Adding records
	go h.domainResponseAsync(in, qName)
}

// domainResponseAsync processes upstream response in background without blocking client response
func (h *ForwardHandler) domainResponseAsync(resp *dns.Msg, qName string) {
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
			// go db.SaveRecord(domain, "A", ip, int(ttl), "", "")
			go api.SendDomain(domain, "A", ip, int(ttl))

		case *dns.AAAA:
			domain := record.Header().Name
			ip := record.AAAA.String()
			ttl := record.Header().Ttl
			// go db.SaveRecord(domain, "AAAA", ip, int(ttl), "", "")
			go api.SendDomain(domain, "AAAA", ip, int(ttl))

		case *dns.CNAME:
			domain := record.Header().Name
			target := record.Target
			ttl := record.Header().Ttl
			// go db.SaveRecord(domain, "CNAME", target, int(ttl), "", "")
			go api.SendDomain(domain, "CNAME", target, int(ttl))

		case *dns.NS:
			domain := record.Header().Name
			ns := record.Ns
			ttl := record.Header().Ttl
			// go db.SaveRecord(domain, "NS", ns, int(ttl), "", "")
			go api.SendDomain(domain, "NS", ns, int(ttl))
		}
	}
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
			// go db.SaveRecord(domain, "A", ip, int(ttl), "", "")
			go db.SaveDNSRecord(domain, "A", ip, ttl)

		case *dns.AAAA:
			domain := record.Header().Name
			ip := record.AAAA.String()
			ttl := record.Header().Ttl
			// go db.SaveRecord(domain, "AAAA", ip, int(ttl), "", "")
			go db.SaveDNSRecord(domain, "AAAA", ip, ttl)

		case *dns.CNAME:
			domain := record.Header().Name
			target := record.Target
			ttl := record.Header().Ttl
			// go db.SaveRecord(domain, "CNAME", target, int(ttl), "", "")
			go db.SaveDNSRecord(domain, "CNAME", target, ttl)

		case *dns.NS:
			domain := record.Header().Name
			ns := record.Ns
			ttl := record.Header().Ttl
			// go db.SaveRecord(domain, "NS", ns, int(ttl), "", "")
			go db.SaveDNSRecord(domain, "NS", ns, ttl)
		}
	}
}
