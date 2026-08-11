package runtimeapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"nofx-ta-lib-runtime/talib"
)

var (
	repoRootOnce sync.Once
	repoRootPath string
	repoRootErr  error
	catalogMu    sync.RWMutex
	catalogCache []talib.FunctionCatalogItem
)

// firstExistingPath 优先读取容器显式配置，避免运行时依赖本地仓库目录结构。
func firstExistingPath(envKey string, fallback func() (string, error)) (string, error) {
	if configured := strings.TrimSpace(os.Getenv(envKey)); configured != "" {
		if fileExists(configured) {
			log.Printf("[ta-lib-runtime] 使用环境变量路径 %s=%s", envKey, configured)
			return configured, nil
		}
		log.Printf("[ta-lib-runtime] 环境变量路径不存在 %s=%s", envKey, configured)
		return "", fmt.Errorf("%s 指向的文件不存在: %s", envKey, configured)
	}
	return fallback()
}

func findRepoRoot() (string, error) {
	repoRootOnce.Do(func() {
		wd, err := os.Getwd()
		if err != nil {
			log.Printf("[ta-lib-runtime] 读取工作目录失败: %v", err)
			repoRootErr = err
			return
		}
		log.Printf("[ta-lib-runtime] 开始探测仓库根目录 cwd=%s", wd)
		current := wd
		for {
			bridgePath := filepath.Join(current, "app", "backend", "ta-lib", "scripts", "talib_bridge.py")
			pythonPath := filepath.Join(current, "app", "backend", "ta-lib", "ta-lib-python", ".venv", "bin", "python")
			if fileExists(bridgePath) && fileExists(pythonPath) {
				repoRootPath = current
				log.Printf("[ta-lib-runtime] 已探测到仓库根目录 root=%s", repoRootPath)
				return
			}
			parent := filepath.Dir(current)
			if parent == current {
				repoRootErr = fmt.Errorf("未找到 ta-lib runtime 仓库根目录")
				log.Printf("[ta-lib-runtime] 仓库根目录探测失败 bridge=%s python=%s", bridgePath, pythonPath)
				return
			}
			current = parent
		}
	})
	if repoRootErr != nil {
		return "", repoRootErr
	}
	return repoRootPath, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func pythonExecutable() (string, error) {
	path, err := firstExistingPath("TA_LIB_PYTHON_PATH", func() (string, error) {
		root, err := findRepoRoot()
		if err != nil {
			return "", err
		}
		return filepath.Join(root, "app", "backend", "ta-lib", "ta-lib-python", ".venv", "bin", "python"), nil
	})
	if err != nil {
		log.Printf("[ta-lib-runtime] 解析 Python 路径失败: %v", err)
		return "", err
	}
	log.Printf("[ta-lib-runtime] Python 路径=%s", path)
	return path, nil
}

func bridgeScript() (string, error) {
	path, err := firstExistingPath("TA_LIB_BRIDGE_SCRIPT_PATH", func() (string, error) {
		root, err := findRepoRoot()
		if err != nil {
			return "", err
		}
		return filepath.Join(root, "app", "backend", "ta-lib", "scripts", "talib_bridge.py"), nil
	})
	if err != nil {
		log.Printf("[ta-lib-runtime] 解析 bridge 脚本路径失败: %v", err)
		return "", err
	}
	log.Printf("[ta-lib-runtime] bridge 脚本路径=%s", path)
	return path, nil
}

func runBridge(command string, payload map[string]any) ([]byte, []byte, error) {
	pythonPath, err := pythonExecutable()
	if err != nil {
		return nil, nil, err
	}
	scriptPath, err := bridgeScript()
	if err != nil {
		return nil, nil, err
	}

	cmd := exec.Command(pythonPath, scriptPath, command)
	cmd.Stdin = bytes.NewReader(mustJSON(payload))
	log.Printf(
		"[ta-lib-runtime] 调用 Python bridge command=%s python=%s script=%s payload_keys=%s",
		command,
		pythonPath,
		scriptPath,
		strings.Join(sortedPayloadKeys(payload), ","),
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		errText := strings.TrimSpace(stderr.String())
		if errText == "" {
			errText = err.Error()
		}
		log.Printf(
			"[ta-lib-runtime] Python bridge 执行失败 command=%s err=%s stdout=%s stderr=%s",
			command,
			errText,
			trimLogText(stdout.String()),
			trimLogText(stderr.String()),
		)
		return stdout.Bytes(), stderr.Bytes(), fmt.Errorf(errText)
	}
	log.Printf(
		"[ta-lib-runtime] Python bridge 执行完成 command=%s stdout=%s stderr=%s",
		command,
		trimLogText(stdout.String()),
		trimLogText(stderr.String()),
	)
	return stdout.Bytes(), stderr.Bytes(), nil
}

func mustJSON(value any) []byte {
	raw, _ := json.Marshal(value)
	return raw
}

func loadCatalog() ([]talib.FunctionCatalogItem, error) {
	catalogMu.RLock()
	if catalogCache != nil {
		items := make([]talib.FunctionCatalogItem, len(catalogCache))
		copy(items, catalogCache)
		catalogMu.RUnlock()
		log.Printf("[ta-lib-runtime] 使用 catalog 缓存 count=%d", len(items))
		return items, nil
	}
	catalogMu.RUnlock()

	stdout, _, err := runBridge("catalog", map[string]any{})
	if err != nil {
		log.Printf("[ta-lib-runtime] 加载 catalog 失败: %v", err)
		return nil, err
	}
	var result struct {
		Items []talib.FunctionCatalogItem `json:"items"`
	}
	if err := json.Unmarshal(stdout, &result); err != nil {
		log.Printf("[ta-lib-runtime] 解析 catalog 响应失败 err=%v body=%s", err, trimLogText(string(stdout)))
		return nil, err
	}

	catalogMu.Lock()
	catalogCache = append([]talib.FunctionCatalogItem(nil), result.Items...)
	items := make([]talib.FunctionCatalogItem, len(catalogCache))
	copy(items, catalogCache)
	catalogMu.Unlock()
	log.Printf("[ta-lib-runtime] catalog 加载成功 count=%d", len(items))

	return items, nil
}

func computeIndicator(input talib.ComputeRequest) (*talib.ComputeResponse, int, error) {
	functionKey := strings.ToUpper(strings.TrimSpace(input.FunctionKey))
	if functionKey == "" {
		log.Printf("[ta-lib-runtime] 试算失败：function_key 为空")
		return nil, 400, fmt.Errorf("function_key 不能为空")
	}

	payload := map[string]any{
		"function_key": functionKey,
		"parameters":   input.Parameters,
		"series":       input.Series,
	}
	stdout, stderr, err := runBridge("compute", payload)
	if err != nil {
		if len(stderr) > 0 {
			log.Printf("[ta-lib-runtime] 试算失败 function=%s stderr=%s", functionKey, trimLogText(string(stderr)))
			return nil, 400, fmt.Errorf(strings.TrimSpace(string(stderr)))
		}
		log.Printf("[ta-lib-runtime] 试算失败 function=%s err=%v", functionKey, err)
		return nil, 400, err
	}

	var result talib.ComputeResponse
	if err := json.Unmarshal(stdout, &result); err != nil {
		log.Printf("[ta-lib-runtime] 解析试算结果失败 function=%s err=%v body=%s", functionKey, err, trimLogText(string(stdout)))
		return nil, 500, err
	}
	result.FunctionKey = functionKey
	log.Printf("[ta-lib-runtime] 试算成功 function=%s warnings=%d", functionKey, len(result.Warnings))
	return &result, 200, nil
}

func sortedPayloadKeys(payload map[string]any) []string {
	if len(payload) == 0 {
		return nil
	}
	keys := make([]string, 0, len(payload))
	for key := range payload {
		keys = append(keys, key)
	}
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[j] < keys[i] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}

func trimLogText(value string) string {
	normalized := strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
	if normalized == "" {
		return ""
	}
	if len(normalized) > 300 {
		return normalized[:300] + "..."
	}
	return normalized
}
