package pandas_ta

// FunctionCatalogItem 描述 pandas-ta runtime 返回的函数目录项。
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

// ComputeRequest 是 pandas-ta runtime 的统一计算输入。
type ComputeRequest struct {
	FunctionKey string         `json:"function_key"`
	Parameters  map[string]any `json:"parameters"`
	Series      map[string]any `json:"series"`
	PriceSource string         `json:"price_source"`
}

// RenderSeries 描述副图或主图中的单条序列。
type RenderSeries struct {
	Key    string    `json:"key"`
	Name   string    `json:"name"`
	Pane   string    `json:"pane"`
	Style  string    `json:"style"`
	Values []any     `json:"values"`
}

// Marker 描述主图标记点。
type Marker struct {
	Time        int64   `json:"time"`
	Value       float64 `json:"value"`
	Label       string  `json:"label"`
	Color       string  `json:"color"`
	Side        string  `json:"side"`
	Description string  `json:"description"`
}

// Event 描述事件类输出。
type Event struct {
	Time        int64  `json:"time"`
	Label       string `json:"label"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
}

// RenderPayload 描述前端试算视图需要的渲染结构。
type RenderPayload struct {
	MainSeries []RenderSeries `json:"main_series"`
	SubSeries  []RenderSeries `json:"sub_series"`
	Markers    []Marker       `json:"markers"`
	Events     []Event        `json:"events"`
}

// ComputeResponse 是 pandas-ta runtime 的统一计算输出。
type ComputeResponse struct {
	FunctionKey  string         `json:"function_key"`
	Summary      string         `json:"summary"`
	Parameters   map[string]any `json:"parameters"`
	SeriesMeta   map[string]any `json:"series_meta"`
	SourceMeta   map[string]any `json:"source_meta"`
	Warnings     []string       `json:"warnings"`
	Values       any            `json:"values"`
	RenderPayload RenderPayload  `json:"render_payload"`
}
