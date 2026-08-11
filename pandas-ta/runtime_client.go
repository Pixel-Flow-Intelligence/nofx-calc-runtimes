package pandas_ta

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"nofx-core/shared/config"
)

// RuntimeClient 负责访问独立 pandas-ta runtime。
type RuntimeClient struct {
	baseURL   string
	timeoutMS int
	enabled   bool
}

func NewRuntimeClient() *RuntimeClient {
	cfg := config.Get()
	return NewRuntimeClientWithOptions(strings.TrimSpace(cfg.PandasTaRuntimeBaseURL), cfg.PandasTaRuntimeTimeoutMS, strings.TrimSpace(cfg.PandasTaRuntimeBaseURL) != "")
}

func NewRuntimeClientWithOptions(baseURL string, timeoutMS int, enabled bool) *RuntimeClient {
	return &RuntimeClient{
		baseURL:   strings.TrimSpace(baseURL),
		timeoutMS: timeoutMS,
		enabled:   enabled,
	}
}

func (c *RuntimeClient) enabledOK() bool {
	return c != nil && c.enabled && strings.TrimSpace(c.baseURL) != ""
}

func (c *RuntimeClient) doJSON(ctx context.Context, method, path string, input any, output any) (int, error) {
	if !c.enabledOK() {
		return http.StatusServiceUnavailable, fmt.Errorf("pandas-ta runtime 未启用或未配置")
	}
	timeout := time.Duration(c.timeoutMS) * time.Millisecond
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	var bodyReader *bytes.Reader
	if input != nil {
		body, err := json.Marshal(input)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("序列化 pandas-ta 请求失败: %w", err)
		}
		bodyReader = bytes.NewReader(body)
	} else {
		bodyReader = bytes.NewReader(nil)
	}
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(c.baseURL, "/")+path, bodyReader)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("创建 pandas-ta 请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return http.StatusBadGateway, fmt.Errorf("调用 pandas-ta runtime 失败: %w", err)
	}
	defer resp.Body.Close()
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return resp.StatusCode, fmt.Errorf("读取 pandas-ta 响应失败: %w", readErr)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		message := strings.TrimSpace(string(body))
		if len(body) > 0 {
			var payload map[string]any
			if err := json.Unmarshal(body, &payload); err == nil {
				if v, ok := payload["error"]; ok && strings.TrimSpace(fmt.Sprint(v)) != "" {
					message = strings.TrimSpace(fmt.Sprint(v))
				}
				if message == "" {
					if v, ok := payload["message"]; ok && strings.TrimSpace(fmt.Sprint(v)) != "" {
						message = strings.TrimSpace(fmt.Sprint(v))
					}
				}
			}
		}
		if message == "" {
			message = fmt.Sprintf("pandas-ta runtime 返回异常状态: %d", resp.StatusCode)
		}
		return resp.StatusCode, fmt.Errorf(message)
	}
	if output != nil {
		if err := json.Unmarshal(body, output); err != nil {
			return resp.StatusCode, fmt.Errorf("解析 pandas-ta 响应失败: %w", err)
		}
		return resp.StatusCode, nil
	}
	return resp.StatusCode, nil
}

type RuntimeHealthView struct {
	Status              string `json:"status"`
	Service             string `json:"service"`
	Version             string `json:"version"`
	RegisteredFunctions int    `json:"registered_functions"`
	StartedAt           string `json:"started_at"`
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
