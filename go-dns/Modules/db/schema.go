package db

import (
	"time"
)

// DNS Records Schema on SQL Database
type DNSRecord struct {
	ID          int       `gorm:"primaryKey;autoIncrement" db:"id"`
	Domain      string    `gorm:"not null;index" db:"domain"`
	RecordType  string    `gorm:"not null" db:"record_type"`
	RecordValue string    `gorm:"not null" db:"value"`
	TTL         int       `gorm:"default:300" db:"ttl"`
	Category    string    `gorm:"default:'unknown'" db:"category"`
	Action      string    `gorm:"default:'forward'" db:"action"`
	CreatedAt   time.Time `gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" db:"updated_at"`
}
