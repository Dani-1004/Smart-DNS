module github.com/Dani-1004/Smart-DNS/dns-server

go 1.23.0

require (
	gorm.io/driver/sqlite v1.6.0 // direct
	gorm.io/gorm v1.31.1 // direct
)

require github.com/miekg/dns v1.1.68

require (
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	golang.org/x/mod v0.24.0 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sync v0.14.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	golang.org/x/tools v0.33.0 // indirect
)

replace github.com/Dani-1004/Smart-DNS/dns-server => ./
