package runtimeapi

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"nofx-ta-lib-runtime/talib"
)

var startedAt = time.Now().UTC().Format(time.RFC3339)

type healthPayload struct {
	Status              string `json:"status"`
	Service             string `json:"service"`
	Version             string `json:"version"`
	RegisteredFunctions int    `json:"registered_functions"`
	StartedAt           string `json:"started_at"`
}

// Start 启动 ta-lib runtime 的 HTTP 服务，路由与原 Node 版本保持一致。
func Start() error {
	port := runtimePort()
	log.Printf(
		"[ta-lib-runtime] 启动参数 port=%d TA_LIB_PYTHON_PATH=%s TA_LIB_BRIDGE_SCRIPT_PATH=%s TA_LIBRARY_PATH=%s TA_INCLUDE_PATH=%s",
		port,
		os.Getenv("TA_LIB_PYTHON_PATH"),
		os.Getenv("TA_LIB_BRIDGE_SCRIPT_PATH"),
		os.Getenv("TA_LIBRARY_PATH"),
		os.Getenv("TA_INCLUDE_PATH"),
	)
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/functions", handleFunctions)
	mux.HandleFunc("/compute", handleCompute)

	server := &http.Server{
		Addr:              "0.0.0.0:" + strconv.Itoa(port),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// 先显式监听端口，只有真正 bind 成功后才输出“已启动”，避免端口冲突时误报已就绪。
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return err
	}
	log.Printf("[ta-lib-runtime] 已启动 port=%d", port)
	return server.Serve(listener)
}

func runtimePort() int {
	if raw := os.Getenv("TA_LIB_RUNTIME_PORT"); raw != "" {
		if port, err := strconv.Atoi(raw); err == nil && port > 0 {
			return port
		}
	}
	return 50121
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	_ = r
	catalog, err := loadCatalog()
	if err != nil {
		log.Printf("[ta-lib-runtime] health 检查失败：加载 catalog 失败 err=%v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"message": "读取 TA-Lib 函数目录失败",
			"error":   err.Error(),
		})
		return
	}
	log.Printf("[ta-lib-runtime] health 检查成功 registered_functions=%d", len(catalog))
	writeJSON(w, http.StatusOK, healthPayload{
		Status:              "ok",
		Service:             "ta-lib-runtime",
		Version:             "0.1.0",
		RegisteredFunctions: len(catalog),
		StartedAt:           startedAt,
	})
}

func handleFunctions(w http.ResponseWriter, r *http.Request) {
	_ = r
	catalog, err := loadCatalog()
	if err != nil {
		log.Printf("[ta-lib-runtime] functions 接口失败：加载 catalog 失败 err=%v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"message": "读取 TA-Lib 函数目录失败",
			"error":   err.Error(),
		})
		return
	}
	log.Printf("[ta-lib-runtime] functions 接口成功 count=%d", len(catalog))
	writeJSON(w, http.StatusOK, map[string]any{"items": catalog})
}

func handleCompute(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var input talib.ComputeRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Printf("[ta-lib-runtime] compute 请求解析失败 err=%v", err)
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"message": "解析 TA-Lib 试算请求失败",
			"error":   err.Error(),
		})
		return
	}
	log.Printf("[ta-lib-runtime] 收到 compute 请求 function=%s", input.FunctionKey)
	result, statusCode, err := computeIndicator(input)
	if err != nil {
		log.Printf("[ta-lib-runtime] compute 执行失败 function=%s status=%d err=%v", input.FunctionKey, statusCode, err)
		writeJSON(w, statusCode, map[string]any{
			"message": err.Error(),
		})
		return
	}
	log.Printf("[ta-lib-runtime] compute 执行成功 function=%s", input.FunctionKey)
	writeJSON(w, http.StatusOK, result)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("[ta-lib-runtime] 写回响应失败: %v", err)
	}
}

var _ = sync.Once{}
