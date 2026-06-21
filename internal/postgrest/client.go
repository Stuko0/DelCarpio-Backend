package postgrest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Client struct {
	baseURL  string
	apikey   string
	authTok  string
	http     *http.Client
}

func New(supabaseURL, serviceRoleKey string) *Client {
	return &Client{
		baseURL: supabaseURL + "/rest/v1",
		apikey:  serviceRoleKey,
		authTok: serviceRoleKey,
		http: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *Client) setHeaders(req *http.Request, extra ...[2]string) {
	req.Header.Set("apikey", c.apikey)
	req.Header.Set("Authorization", "Bearer "+c.authTok)
	req.Header.Set("Content-Type", "application/json")
	for _, h := range extra {
		req.Header.Set(h[0], h[1])
	}
}

func (c *Client) List(ctx context.Context, table string, filters url.Values, result interface{}) error {
	u := c.baseURL + "/" + table
	if len(filters) > 0 {
		u += "?" + filters.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	c.setHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("postgrest: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("postgrest %s: %s", resp.Status, string(body))
	}
	return json.Unmarshal(body, result)
}

func (c *Client) GetOne(ctx context.Context, table string, filters url.Values, result interface{}) error {
	var arr []json.RawMessage
	if err := c.List(ctx, table, filters, &arr); err != nil {
		return err
	}
	if len(arr) == 0 {
		return ErrNoRows
	}
	return json.Unmarshal(arr[0], result)
}

func (c *Client) Create(ctx context.Context, table string, payload interface{}, result interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	u := c.baseURL + "/" + table
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return err
	}
	c.setHeaders(req, [2]string{"Prefer", "return=representation"})

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("postgrest: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("postgrest %s: %s", resp.Status, string(respBody))
	}
	if result != nil {
		var arr []json.RawMessage
		if err := json.Unmarshal(respBody, &arr); err != nil {
			return json.Unmarshal(respBody, result)
		}
		if len(arr) > 0 {
			return json.Unmarshal(arr[0], result)
		}
	}
	return nil
}

var ErrNoRows = fmt.Errorf("no rows")

func EqFilter(col, val string) url.Values {
	v := url.Values{}
	v.Set(col, "eq."+val)
	return v
}

func ListFilters(selectCols, filterCol, filterVal, orderCol, orderDir string, limit int) url.Values {
	v := url.Values{}
	v.Set("select", selectCols)
	v.Set(filterCol, "eq."+filterVal)
	v.Set("order", orderCol+"."+orderDir)
	v.Set("limit", strconv.Itoa(limit))
	return v
}
