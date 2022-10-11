// Package netlify contains code to make requests to the Netlify DNS API
package netlify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// API makes it possible to make calls to the Netlify DNS API
type API struct {
	token  string
	client http.Client
}

// GetDNSZoneForDomain fetches the DNS zone from the Netlify API with the matching domain name
func (n *API) GetDNSZoneForDomain(domain string) (DNSZone, error) {
	dnsZones, err := n.getDNSZones(domain)
	if err != nil {
		return DNSZone{}, err
	}

	var zone DNSZone

	for _, z := range dnsZones {
		if z.Name == domain {
			zone = z
		}
	}

	if zone.ID == "" {
		return DNSZone{}, fmt.Errorf("couldn't find zone for domain '%s'", domain)
	}

	return zone, nil
}

// GetDNSRecordsForZone fetches all DNS records from the Netlify API in the given DNS zone
func (n *API) GetDNSRecordsForZone(zoneID string) ([]DNSRecord, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.netlify.com/api/v1/dns_zones/%s/dns_records", zoneID), nil)
	if err != nil {
		return nil, err
	}

	res, err := n.handleRequest(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("response error: %v", string(body))
	}
	decoder := json.NewDecoder(res.Body)
	var records []DNSRecord
	err = decoder.Decode(&records)
	if err != nil {
		return nil, err
	}

	return records, nil
}

// FindDNSForHostname finds a certain DNS record in a list by matching the hostname with the one given
func (n *API) FindDNSForHostname(records []DNSRecord, hostname string) (DNSRecord, error) {
	for _, v := range records {
		if v.Hostname == hostname {
			return v, nil
		}
	}

	return DNSRecord{}, nil
}

// DeleteDNSRecord deletes the DNS Record with the given ID
func (n *API) DeleteDNSRecord(zoneID string, recordID string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("https://api.netlify.com/api/v1/dns_zones/%s/dns_records/%s", zoneID, recordID), nil)
	if err != nil {
		return err
	}
	res, err := n.handleRequest(req)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusNoContent {
		err = n.handleErrorResponse(res)
		return err
	}
	res.Body.Close()
	return nil
}

// CreateDNSRecord creates a DNS record for the given some with the given record data
func (n *API) CreateDNSRecord(zoneID string, record DNSRecordInput) (*DNSRecord, error) {
	body, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://api.netlify.com/api/v1/dns_zones/%s/dns_records", zoneID), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := n.handleRequest(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		err = n.handleErrorResponse(res)
		return nil, err
	}
	decoder := json.NewDecoder(res.Body)
	var newRecord DNSRecord
	err = decoder.Decode(&newRecord)
	if err != nil {
		return nil, err
	}
	return &newRecord, nil
}

func (n *API) getDNSZones(domain string) ([]DNSZone, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.netlify.com/api/v1/dns_zones", nil)
	if err != nil {
		return nil, err
	}
	res, err := n.handleRequest(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		err = n.handleErrorResponse(res)
		return nil, err
	}
	decoder := json.NewDecoder(res.Body)
	var dnsZones []DNSZone
	err = decoder.Decode(&dnsZones)
	if err != nil {
		return nil, err
	}
	return dnsZones, nil
}

func (n *API) handleErrorResponse(res *http.Response) error {
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	return fmt.Errorf("response error - %s - %v - %v", res.Request.URL.Path, res.StatusCode, body)
}

func (n *API) handleRequest(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", n.token))
	req.Header.Add("Accept", "application/json")
	res, err := n.client.Do(req)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return res, nil
}

// NewAPI creates a new API instance
func NewAPI(token string) *API {
	return &API{
		token:  token,
		client: http.Client{},
	}
}

// DNSRecord contains all DNS record data provided by the Netlify DNS API
type DNSRecord struct {
	ID         string `json:"id"`
	Hostname   string `json:"hostname"`
	RecordType string `json:"type"`
	Value      string `json:"value"`
	TTL        int64  `json:"ttl"`
	Priority   int64  `json:"priority"`
	DNSZoneID  string `json:"dns_zone_id"`
	SiteID     string `json:"site_id"`
	Flag       int64  `json:"flag"`
	Tag        string `json:"tag"`
	Tagged     bool   `json:"tagged"`
}

// DNSRecordInput contains all data necessary to create a `A` record with the Netlify APi
type DNSRecordInput struct {
	RecordType string `json:"type"`
	Hostname   string `json:"hostname"`
	Value      string `json:"value"`
	TTL        int64  `json:"ttl"`
}

// DNSZone contains all DNS zone data provided by the Netlify DNS API
type DNSZone struct {
	ID                   string      `json:"id"`
	Name                 string      `json:"name"`
	Errors               []string    `json:"errors"`
	SupportedRecordTypes []string    `json:"supported_record_types"`
	UserID               string      `json:"user_id"`
	CreatedAt            string      `json:"created_at"`
	UpdatedAt            string      `json:"updated_at"`
	Records              []DNSRecord `json:"records"`
	DNSServers           []string    `json:"dns_servers"`
	AccountID            string      `json:"account_id"`
	SiteID               string      `json:"site_id"`
	AccountSlug          string      `json:"account_slug"`
	AccountName          string      `json:"account_name"`
}
