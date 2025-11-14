package dns

import (
	"fmt"
	"log"

	"github.com/miekg/dns"
)

// StartServer memulai server DNS menggunakan miekg/dns
func StartServer(listenAddr, upstreamAddr string) error {
	// Buat handler yang akan meneruskan request
	forwardHandler := NewForwardHandler(upstreamAddr)

	// Inisialisasi server DNS (UDP)
	server := &dns.Server{Addr: listenAddr, Net: "udp"}

	// Daftarkan handler untuk semua query
	dns.HandleFunc(".", forwardHandler.ServeDNS)

	fmt.Printf("DNS server listening on %s (UDP), forwarding to %s\n", listenAddr, upstreamAddr)

	// Server.ListenAndServe() akan memblokir, jadi panggil dalam goroutine jika ingin melanjutkan
	err := server.ListenAndServe()
	if err != nil {
		log.Printf("Failed to start UDP server: %v\n", err)
	}
	return err
}
