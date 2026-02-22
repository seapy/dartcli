package api

import "net/url"

// GetDocumentZIP downloads the disclosure document as a ZIP archive.
// Returns raw ZIP bytes.
func (c *Client) GetDocumentZIP(rceptNo string) ([]byte, error) {
	params := url.Values{}
	params.Set("rcept_no", rceptNo)
	return c.getRaw("/api/document.xml", params)
}
