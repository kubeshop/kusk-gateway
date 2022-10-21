package cloudentity

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var (
	HeaderAuthorizerURL = "X-Kusk-Authorizer-URL"
	HeaderAPIGroup      = "X-Kusk-API-Group"
)

type Client struct {
	hc  http.Client
	url string
}

func New(url string) *Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &Client{
		hc:  http.Client{Timeout: 1 * time.Minute, Transport: tr},
		url: url,
	}
}

func (c *Client) PutAPIGroups(ctx context.Context, req *APIGroupsRequest) error {
	b, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPut, c.url+"/apis", bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("http new request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := c.hc.Do(httpReq)
	if err != nil {
		return fmt.Errorf("http client do: %w", err)
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	statusOK := resp.StatusCode >= 200 && resp.StatusCode < 300
	if !statusOK {
		return fmt.Errorf("error response status: %s", resp.Status)
	}
	return nil
}

func (c *Client) GetAPIGroups(ctx context.Context) ([]APIGroups, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url+"/apis", nil)
	if err != nil {
		return nil, fmt.Errorf("http new request: %w", err)
	}

	resp, err := c.hc.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http client do: %w", err)
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error response status: %s", resp.Status)
	}

	var apiGroups APIGroupsRequest
	err = json.NewDecoder(resp.Body).Decode(&apiGroups)
	if err != nil {
		return nil, fmt.Errorf("json decode: %w", err)
	}

	return apiGroups.APIGroups, nil
}

func (c *Client) Validate(ctx context.Context, req *ValidateRequest) error {
	b, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url+"/request/validate", bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("http new request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := c.hc.Do(httpReq)
	if err != nil {
		return fmt.Errorf("http client do: %w", err)
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	statusOK := resp.StatusCode >= 200 && resp.StatusCode < 300
	if !statusOK {
		return fmt.Errorf("error response status: %s", resp.Status)
	}
	return nil
}

type APIGroupsRequest struct {
	APIGroups []APIGroups `json:"api_groups"`
}

type APIGroups struct {
	Name string `json:"name"`
	ID   string `json:"id"`
	Apis []API  `json:"apis"`
}
type API struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}

type ValidateRequest struct {
	APIGroup    string              `json:"api_group"`
	Method      string              `json:"method"`
	Path        string              `json:"path"`
	Headers     http.Header         `json:"headers"`
	QueryParams map[string][]string `json:"query_params"`
	Body        string              `json:"body"`
}
