package db

import (
	"time"

	"gorm.io/gorm"
)

// CheckBlacklist returns a non-nil DNSRecord and true if the domain is blacklisted/blocked.
// It also respects TTL (returns not found if expired).
func CheckBlacklist(domain string) (*DNSRecord, bool) {
	var rec DNSRecord
	err := DB.Where("domain = ? AND (category = ? OR action = ?)", domain, "Blacklist", "Block").
		Order("updated_at DESC").
		First(&rec).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, false
		}
		// on other DB error, be conservative: not blocked
		return nil, false
	}

	/*if time.Since(rec.UpdatedAt).Seconds() > float64(rec.TTL) {
		return nil, false
	}*/return &rec, true
}

// FindExactRecord finds a row matching domain+type+value
func FindExactRecord(domain, recordType, recordValue string) (*DNSRecord, error) {
	var rec DNSRecord
	err := DB.Where("domain = ? AND record_type = ? AND record_value = ?", domain, recordType, recordValue).
		First(&rec).Error
	return &rec, err
}

// GetRecords returns all (non-expired) records for domain+type, newest first
func GetRecords(domain, recordType string) ([]DNSRecord, error) {
	var records []DNSRecord
	if err := DB.Where("domain = ? AND record_type = ?", domain, recordType).
		Order("updated_at DESC").Find(&records).Error; err != nil {
		return nil, err
	}
	now := time.Now()
	out := records[:0]
	for _, r := range records {
		if now.Sub(r.UpdatedAt).Seconds() <= float64(r.TTL) {
			out = append(out, r)
		}
	}
	// Jika TTL Habis atau tidak ada record
	if len(out) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return out, nil
}
