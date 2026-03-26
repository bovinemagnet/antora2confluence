package confluence

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *Client) GetPageProperty(pageID, key string) (*Property, error) {
	reqURL := fmt.Sprintf("%s/rest/api/content/%s/property/%s", c.baseURL, pageID, key)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	var prop Property
	if err := json.NewDecoder(resp.Body).Decode(&prop); err != nil {
		return nil, err
	}
	return &prop, nil
}

func (c *Client) SetPageProperty(pageID string, prop Property) error {
	body, err := json.Marshal(prop)
	if err != nil {
		return err
	}
	var method, reqURL string
	if prop.Version != nil && prop.Version.Number > 1 {
		method = "PUT"
		reqURL = fmt.Sprintf("%s/rest/api/content/%s/property/%s", c.baseURL, pageID, prop.Key)
	} else {
		method = "POST"
		reqURL = fmt.Sprintf("%s/rest/api/content/%s/property", c.baseURL, pageID)
	}
	req, err := http.NewRequest(method, reqURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
