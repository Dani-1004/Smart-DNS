package main

import (
	"fmt"
	"log"

	db "github.com/Dani-1004/Smart-DNS/dns-server/Modules/db"
	dns "github.com/Dani-1004/Smart-DNS/dns-server/Modules/dns"
)

func main() {
	db.InitRedis()
	listenAddr := "0.0.0.0:53"
	upstreamAddr := "208.67.222.222:443"
	landingIP := "172.30.182.85"

	fmt.Printf("Starting Smart DNS Server...\n")
	if err := dns.StartServer(listenAddr, upstreamAddr, landingIP); err != nil {
		log.Fatalf("Failed to start DNS server: %v", err)
	}
}
