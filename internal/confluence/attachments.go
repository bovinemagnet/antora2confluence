package confluence

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

func (c *Client) UploadAttachment(pageID, filename string, reader io.Reader) error {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return fmt.Errorf("creating form file: %w", err)
	}
	if _, err := io.Copy(part, reader); err != nil {
		return fmt.Errorf("copying file data: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("closing multipart writer: %w", err)
	}
	reqURL := fmt.Sprintf("%s/rest/api/content/%s/child/attachment", c.baseURL, pageID)
	req, err := http.NewRequest("PUT", reqURL, &body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Atlassian-Token", "nocheck")
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
