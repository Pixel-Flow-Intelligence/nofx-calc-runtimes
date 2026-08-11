package talib

import (
	"context"
	"encoding/json"
	"strings"

	"gorm.io/gorm"
	"nofx-core/shared/config"
	store "nofx-core/shared/db"
	"nofx-core/shared/logger"
	talibstore "nofx-ta-lib-runtime/talib/store"
)

type Service struct {
	st     *store.Store
	client *RuntimeClient
}

func NewService(st *store.Store) *Service {
	return &Service{
		st:     st,
		client: newRuntimeClientFromStore(st),
	}
}

func newRuntimeClientFromStore(st *store.Store) *RuntimeClient {
	envCfg := config.Get()
	baseURL := strings.TrimSpace(envCfg.TALibRuntimeBaseURL)
	timeoutMS := envCfg.TALibRuntimeTimeoutMS
	enabled := baseURL != ""
	if st != nil {
		cfg, err := talibstore.NewTaLibRuntimeConfigStore(st.GormDB()).GetLatest()
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
			logger.Warnf("读取 ta-lib runtime 数据库配置失败，回退到环境变量：%v", err)
		}
	}
	return NewRuntimeClientWithOptions(baseURL, timeoutMS, enabled)
}

// GetRuntimeView 返回 ta-lib-runtime 的健康态与管理配置，便于 admin 页面统一展示。
func (s *Service) GetRuntimeView(ctx context.Context) (map[string]any, error) {
	view := map[string]any{
		"enabled": false,
	}
	health, err := s.client.Health(ctx)
	if err != nil {
		view["error"] = err.Error()
	} else {
		view["enabled"] = true
		view["health"] = health
	}
	cfg, err := talibstore.NewTaLibRuntimeConfigStore(s.st.GormDB()).GetLatest()
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if cfg != nil {
		view["config"] = cfg
	}
	return view, nil
}

// ListFunctions 合并运行时函数目录与管理侧治理状态，供 admin 前端统一编辑。
func (s *Service) ListFunctions(ctx context.Context) ([]map[string]any, error) {
	govItems, err := talibstore.NewTaLibFunctionGovernanceStore(s.st.GormDB()).ListAll()
	if err != nil {
		return nil, err
	}
	catalog := make([]FunctionCatalogItem, 0)
	if s.client.enabled() {
		var runtimeItems []FunctionCatalogItem
		runtimeItems, err = s.client.Functions(ctx)
		if err != nil {
			logger.Warnf("读取 ta-lib runtime 函数目录失败，改为返回治理侧数据：%v", err)
		} else {
			catalog = runtimeItems
		}
	} else {
		// runtime 未启用时直接使用治理侧数据，避免把预期状态刷成告警噪音。
		logger.Infof("ta-lib runtime 未启用，函数目录仅返回治理侧数据")
	}
	govMap := make(map[string]talibstore.TaLibFunctionGovernance, len(govItems))
	for _, item := range govItems {
		govMap[item.FunctionKey] = item
	}

	out := make([]map[string]any, 0, max(len(catalog), len(govItems)))
	for _, item := range catalog {
		gov, ok := govMap[item.FunctionKey]
		if !ok {
			gov = talibstore.TaLibFunctionGovernance{
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
		out = append(out, map[string]any{
			"function_key":              item.FunctionKey,
			"display_name":              item.DisplayName,
			"group_name":                item.GroupName,
			"description_zh":            item.DescriptionZH,
			"parameters_schema":         item.ParametersSchema,
			"input_series_requirements": item.InputSeriesRequirements,
			"output_schema":             item.OutputSchema,
			"warmup_bars":               item.WarmupBars,
			"tags":                      item.Tags,
			"enabled":                   gov.Enabled,
			"agent_enabled":             gov.AgentEnabled,
			"admin_trial_enabled":       gov.AdminTrialEnabled,
			"default_params_json":       gov.DefaultParamsJSON,
			"risk_note":                 gov.RiskNote,
			"reviewed_by":               gov.ReviewedBy,
			"reviewed_at":               gov.ReviewedAt,
		})
	}
	for _, gov := range govItems {
		if _, exists := findFunctionByKey(catalog, gov.FunctionKey); exists {
			continue
		}
		out = append(out, map[string]any{
			"function_key":              gov.FunctionKey,
			"display_name":              gov.DisplayName,
			"group_name":                gov.GroupName,
			"description_zh":            gov.DescriptionZH,
			"parameters_schema":         map[string]any{},
			"input_series_requirements": []string{},
			"output_schema":             map[string]any{},
			"warmup_bars":               0,
			"tags":                      []string{},
			"enabled":                   gov.Enabled,
			"agent_enabled":             gov.AgentEnabled,
			"admin_trial_enabled":       gov.AdminTrialEnabled,
			"default_params_json":       gov.DefaultParamsJSON,
			"risk_note":                 gov.RiskNote,
			"reviewed_by":               gov.ReviewedBy,
			"reviewed_at":               gov.ReviewedAt,
		})
	}
	return out, nil
}

// UpdateRuntimeConfig 保存运行时接入配置，敏感信息继续由 env 兜底。
func (s *Service) UpdateRuntimeConfig(ctx context.Context, item *talibstore.TaLibRuntimeConfig) error {
	_ = ctx
	return talibstore.NewTaLibRuntimeConfigStore(s.st.GormDB()).Upsert(item)
}

// UpdateFunctionGovernance 保存函数治理开关和说明。
func (s *Service) UpdateFunctionGovernance(ctx context.Context, item *talibstore.TaLibFunctionGovernance) error {
	_ = ctx
	return talibstore.NewTaLibFunctionGovernanceStore(s.st.GormDB()).Upsert(item)
}

// Trial 先调用独立 runtime，再把试算结果写入历史，方便 admin 页面查看。
func (s *Service) Trial(ctx context.Context, req ComputeRequest, createdBy string) (map[string]any, error) {
	result, statusCode, err := s.client.Compute(ctx, req)
	history := &talibstore.TaLibTrialHistory{
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
		history.ErrorMessage = errString(err)
	}
	if histErr := talibstore.NewTaLibTrialHistoryStore(s.st.GormDB()).Create(history); histErr != nil {
		logger.Warnf("写入 ta-lib 试算历史失败：%v", histErr)
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
