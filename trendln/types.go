package trendln

// FunctionCatalogItem 描述 trendln runtime 返回的结构分析能力目录项。
type FunctionCatalogItem struct {
	FunctionKey             string         `json:"function_key"`
	DisplayName             string         `json:"display_name"`
	GroupName               string         `json:"group_name"`
	DescriptionZH           string         `json:"description_zh"`
	FunctionType            string         `json:"function_type"`
	ParametersSchema        map[string]any `json:"parameters_schema"`
	InputSeriesRequirements []string       `json:"input_series_requirements"`
	OutputSchema            map[string]any `json:"output_schema"`
	RenderSchema            map[string]any `json:"render_schema"`
	WarmupBars              int            `json:"warmup_bars"`
	Tags                    []string       `json:"tags"`
}

// RuntimeHealthView 描述 trendln runtime 的健康检查结果。
type RuntimeHealthView struct {
	Status              string `json:"status"`
	Service             string `json:"service"`
	Version             string `json:"version"`
	RegisteredFunctions int    `json:"registered_functions"`
	StartedAt           string `json:"started_at"`
}

// ComputeRequest 是 trendln runtime 的统一计算输入。
type ComputeRequest struct {
	FunctionKey string         `json:"function_key"`
	Parameters  map[string]any `json:"parameters"`
	Series      map[string]any `json:"series"`
	PriceSource string         `json:"price_source"`
}

// RenderSeries 描述主图或副图中的单条序列。
type RenderSeries struct {
	Key    string    `json:"key"`
	Name   string    `json:"name"`
	Pane   string    `json:"pane"`
	Style  string    `json:"style"`
	Values []any     `json:"values"`
}

// Marker 描述主图上的极值或关键位标记。
type Marker struct {
	Time        int64   `json:"time"`
	Value       float64 `json:"value"`
	Label       string  `json:"label"`
	Color       string  `json:"color"`
	Side        string  `json:"side"`
	Description string  `json:"description"`
}

// Event 描述结构分析事件。
type Event struct {
	Time        int64  `json:"time"`
	Label       string `json:"label"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
}

// TrendLine 描述一条支撑线或阻力线。
type TrendLine struct {
	Key         string    `json:"key"`
	Name        string    `json:"name"`
	StartTime   int64     `json:"start_time"`
	EndTime     int64     `json:"end_time"`
	StartValue  float64   `json:"start_value"`
	EndValue    float64   `json:"end_value"`
	Color       string    `json:"color"`
	Style       string    `json:"style"`
	Description string    `json:"description"`
}

// Zone 描述水平位或价格区间。
type Zone struct {
	Key         string  `json:"key"`
	Name        string  `json:"name"`
	Top         float64 `json:"top"`
	Bottom      float64 `json:"bottom"`
	Color       string  `json:"color"`
	Description string  `json:"description"`
}

// RenderPayload 描述前端试算视图需要的渲染结构。
type RenderPayload struct {
	MainSeries []RenderSeries `json:"main_series"`
	SubSeries  []RenderSeries `json:"sub_series"`
	Markers    []Marker       `json:"markers"`
	TrendLines []TrendLine    `json:"trend_lines"`
	Zones      []Zone         `json:"zones"`
	Events     []Event        `json:"events"`
}

// ComputeResponse 是 trendln runtime 的统一计算输出。
type ComputeResponse struct {
	FunctionKey   string         `json:"function_key"`
	Summary       string         `json:"summary"`
	Parameters    map[string]any `json:"parameters"`
	SeriesMeta    map[string]any `json:"series_meta"`
	SourceMeta    map[string]any `json:"source_meta"`
	Warnings      []string       `json:"warnings"`
	Values        any            `json:"values"`
	RenderPayload RenderPayload  `json:"render_payload"`
}
