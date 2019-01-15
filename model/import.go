package model

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// Import some resource to ClearPass
func (c *clearpass) Import(ctx context.Context, fileName, importType, pass string) error {
	baseURL := c.webURL
	fullURL := baseURL + "/tipsImport.action"
	req, err := http.NewRequest(string(GET), fullURL, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Add("Referer", baseURL+"/tipsContent.action")
	detail, _ := rawRequest(ctx, c.client, req, nil, false)
	if detail.Err != nil {
		return detail
	}
	if detail.StatusCode != 200 {
		detail.Err = errors.New("Failed to get import token with status != 200")
		return detail
	}
	// Get the struts anti-reply token
	parts := strings.SplitN(string(detail.Reply), `input type="hidden" name="token" value="`, 2)
	if len(parts) < 2 {
		detail.Err = errors.New("Failed to find token in response")
	}
	parts = strings.SplitN(parts[1], `"`, 2)
	if len(parts) < 2 {
		detail.Err = errors.New("Failed to find end delimiter in response")
	}
	token := parts[0]
	// Read the file
	inFile, err := os.Open(fileName)
	if err != nil {
		return errors.Wrapf(err, "Failed to open inFile %s to POST", inFile)
	}
	defer inFile.Close()
	// Build the import body
	buffer := &bytes.Buffer{}
	multip := multipart.NewWriter(buffer)
	fields := map[string]string{
		"struts.token.name": "token",
		"token":             url.QueryEscape(token),
		"type":              url.QueryEscape(importType),
	}
	if pass != "" {
		fields["password"] = url.QueryEscape(pass)
	}
	for k, v := range fields {
		w, err := multip.CreateFormField(k)
		if err != nil {
			return errors.Wrapf(err, "Failed to create multipart field %s")
		}
		w.Write([]byte(v))
	}
	w, err := multip.CreateFormFile("upload", filepath.Base(fileName))
	if err != nil {
		return errors.Wrapf(err, "Failed to create multipart file %s", fileName)
	}
	if _, err := io.Copy(w, inFile); err != nil {
		return errors.Wrapf(err, "Failed to dump multipart file %s to internal buffer", fileName)
	}
	multip.Close()
	// Post the import
	fullURL = baseURL + "/tipsUploadImport.action"
	req, err = http.NewRequest(string(POST), fullURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", multip.FormDataContentType())
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	detail, _ = rawRequest(ctx, c.client, req, buffer.Bytes(), false)
	if detail.Err != nil {
		return detail
	}
	if detail.StatusCode != 200 {
		detail.Err = errors.Errorf("Failed to import file %s with status != 200", fileName)
		return detail
	}
	return nil
}
