package pandas_ta

import (
	"context"
	"encoding/json"
	"strings"

	"gorm.io/gorm"
	"nofx-core/shared/config"
	coredb "nofx-core/shared/db"
	"nofx-core/shared/logger"
	"nofx-pandas-ta/store"
)

// Service 聚合 pandas-ta runtime client 和治理存储。
type Service struct {
	st     *coredb.Store
	client *RuntimeClient
}

func NewService(st *coredb.Store) *Service {
	return &Service{
		st:     st,
		client: newRuntimeClientFromStore(st),
	}
}

func newRuntimeClientFromStore(st *coredb.Store) *RuntimeClient {
	envCfg := config.Get()
	baseURL := strings.TrimSpace(envCfg.PandasTaRuntimeBaseURL)
	timeoutMS := envCfg.PandasTaRuntimeTimeoutMS
	enabled := baseURL != ""
	if st != nil {
		cfg, err := store.NewPandasTaRuntimeConfigStore(st.GormDB()).GetLatest()
		if err == nil && cfg != nil {
			// 优先使用管理端保存的运行时地址，避免页面已配置但服务端仍读取为空环境变量。
			if strings.TrimSpace(cfg.BaseURL) != "" {
				baseURL = strings.TrimSpace(cfg.BaseURL)
			}
			if cfg.TimeoutMS > 0 {
				timeoutMS = cfg.TimeoutMS
			}
			enabled = cfg.Enabled && baseURL != ""
			return NewRuntimeClientWithOptions(baseURL, timeoutMS, enabled)
		}
		if err != nil && err != gorm.ErrRecordNotFound {
			logger.Warnf("读取 pandas-ta runtime 数据库配置失败，回退到环境变量：%v", err)
		}
	}
	return NewRuntimeClientWithOptions(baseURL, timeoutMS, enabled)
}

// GetRuntimeView 返回运行时健康态与本地配置。
func (s *Service) GetRuntimeView(ctx context.Context) (map[string]any, error) {
	view := map[string]any{"enabled": false}
	health, err := s.client.Health(ctx)
	if err != nil {
		view["error"] = err.Error()
	} else {
		view["enabled"] = true
		view["health"] = health
	}
	cfg, err := store.NewPandasTaRuntimeConfigStore(s.st.GormDB()).GetLatest()
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if cfg != nil {
		view["config"] = cfg
	}
	return view, nil
}

// ListFunctions 返回 runtime 函数目录与治理状态合并结果。
func (s *Service) ListFunctions(ctx context.Context) ([]map[string]any, error) {
	govItems, err := store.NewPandasTaFunctionGovernanceStore(s.st.GormDB()).ListAll()
	if err != nil {
		return nil, err
	}
	catalog := make([]FunctionCatalogItem, 0)
	if s.client.enabledOK() {
		var runtimeItems []FunctionCatalogItem
		runtimeItems, err = s.client.Functions(ctx)
		if err != nil {
			logger.Warnf("读取 pandas-ta runtime 函数目录失败，改为返回治理侧数据：%v", err)
		} else {
			catalog = runtimeItems
		}
	} else {
		logger.Infof("pandas-ta runtime 未启用，函数目录仅返回治理侧数据")
	}
	govMap := make(map[string]store.PandasTaFunctionGovernance, len(govItems))
	for _, item := range govItems {
		govMap[item.FunctionKey] = item
	}
	out := make([]map[string]any, 0, max(len(catalog), len(govItems)))
	for _, item := range catalog {
		gov := defaultFunctionGovernance(item)
		if override, ok := govMap[item.FunctionKey]; ok {
			gov = mergeFunctionGovernance(item, override)
		}
		out = append(out, mergeFunctionCatalog(item, gov))
	}
	for _, gov := range govItems {
		if _, exists := findFunctionByKey(catalog, gov.FunctionKey); exists {
			continue
		}
		out = append(out, mergeGovernanceOnly(gov))
	}
	return out, nil
}

func defaultFunctionGovernance(item FunctionCatalogItem) store.PandasTaFunctionGovernance {
	return store.PandasTaFunctionGovernance{
		FunctionKey:       item.FunctionKey,
		GroupName:         item.GroupName,
		DisplayName:       item.DisplayName,
		DescriptionZH:     item.DescriptionZH,
		Enabled:           true,
		AgentEnabled:      true,
		AdminTrialEnabled: true,
		DefaultParamsJSON: "{}",
	}
}

