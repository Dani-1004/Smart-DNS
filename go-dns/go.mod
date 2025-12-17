module github.com/Dani-1004/Smart-DNS/dns-server

go 1.23.0

//gorm.io/driver/sqlite v1.6.0 // direct
require gorm.io/gorm v1.31.1 // direct

require (
	github.com/go-resty/resty/v2 v2.16.5
	github.com/google/uuid v1.6.0
	github.com/miekg/dns v1.1.68
	github.com/redis/go-redis/v9 v9.17.1
	gorm.io/driver/mysql v1.6.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/mod v0.24.0 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sync v0.14.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	golang.org/x/tools v0.33.0 // indirect
)

replace github.com/Dani-1004/Smart-DNS/dns-server => ./
