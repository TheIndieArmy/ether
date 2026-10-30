package arr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func (c *Client) doJSON(ctx context.Context, method, path string, body any, dest any) error {
	if c.BaseURL == "" {
		return fmt.Errorf("url is not configured")
	}
	if c.APIKey == "" {
		return fmt.Errorf("api key is not configured")
	}
	pathOnly := path
	rawQuery := ""
	if i := strings.IndexByte(path, '?'); i >= 0 {
		pathOnly = path[:i]
		rawQuery = path[i+1:]
	}
	full, err := url.JoinPath(c.BaseURL, pathOnly)
	if err != nil {
		return err
	}
	if rawQuery != "" {
		full = full + "?" + rawQuery
	}
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, full, rdr)
	if err != nil {
		return err
	}
	req.Header.Set("X-Api-Key", c.APIKey)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	respBody, err := io.ReadAll(io.LimitReader(res.Body, 8<<20))
	if err != nil {
		return err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		msg := strings.TrimSpace(string(respBody))
		if len(msg) > 200 {
			msg = msg[:200] + "…"
		}
		return fmt.Errorf("HTTP %d: %s", res.StatusCode, msg)
	}
	if dest == nil || len(respBody) == 0 || string(respBody) == "null" {
		return nil
	}
	if err := json.Unmarshal(respBody, dest); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func (c *Client) getJSON(ctx context.Context, path string, dest any) error {
	return c.doJSON(ctx, http.MethodGet, path, nil, dest)
}