func mergeFunctionGovernance(item FunctionCatalogItem, override store.PandasTaFunctionGovernance) store.PandasTaFunctionGovernance {
	result := defaultFunctionGovernance(item)
	if strings.TrimSpace(override.FunctionKey) != "" {
		result.FunctionKey = override.FunctionKey
	}
	if strings.TrimSpace(override.GroupName) != "" {
		result.GroupName = override.GroupName
	}
	if strings.TrimSpace(override.DisplayName) != "" {
		result.DisplayName = override.DisplayName
	}
	if strings.TrimSpace(override.DescriptionZH) != "" {
		result.DescriptionZH = override.DescriptionZH
	}
	result.Enabled = override.Enabled
	result.AgentEnabled = override.AgentEnabled
	result.AdminTrialEnabled = override.AdminTrialEnabled
	if strings.TrimSpace(override.DefaultParamsJSON) != "" {
		result.DefaultParamsJSON = override.DefaultParamsJSON
	}
	if strings.TrimSpace(override.RiskNote) != "" {
		result.RiskNote = override.RiskNote
	}
	if strings.TrimSpace(override.ReviewedBy) != "" {
		result.ReviewedBy = override.ReviewedBy
	}
	if !override.ReviewedAt.IsZero() {
		result.ReviewedAt = override.ReviewedAt
	}
	return result
}

func mergeFunctionCatalog(item FunctionCatalogItem, gov store.PandasTaFunctionGovernance) map[string]any {
	return map[string]any{
		"function_key":              item.FunctionKey,
		"display_name":              item.DisplayName,
		"group_name":                item.GroupName,
		"description_zh":            item.DescriptionZH,
		"function_type":             item.FunctionType,
		"parameters_schema":         item.ParametersSchema,
		"input_series_requirements": item.InputSeriesRequirements,
		"output_schema":             item.OutputSchema,
		"render_schema":             item.RenderSchema,
		"warmup_bars":               item.WarmupBars,
		"tags":                      item.Tags,
		"enabled":                   gov.Enabled,
		"agent_enabled":             gov.AgentEnabled,
		"admin_trial_enabled":       gov.AdminTrialEnabled,
		"default_params_json":       gov.DefaultParamsJSON,
		"risk_note":                 gov.RiskNote,
		"reviewed_by":               gov.ReviewedBy,
		"reviewed_at":               gov.ReviewedAt,
	}
}

func mergeGovernanceOnly(gov store.PandasTaFunctionGovernance) map[string]any {
	return map[string]any{
		"function_key":              gov.FunctionKey,
		"display_name":              gov.DisplayName,
		"group_name":                gov.GroupName,
		"description_zh":            gov.DescriptionZH,
		"function_type":             "",
		"parameters_schema":         map[string]any{},
		"input_series_requirements": []string{},
		"output_schema":             map[string]any{},
		"render_schema":             map[string]any{},
		"warmup_bars":               0,
		"tags":                      []string{},
		"enabled":                   gov.Enabled,
		"agent_enabled":             gov.AgentEnabled,
		"admin_trial_enabled":       gov.AdminTrialEnabled,
		"default_params_json":       gov.DefaultParamsJSON,
		"risk_note":                 gov.RiskNote,
		"reviewed_by":               gov.ReviewedBy,
		"reviewed_at":               gov.ReviewedAt,
	}
}

// UpdateRuntimeConfig 保存运行时接入配置。
func (s *Service) UpdateRuntimeConfig(ctx context.Context, item *store.PandasTaRuntimeConfig) error {
	_ = ctx
	return store.NewPandasTaRuntimeConfigStore(s.st.GormDB()).Upsert(item)
}

// UpdateFunctionGovernance 保存函数治理信息。
func (s *Service) UpdateFunctionGovernance(ctx context.Context, item *store.PandasTaFunctionGovernance) error {
	_ = ctx
	return store.NewPandasTaFunctionGovernanceStore(s.st.GormDB()).Upsert(item)
}

// Trial 执行独立 runtime 试算并记录历史。
func (s *Service) Trial(ctx context.Context, req ComputeRequest, createdBy string) (map[string]any, error) {
	result, statusCode, err := s.client.Compute(ctx, req)
	history := &store.PandasTaTrialHistory{
		Symbol:           strings.TrimSpace(req.PriceSource),
		Timeframe:        "",
		FunctionKey:      req.FunctionKey,
		InputSummaryJSON: mustJSON(req),
		ResultSummaryJSON: mustJSON(result),
		Status:           "ok",
		ErrorMessage:     "",
		CreatedBy:        strings.TrimSpace(createdBy),
	}
	if err != nil || statusCode >= 400 {
		history.Status = "failed"
		if err != nil {
			history.ErrorMessage = errString(err)
		} else {
			history.ErrorMessage = "pandas-ta runtime 返回失败状态"
		}
	}
	if histErr := store.NewPandasTaTrialHistoryStore(s.st.GormDB()).Create(history); histErr != nil {
		logger.Warnf("写入 pandas-ta 试算历史失败：%v", histErr)
	}
	if err != nil {
		return nil, err
	}
	return map[string]any{"result": result}, nil
}

func mustJSON(value any) string {
	raw, _ := json.Marshal(value)
	return string(raw)
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func findFunctionByKey(catalog []FunctionCatalogItem, key string) (FunctionCatalogItem, bool) {
	for _, item := range catalog {
		if item.FunctionKey == key {
			return item, true
		}
	}
	return FunctionCatalogItem{}, false
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
