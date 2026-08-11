package talib

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"nofx-core/shared/config"
)

type RuntimeClient struct {
	baseURL   string
	timeoutMS int
	isEnabled bool
}

func NewRuntimeClient() *RuntimeClient {
	cfg := config.Get()
	baseURL := strings.TrimSpace(cfg.TALibRuntimeBaseURL)
	return NewRuntimeClientWithOptions(baseURL, cfg.TALibRuntimeTimeoutMS, baseURL != "")
}

func NewRuntimeClientWithOptions(baseURL string, timeoutMS int, enabled bool) *RuntimeClient {
	return &RuntimeClient{
		baseURL:   strings.TrimSpace(baseURL),
		timeoutMS: timeoutMS,
		isEnabled: enabled,
	}
}

func (c *RuntimeClient) enabled() bool {
	return c != nil && c.isEnabled && strings.TrimSpace(c.baseURL) != ""
}

func (c *RuntimeClient) doJSON(ctx context.Context, method, path string, input any, output any) (int, error) {
	if !c.enabled() {
		return http.StatusServiceUnavailable, fmt.Errorf("ta-lib runtime 未启用或未配置")
	}
	timeout := time.Duration(c.timeoutMS) * time.Millisecond
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	var bodyReader *bytes.Reader
	if input != nil {
		body, err := json.Marshal(input)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("序列化 ta-lib 请求失败: %w", err)
		}
		bodyReader = bytes.NewReader(body)
	} else {
		bodyReader = bytes.NewReader(nil)
	}
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(c.baseURL, "/")+path, bodyReader)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("创建 ta-lib 请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return http.StatusBadGateway, fmt.Errorf("调用 ta-lib runtime 失败: %w", err)
	}
	defer resp.Body.Close()
	if output != nil {
		if err := json.NewDecoder(resp.Body).Decode(output); err != nil {
			return resp.StatusCode, fmt.Errorf("解析 ta-lib 响应失败: %w", err)
		}
		return resp.StatusCode, nil
	}
	return resp.StatusCode, nil
}

func (c *RuntimeClient) Health(ctx context.Context) (*RuntimeHealthView, error) {
	var result RuntimeHealthView
	if _, err := c.doJSON(ctx, http.MethodGet, "/health", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *RuntimeClient) Functions(ctx context.Context) ([]FunctionCatalogItem, error) {
	var result struct {
		Items []FunctionCatalogItem `json:"items"`
	}
	if _, err := c.doJSON(ctx, http.MethodGet, "/functions", nil, &result); err != nil {
		return nil, err
	}
	return result.Items, nil
}

func (c *RuntimeClient) Compute(ctx context.Context, req ComputeRequest) (*ComputeResponse, int, error) {
	var result ComputeResponse
	statusCode, err := c.doJSON(ctx, http.MethodPost, "/compute", req, &result)
	if err != nil {
		return nil, statusCode, err
	}
	return &result, statusCode, nil
}
