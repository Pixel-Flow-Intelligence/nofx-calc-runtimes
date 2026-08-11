package runtimeapi

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	pandasta "nofx-pandas-ta"
)

var startedAt = time.Now().UTC().Format(time.RFC3339)

// Start 启动 pandas-ta runtime 的 HTTP 服务。
func Start() error {
	port := runtimePort()
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/functions", handleFunctions)
	mux.HandleFunc("/compute", handleCompute)

	server := &http.Server{
		Addr:              "0.0.0.0:" + strconv.Itoa(port),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("[pandas-ta-runtime] 已启动 port=%d", port)
	return server.ListenAndServe()
}

func runtimePort() int {
	if raw := os.Getenv("PANDAS_TA_RUNTIME_PORT"); raw != "" {
		if port, err := strconv.Atoi(raw); err == nil && port > 0 {
			return port
		}
	}
	return 50131
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	_ = r
	items, err := loadCatalog()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"message": "读取 pandas-ta 函数目录失败",
			"error":   err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":               "ok",
		"service":              "pandas-ta-runtime",
		"version":              "0.1.0",
		"registered_functions": len(items),
		"started_at":           startedAt,
	})
}

func handleFunctions(w http.ResponseWriter, r *http.Request) {
	_ = r
	items, err := loadCatalog()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"message": "读取 pandas-ta 函数目录失败",
			"error":   err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func handleCompute(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req pandasta.ComputeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"message": "解析 pandas-ta 试算请求失败",
			"error":   err.Error(),
		})
		return
	}
	result, statusCode, err := runBridge("compute", map[string]any{
		"function_key": req.FunctionKey,
		"parameters":   req.Parameters,
		"series":       req.Series,
		"price_source": req.PriceSource,
	})
	if err != nil {
		writeJSON(w, statusCode, map[string]any{"message": err.Error()})
		return
	}
	var response map[string]any
	if err := json.Unmarshal(result, &response); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"message": "解析 pandas-ta 试算结果失败",
			"error":   err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func loadCatalog() ([]pandasta.FunctionCatalogItem, error) {
	payload, statusCode, err := runBridge("catalog", map[string]any{})
	if err != nil {
		return nil, err
	}
	if statusCode >= http.StatusBadRequest {
		return nil, errTextError("读取 pandas-ta 函数目录失败")
	}
	var result struct {
		Items []pandasta.FunctionCatalogItem `json:"items"`
	}
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	return result.Items, nil
}

func runBridge(command string, payload map[string]any) ([]byte, int, error) {
	pythonPath, err := pythonExecutable()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	scriptPath, err := bridgeScript()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	cmd := exec.Command(pythonPath, scriptPath, command)
	cmd.Stdin = bytes.NewReader(mustJSON(payload))
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		errText := strings.TrimSpace(stderr.String())
		if errText == "" {
			errText = err.Error()
		}
		return stdout.Bytes(), http.StatusInternalServerError, errTextError(errText)
	}
	return stdout.Bytes(), http.StatusOK, nil
}

func pythonExecutable() (string, error) {
	if raw := strings.TrimSpace(os.Getenv("PANDAS_TA_PYTHON_PATH")); raw != "" {
		return raw, nil
	}
	if raw := strings.TrimSpace(os.Getenv("PYTHON")); raw != "" {
		return raw, nil
	}
	if root, err := repoRoot(); err == nil {
		candidate := filepath.Join(root, "app", "backend", "pandas-ta", ".venv", "bin", "python")
		if fileExists(candidate) {
			return candidate, nil
		}
	}
	if raw := strings.TrimSpace(os.Getenv("VIRTUAL_ENV")); raw != "" {
		candidate := filepath.Join(raw, "bin", "python")
		if fileExists(candidate) {
			return candidate, nil
		}
	}
	return "python3", nil
}

func bridgeScript() (string, error) {
	root, err := repoRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "app", "backend", "pandas-ta", "scripts", "pandas_ta_bridge.py"), nil
}

func repoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	current := wd
	for {
		candidate := filepath.Join(current, "app", "backend", "pandas-ta", "scripts", "pandas_ta_bridge.py")
		if fileExists(candidate) {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", os.ErrNotExist
		}
		current = parent
	}
}

func mustJSON(value any) []byte {
	raw, _ := json.Marshal(value)
	return raw
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

type errTextError string

func (e errTextError) Error() string { return string(e) }

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("[pandas-ta-runtime] 写回响应失败: %v", err)
	}
}
