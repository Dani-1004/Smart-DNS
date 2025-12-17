package db

import (
	"net"

	"github.com/miekg/dns"

	"strconv"

	"github.com/google/uuid"
)

func SaveDNSRecord(domain, rType, rValue string, rTTL uint32) error {
	id := uuid.New().String()
	recordKey := "rec:" + id
	domainKey := "domain:" + domain

	// 1. Simpan record detail
	err := Rdb.HSet(Ctx, recordKey, map[string]interface{}{
		"type":  rType,
		"value": rValue,
		"ttl":   rTTL,
	}).Err()
	if err != nil {
		return err
	}

	// 2. Tambahkan record ke list domain
	return Rdb.SAdd(Ctx, domainKey, recordKey).Err()
}

func GetDNSRecords(domain string) ([]DNSCacheRecord, error) {
	domainKey := "domain:" + domain

	recordKeys, err := Rdb.SMembers(Ctx, domainKey).Result()
	if err != nil {
		return nil, err
	}

	var list []DNSCacheRecord

	for _, key := range recordKeys {
		data, err := Rdb.HGetAll(Ctx, key).Result()
		if err != nil {
			continue
		}
		if len(data) == 0 {
			continue
		}

		ttl, err := strconv.Atoi(data["ttl"])
		if err != nil {
			ttl = 0
		}
		list = append(list, DNSCacheRecord{
			rType:  data["type"],
			rValue: data["value"],
			rTTL:   ttl,
		})
	}

	return list, nil
}

func ConvertToDNSResponse(records []DNSCacheRecord, qName string) []dns.RR {
	var rrList []dns.RR

	for _, rec := range records {
		switch rec.rType {
		case "A":
			// Convert IP string → dns.A
			ip := net.ParseIP(rec.rValue)
			if ip == nil {
				continue
			}
			rr := &dns.A{
				Hdr: dns.RR_Header{
					Name:   qName,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    uint32(rec.rTTL),
				},
				A: ip,
			}
			rrList = append(rrList, rr)

		case "CNAME":
			rr := &dns.CNAME{
				Hdr: dns.RR_Header{
					Name:   qName,
					Rrtype: dns.TypeCNAME,
					Class:  dns.ClassINET,
					Ttl:    uint32(rec.rTTL),
				},
				Target: rec.rValue,
			}
			rrList = append(rrList, rr)

		case "AAAA":
			rr := &dns.AAAA{
				Hdr: dns.RR_Header{
					Name:   qName,
					Rrtype: dns.TypeAAAA,
					Class:  dns.ClassINET,
					Ttl:    uint32(rec.rTTL),
				},
				AAAA: net.ParseIP(rec.rValue),
			}
			rrList = append(rrList, rr)

		case "NS":
			rr := &dns.NS{
				Hdr: dns.RR_Header{
					Name:   qName,
					Rrtype: dns.TypeNS,
					Class:  dns.ClassINET,
					Ttl:    uint32(rec.rTTL),
				},
				Ns: rec.rValue,
			}
			rrList = append(rrList, rr)
		}
	}
	return rrList
}
