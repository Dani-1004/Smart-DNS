package main

import (
	"fmt"
	"log"

	db "github.com/Dani-1004/Smart-DNS/dns-server/Modules/db"
	dns "github.com/Dani-1004/Smart-DNS/dns-server/Modules/dns"
)

func main() {
	db.ConnectDatabase()
	listenAddr := "0.0.0.0:53"
	upstreamAddr := "8.8.8.8:53"
	landingIP := "172.24.183.245"

	fmt.Printf("Starting Smart DNS Server...\n")
	if err := dns.StartServer(listenAddr, upstreamAddr, landingIP); err != nil {
		log.Fatalf("Failed to start DNS server: %v", err)
	}
}
