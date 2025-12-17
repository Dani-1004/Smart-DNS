package dns

import (
	"time"

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
