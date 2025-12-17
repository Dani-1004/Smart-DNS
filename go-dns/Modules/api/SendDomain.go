package api

import (
	"fmt"

	"github.com/go-resty/resty/v2"
)

type DomainRequest struct {
	Domain      string `json:"domain"`
	RecordType  string `json:"record_type"`
	RecordValue string `json:"record_value"`
	TTL         int    `json:"ttl"`
}

type DomainResponse struct {
	Status  string `json:"Status"`
	Message string `json:"Message"`
}

func SendDomain(domain, recordType, recordValue string, ttl int) {
	request := DomainRequest{
		Domain:      domain,
		RecordType:  recordType,
		RecordValue: recordValue,
		TTL:         ttl,
	}

	client := resty.New()

	respBody := DomainResponse{}

	resp, err := client.R().
		SetHeader("Content-type", "application/json").
		SetBody(request).
		SetResult(&respBody).
		Post("http://192.168.50.13:8000/api/domain")
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	fmt.Println("Response Status Code:", resp.StatusCode())
	fmt.Println("Response Body:", respBody)
}
