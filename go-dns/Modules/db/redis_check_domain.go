package db

import (
	"context"
)

var ctx = context.Background()

func CheckDomainExists(domain string) (bool, error) {
	key := "class:" + domain

	exists, err := Rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return exists == 1, nil
}

func GetDomainInfo(domain string) (*DomainInfo, error) {
	key := "class:" + domain

	data, err := Rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil // not found
	}

	info := &DomainInfo{
		Category:  data["category"],
		Action:    data["action"],
		UpdatedAt: data["updated_at"],
	}

	return info, nil
}

func GetCategory(domain string) string {
	info, err := GetDomainInfo(domain)
	if err != nil {
		return ""
	}
	if info == nil {
		return "" // not found
	}
	return info.Category
}
