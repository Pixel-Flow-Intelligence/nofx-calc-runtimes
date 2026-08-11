package trendln

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	"gorm.io/gorm"
	"nofx-core/shared/config"
	coredb "nofx-core/shared/db"
	"nofx-core/shared/logger"
	"nofx-trendln/store"
)

// Service 聚合 trendln runtime client 和治理存储。
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
	baseURL := strings.TrimSpace(envCfg.TrendlnRuntimeBaseURL)
	timeoutMS := envCfg.TrendlnRuntimeTimeoutMS
	enabled := baseURL != ""
	if st != nil {
		cfg, err := store.NewTrendlnRuntimeConfigStore(st.GormDB()).GetLatest()
		if err == nil && cfg != nil {
			// 优先使用管理后台保存的运行时地址，避免页面已配置但服务端仍回退为空环境变量。
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
			logger.Warnf("读取 trendln runtime 数据库配置失败，回退到环境变量：%v", err)
		}
	}
	return NewRuntimeClientWithOptions(baseURL, timeoutMS, enabled)
}

// GetRuntimeView 返回 trendln-runtime 的健康态与管理配置。
func (s *Service) GetRuntimeView(ctx context.Context) (map[string]any, error) {
	view := map[string]any{"enabled": false}
	health, err := s.client.Health(ctx)
	if err != nil {
		view["error"] = err.Error()
	} else {
		view["enabled"] = true
		view["health"] = health
	}
	cfg, err := store.NewTrendlnRuntimeConfigStore(s.st.GormDB()).GetLatest()
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if cfg != nil {
		view["config"] = cfg
	}
	return view, nil
}

// ListFunctions 合并运行时函数目录与管理侧治理状态。
func (s *Service) ListFunctions(ctx context.Context) ([]map[string]any, error) {
	govItems, err := store.NewTrendlnFunctionGovernanceStore(s.st.GormDB()).ListAll()
	if err != nil {
		return nil, err
	}
	catalog := make([]FunctionCatalogItem, 0)
	if s.client.enabled() {
		var runtimeItems []FunctionCatalogItem
		runtimeItems, err = s.client.Functions(ctx)
		if err != nil {
			logger.Warnf("读取 trendln runtime 函数目录失败，改为返回治理侧数据：%v", err)
		} else {
			catalog = runtimeItems
		}
	} else {
		logger.Infof("trendln runtime 未启用，函数目录仅返回治理侧数据")
	}
	govMap := make(map[string]store.TrendlnFunctionGovernance, len(govItems))
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

func defaultFunctionGovernance(item FunctionCatalogItem) store.TrendlnFunctionGovernance {
	return store.TrendlnFunctionGovernance{
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

func mergeFunctionGovernance(item FunctionCatalogItem, override store.TrendlnFunctionGovernance) store.TrendlnFunctionGovernance {
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

func mergeFunctionCatalog(item FunctionCatalogItem, gov store.TrendlnFunctionGovernance) map[string]any {
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

func mergeGovernanceOnly(gov store.TrendlnFunctionGovernance) map[string]any {
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
func (s *Service) UpdateRuntimeConfig(ctx context.Context, item *store.TrendlnRuntimeConfig) error {
	_ = ctx
	return store.NewTrendlnRuntimeConfigStore(s.st.GormDB()).Upsert(item)
}

// UpdateFunctionGovernance 保存函数治理信息。
func (s *Service) UpdateFunctionGovernance(ctx context.Context, item *store.TrendlnFunctionGovernance) error {
	_ = ctx
	return store.NewTrendlnFunctionGovernanceStore(s.st.GormDB()).Upsert(item)
}

// Trial 执行 trendln 试算并记录历史。
func (s *Service) Trial(ctx context.Context, req ComputeRequest, createdBy string) (map[string]any, error) {
	result, statusCode, err := s.client.Compute(ctx, req)
	if result == nil {
		result = &ComputeResponse{
			FunctionKey:   req.FunctionKey,
			Summary:       "",
			Parameters:    req.Parameters,
			SeriesMeta:    map[string]any{"fields": sortedSeriesFields(req.Series), "length": 0},
			SourceMeta:    map[string]any{"output_names": []string{"values"}, "render_pane": "sub", "warmup_bars": 0},
			Warnings:      []string{},
			Values:        []any{},
			RenderPayload: RenderPayload{MainSeries: []RenderSeries{}, SubSeries: []RenderSeries{}, Markers: []Marker{}, TrendLines: []TrendLine{}, Zones: []Zone{}, Events: []Event{}},
		}
	}
	history := &store.TrendlnTrialHistory{
		Symbol:            strings.TrimSpace(req.PriceSource),
		Timeframe:         "",
		FunctionKey:       req.FunctionKey,
		InputSummaryJSON:  mustJSON(req),
		ResultSummaryJSON: mustJSON(result),
		Status:            "ok",
		ErrorMessage:      "",
		CreatedBy:         strings.TrimSpace(createdBy),
	}
	if err != nil || statusCode >= 400 {
		history.Status = "failed"
		if err != nil {
			history.ErrorMessage = errString(err)
		} else {
			history.ErrorMessage = "trendln runtime 返回失败状态"
		}
	}
	if histErr := store.NewTrendlnTrialHistoryStore(s.st.GormDB()).Create(history); histErr != nil {
		logger.Warnf("写入 trendln 试算历史失败：%v", histErr)
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

func sortedSeriesFields(series map[string]any) []string {
	fields := make([]string, 0, len(series))
	for key := range series {
		fields = append(fields, strings.TrimSpace(key))
	}
	sort.Strings(fields)
	return fields
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
