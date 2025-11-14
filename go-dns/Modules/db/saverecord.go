// ...existing code...
package db

import (
	"log"
	"time"

	"gorm.io/gorm"
)

/*func normalizeDomain(d string) string {
	if len(d) > 0 && d[len(d)-1] == '.' {
		return d[:len(d)-1]
	}
	return d
}*/

// update an existing record struct
func UpdateRecord(rec *DNSRecord, ttl int, category, action string) error {
	rec.TTL = ttl
	if category != "" {
		rec.Category = category
	}
	if action != "" {
		rec.Action = action
	}
	rec.UpdatedAt = time.Now()
	return DB.Save(rec).Error
}

// create a new record row
func CreateRecord(domain, recordType, recordValue string, ttl int, category, action string) error {
	rec := DNSRecord{
		Domain:      domain,
		RecordType:  recordType,
		RecordValue: recordValue,
		TTL:         ttl,
		Category:    category,
		Action:      action,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	return DB.Create(&rec).Error
}

// SaveRecord upsert by domain+type+value (doesn't overwrite other distinct values)
func SaveRecord(domain, recordType, recordValue string, ttl int, category, action string) {
	if category == "" {
		category = "Unknown"
	}
	if action == "" {
		action = "Forward"
	}

	// try exact match first
	if rec, err := FindExactRecord(domain, recordType, recordValue); err == nil {
		if err := UpdateRecord(rec, ttl, category, action); err != nil {
			log.Printf("db.SaveRecord: failed to update record: %v", err)
		} else {
			log.Printf("db.SaveRecord: updated %s %s -> %s", domain, recordType, recordValue)
		}
		return
	} else if err != gorm.ErrRecordNotFound {
		log.Printf("db.SaveRecord: find error: %v", err)
		return
	}

	// create new row if no exact match
	if err := CreateRecord(domain, recordType, recordValue, ttl, category, action); err != nil {
		log.Printf("db.SaveRecord: failed to create record: %v", err)
	} else {
		log.Printf("db.SaveRecord: created %s %s -> %s", domain, recordType, recordValue)
	}
}
