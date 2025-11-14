package main

import (
	"fmt"
	"log"

	db "github.com/Dani-1004/Smart-DNS/dns-server/Modules/db"
	dns "github.com/Dani-1004/Smart-DNS/dns-server/Modules/dns"
)

func main() {
	db.ConnectDatabase()
	listenAddr := "127.0.0.1:8053"
	upstreamAddr := "8.8.8.8:53"

	fmt.Printf("Starting Smart DNS Server...\n")
	if err := dns.StartServer(listenAddr, upstreamAddr); err != nil {
		log.Fatalf("Failed to start DNS server: %v", err)
	}
}
