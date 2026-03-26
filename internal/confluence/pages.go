package confluence

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func (c *Client) GetPage(id string) (*Page, error) {
	reqURL := fmt.Sprintf("%s/api/v2/pages/%s?body-format=storage", c.baseURL, id)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	var page Page
	if err := c.decodeResponse(resp, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

func (c *Client) GetChildPages(parentID string) ([]Page, error) {
	reqURL := fmt.Sprintf("%s/api/v2/pages/%s/children", c.baseURL, parentID)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	var list PageList
	if err := c.decodeResponse(resp, &list); err != nil {
		return nil, err
	}
	return list.Results, nil
}

func (c *Client) GetPageByTitle(spaceID, title string) (*Page, error) {
	reqURL := fmt.Sprintf("%s/api/v2/pages?space-id=%s&title=%s&status=current",
		c.baseURL, spaceID, url.QueryEscape(title))
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	var list PageList
	if err := c.decodeResponse(resp, &list); err != nil {
		return nil, err
	}
	if len(list.Results) == 0 {
		return nil, nil
	}
	return &list.Results[0], nil
}

func (c *Client) CreatePage(cpr CreatePageRequest) (*Page, error) {
	body, err := json.Marshal(cpr)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v2/pages", c.baseURL), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	var page Page
	if err := c.decodeResponse(resp, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

func (c *Client) UpdatePage(id string, upr UpdatePageRequest) (*Page, error) {
	body, err := json.Marshal(upr)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/api/v2/pages/%s", c.baseURL, id), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	var page Page
	if err := c.decodeResponse(resp, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

func (c *Client) AddLabels(pageID string, labels []string) error {
	payload := make([]Label, len(labels))
	for i, l := range labels {
		payload[i] = Label{Prefix: "global", Name: l}
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST",
		fmt.Sprintf("%s/rest/api/content/%s/label", c.baseURL, pageID),
		bytes.NewReader(body))
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
