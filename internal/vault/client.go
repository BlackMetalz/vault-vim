package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/BlackMetalz/vault-vim/internal/config"
)

type Client struct {
	addr       string
	token      string
	httpClient *http.Client
}

type Mount struct {
	Path string
	Type string
}

type Secret struct {
	Data     map[string]string
	Metadata map[string]interface{}
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		addr:       cfg.VaultAddr,
		token:      cfg.VaultToken,
		httpClient: &http.Client{},
	}
}

// ListMounts returns all KV v2 secret engine mounts.
func (c *Client) ListMounts() ([]Mount, error) {
	body, err := c.request("GET", "/v1/sys/mounts", nil)
	if err != nil {
		return nil, fmt.Errorf("list mounts: %w", err)
	}

	var resp map[string]json.RawMessage
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse mounts response: %w", err)
	}

	// The response has mount paths as keys with their config as values.
	// We also need to handle the case where mounts are nested under a "data" key
	// or directly at the top level.
	var mounts []Mount
	for path, raw := range resp {
		// Skip non-mount keys
		if path == "request_id" || path == "lease_id" || path == "renewable" ||
			path == "lease_duration" || path == "wrap_info" || path == "warnings" ||
			path == "auth" || path == "data" {
			continue
		}

		var info struct {
			Type    string                 `json:"type"`
			Options map[string]interface{} `json:"options"`
		}
		if err := json.Unmarshal(raw, &info); err != nil {
			continue
		}

		if info.Type == "kv" {
			if v, ok := info.Options["version"]; ok && fmt.Sprint(v) == "2" {
				mounts = append(mounts, Mount{Path: path, Type: "kv-v2"})
			}
		}
	}

	sort.Slice(mounts, func(i, j int) bool {
		return mounts[i].Path < mounts[j].Path
	})

	return mounts, nil
}

// ListSecrets lists keys at the given path within a mount.
func (c *Client) ListSecrets(mount, path string) ([]string, error) {
	url := fmt.Sprintf("/v1/%smetadata/%s", mount, path)
	body, err := c.request("LIST", url, nil)
	if err != nil {
		// 404 means empty path — not an error
		if strings.Contains(err.Error(), "HTTP 404") {
			return []string{}, nil
		}
		return nil, fmt.Errorf("list secrets at %s%s: %w", mount, path, err)
	}

	var resp struct {
		Data struct {
			Keys []string `json:"keys"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse list response: %w", err)
	}

	sort.Strings(resp.Data.Keys)
	return resp.Data.Keys, nil
}

// GetSecret reads a secret's KV pairs.
func (c *Client) GetSecret(mount, path string) (*Secret, error) {
	// Remove trailing slash if present
	path = strings.TrimSuffix(path, "/")
	url := fmt.Sprintf("/v1/%sdata/%s", mount, path)
	body, err := c.request("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("get secret %s%s: %w", mount, path, err)
	}

	var resp struct {
		Data struct {
			Data     map[string]interface{} `json:"data"`
			Metadata map[string]interface{} `json:"metadata"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse secret response: %w", err)
	}

	data := make(map[string]string)
	for k, v := range resp.Data.Data {
		data[k] = fmt.Sprint(v)
	}

	return &Secret{
		Data:     data,
		Metadata: resp.Data.Metadata,
	}, nil
}

// PatchSecret updates specific keys without overwriting others.
// Tries PATCH first. Falls back to read-merge-write (GET + PUT) if PATCH is
// not permitted (403) or secret doesn't exist yet (404).
func (c *Client) PatchSecret(mount, path string, updates map[string]string) error {
	path = strings.TrimSuffix(path, "/")
	url := fmt.Sprintf("/v1/%sdata/%s", mount, path)

	payload := map[string]interface{}{
		"data": updates,
	}
	payloadBytes, _ := json.Marshal(payload)

	_, err := c.requestWithMethod("PATCH", url, payloadBytes, "application/merge-patch+json")
	if err == nil {
		return nil
	}

	// PATCH failed — fall back to read-merge-write
	errStr := err.Error()
	if strings.Contains(errStr, "HTTP 403") || strings.Contains(errStr, "HTTP 404") {
		return c.readMergeWrite(mount, path, updates)
	}
	return fmt.Errorf("patch secret %s%s: %w", mount, path, err)
}

// readMergeWrite reads existing secret, merges updates, and writes back via PUT.
func (c *Client) readMergeWrite(mount, path string, updates map[string]string) error {
	existing := make(map[string]string)

	// Try to read existing data (ignore 404 — secret may not exist yet)
	secret, err := c.GetSecret(mount, path)
	if err == nil && secret != nil {
		existing = secret.Data
	}

	// Merge updates into existing
	for k, v := range updates {
		existing[k] = v
	}

	return c.PutSecret(mount, path, existing)
}

// PutSecret writes a full secret (used for adding keys - reads first, merges, then writes).
func (c *Client) PutSecret(mount, path string, data map[string]string) error {
	path = strings.TrimSuffix(path, "/")
	url := fmt.Sprintf("/v1/%sdata/%s", mount, path)

	payload := map[string]interface{}{
		"data": data,
	}
	payloadBytes, _ := json.Marshal(payload)

	_, err := c.request("POST", url, payloadBytes)
	if err != nil {
		return fmt.Errorf("put secret %s%s: %w", mount, path, err)
	}
	return nil
}

// DeleteKey removes a specific key from a secret by reading, removing the key, and writing back.
func (c *Client) DeleteKey(mount, path, key string) error {
	secret, err := c.GetSecret(mount, path)
	if err != nil {
		return fmt.Errorf("read before delete: %w", err)
	}

	delete(secret.Data, key)
	return c.PutSecret(mount, path, secret.Data)
}

// RenameKey renames a key by reading the secret, copying the value to the new key, and removing the old key.
func (c *Client) RenameKey(mount, path, oldKey, newKey string) error {
	secret, err := c.GetSecret(mount, path)
	if err != nil {
		return fmt.Errorf("read before rename: %w", err)
	}

	val, ok := secret.Data[oldKey]
	if !ok {
		return fmt.Errorf("key %q not found", oldKey)
	}

	secret.Data[newKey] = val
	delete(secret.Data, oldKey)
	return c.PutSecret(mount, path, secret.Data)
}

// DeleteSecret deletes a secret (or recursively deletes all secrets under a path).
func (c *Client) DeleteSecret(mount, path string) error {
	path = strings.TrimSuffix(path, "/")

	// Try to delete as a single secret first
	url := fmt.Sprintf("/v1/%smetadata/%s", mount, path)
	_, err := c.request("DELETE", url, nil)
	if err != nil {
		// If it's a folder, list and delete recursively
		items, listErr := c.ListSecrets(mount, path+"/")
		if listErr != nil {
			return fmt.Errorf("delete %s%s: %w", mount, path, err)
		}
		for _, item := range items {
			subPath := path + "/" + item
			if err := c.DeleteSecret(mount, subPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Client) request(method, path string, body []byte) ([]byte, error) {
	return c.requestWithMethod(method, path, body, "application/json")
}

func (c *Client) requestWithMethod(method, path string, body []byte, contentType string) ([]byte, error) {
	url := c.addr + path

	var reqBody io.Reader
	if body != nil {
		reqBody = strings.NewReader(string(body))
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Vault-Token", c.token)
	if body != nil {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("vault API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
