package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	apiVersionHeader = "2"
	requestTimeout   = 10 * time.Second
)

type apiMode string

const (
	apiModeV1 apiMode = "v1"
	apiModeV2 apiMode = "v2"
)

type deviceInfo map[string]any
type measurement map[string]any

type homeWizardClient struct {
	baseURL    string
	httpClient *http.Client
	tokenStore *tokenStore
	userName   string
	mode       apiMode
}

type homeWizardError struct {
	Error string `json:"error"`
}

type pairResponse struct {
	Name  string `json:"name"`
	Token string `json:"token"`
}

func newHomeWizardClient(cfg config, store *tokenStore) *homeWizardClient {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		MinVersion:         tls.VersionTLS12,
	}

	client := &homeWizardClient{
		baseURL: cfg.HomeWizardHost,
		httpClient: &http.Client{
			Timeout:   requestTimeout,
			Transport: transport,
		},
		tokenStore: store,
		userName:   cfg.HomeWizardUserName,
		mode:       apiModeFromURL(cfg.HomeWizardHost),
	}

	if cfg.Token != "" {
		_ = client.tokenStore.Set(cfg.Token, cfg.HomeWizardUserName)
	}

	return client
}

func apiModeFromURL(raw string) apiMode {
	if strings.HasPrefix(strings.ToLower(raw), "https://") {
		return apiModeV2
	}
	return apiModeV1
}

func (c *homeWizardClient) AuthRequired() bool {
	return c.mode == apiModeV2
}

func (c *homeWizardClient) Pair(ctx context.Context) (*pairResponse, int, error) {
	if !c.AuthRequired() {
		return nil, http.StatusOK, nil
	}

	body := map[string]string{"name": c.userName}
	req, err := c.newJSONRequest(ctx, http.MethodPost, "/api/user", body, "")
	if err != nil {
		return nil, 0, err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("pair with HomeWizard: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusForbidden {
		var apiErr homeWizardError
		_ = json.NewDecoder(res.Body).Decode(&apiErr)
		return nil, res.StatusCode, fmt.Errorf("pairing not enabled yet: %s", strings.TrimSpace(apiErr.Error))
	}

	if res.StatusCode != http.StatusOK {
		return nil, res.StatusCode, decodeHTTPError(res)
	}

	var paired pairResponse
	if err := json.NewDecoder(res.Body).Decode(&paired); err != nil {
		return nil, res.StatusCode, fmt.Errorf("decode pair response: %w", err)
	}

	if err := c.tokenStore.Set(paired.Token, paired.Name); err != nil {
		return nil, res.StatusCode, err
	}

	return &paired, res.StatusCode, nil
}

func (c *homeWizardClient) GetDeviceInfo(ctx context.Context) (deviceInfo, error) {
	var out deviceInfo
	if err := c.getJSON(ctx, http.MethodGet, "/api", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *homeWizardClient) GetMeasurement(ctx context.Context) (measurement, error) {
	var out measurement
	path := "/api/measurement"
	if c.mode == apiModeV1 {
		path = "/api/v1/data"
	}
	if err := c.getJSON(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *homeWizardClient) GetTelegram(ctx context.Context) (string, error) {
	path := "/api/telegram"
	if c.mode == apiModeV1 {
		path = "/api/v1/telegram"
	}
	token := ""
	if c.AuthRequired() {
		token = c.tokenStore.Get()
		if token == "" {
			return "", errors.New("no HomeWizard token configured; call POST /pair first")
		}
	}
	req, err := c.newRequest(ctx, http.MethodGet, path, nil, token)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "text/plain")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request telegram: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", decodeHTTPError(res)
	}

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("read telegram response: %w", err)
	}

	return string(raw), nil
}

func (c *homeWizardClient) getJSON(ctx context.Context, method, path string, body any, out any) error {
	token := ""
	if c.AuthRequired() {
		token = c.tokenStore.Get()
		if token == "" {
			return errors.New("no HomeWizard token configured; call POST /pair first")
		}
	}

	req, err := c.newJSONRequest(ctx, method, path, body, token)
	if err != nil {
		return err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request %s: %w", path, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return decodeHTTPError(res)
	}

	if err := json.NewDecoder(res.Body).Decode(out); err != nil {
		return fmt.Errorf("decode %s response: %w", path, err)
	}

	return nil
}

func (c *homeWizardClient) newJSONRequest(ctx context.Context, method, path string, body any, token string) (*http.Request, error) {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		reader = strings.NewReader(string(payload))
	}
	return c.newRequest(ctx, method, path, reader, token)
}

func (c *homeWizardClient) newRequest(ctx context.Context, method, path string, body io.Reader, token string) (*http.Request, error) {
	target, err := url.JoinPath(c.baseURL, path)
	if err != nil {
		return nil, fmt.Errorf("build request URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, target, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	if c.mode == apiModeV2 {
		req.Header.Set("X-Api-Version", apiVersionHeader)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return req, nil
}

func decodeHTTPError(res *http.Response) error {
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("remote status %d", res.StatusCode)
	}

	var apiErr homeWizardError
	if json.Unmarshal(body, &apiErr) == nil && apiErr.Error != "" {
		return fmt.Errorf("remote status %d: %s", res.StatusCode, apiErr.Error)
	}

	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return fmt.Errorf("remote status %d", res.StatusCode)
	}

	return fmt.Errorf("remote status %d: %s", res.StatusCode, trimmed)
}
