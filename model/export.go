package model

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Export some resource from ClearPass, return the exported stream.
func (c *clearpass) Export(ctx context.Context, resource, pass string) (string, io.ReadCloser, error) {
	baseURL := c.webURL
	fullURL := baseURL + "/tipsExport.action"
	req, err := http.NewRequest(string(POST), fullURL, nil)
	if err != nil {
		return "", nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Add("Accept-Encoding", "gzip")
	req.Header.Add("Accept-Encoding", "deflate")
	req.Header.Add("Accept-Encoding", "br")
	data := make(url.Values)
	data.Add("type", resource)
	data.Add("encyptionPassword", pass) // yes, it is mispelled in ClearPass
	detail, stream := rawRequest(ctx, c.client, req, []byte(data.Encode()), true)
	if detail.Err != nil {
		if stream != nil {
			stream.Close()
		}
		return "", nil, detail
	}
	if detail.StatusCode != 200 {
		detail.Err = errors.New("Failed to download resource with status != 200")
		if stream != nil {
			stream.Close()
		}
		return "", nil, detail
	}
	cd := detail.ReplyHeader.Get("Content-Disposition")
	if cd == "" || !strings.Contains(cd, "filename=") {
		detail.Err = errors.New("Missing Content-Disposition header, or filename in it")
		return "", nil, detail
	}
	parts := strings.Split(cd, "filename=")
	if len(parts) < 2 {
		detail.Err = errors.New("Content-Disposition header incorrectly formatted")
		return "", nil, detail
	}
	return strings.TrimSpace(parts[len(parts)-1]), stream, nil
}
