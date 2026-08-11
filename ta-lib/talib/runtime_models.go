package talib

type FunctionCatalogItem struct {
	FunctionKey             string         `json:"function_key"`
	DisplayName             string         `json:"display_name"`
	GroupName               string         `json:"group_name"`
	DescriptionZH           string         `json:"description_zh"`
	ParametersSchema        map[string]any `json:"parameters_schema"`
	InputSeriesRequirements []string       `json:"input_series_requirements"`
	OutputSchema            map[string]any `json:"output_schema"`
	WarmupBars              int            `json:"warmup_bars"`
	Tags                    []string       `json:"tags"`
}

type RuntimeHealthView struct {
	Status              string `json:"status"`
	Service             string `json:"service"`
	Version             string `json:"version"`
	RegisteredFunctions int    `json:"registered_functions"`
	StartedAt           string `json:"started_at"`
}

type ComputeRequest struct {
	FunctionKey string         `json:"function_key"`
	Parameters  map[string]any `json:"parameters"`
	Series      map[string]any `json:"series"`
	PriceSource string         `json:"price_source"`
}

type ComputeResponse struct {
	FunctionKey string         `json:"function_key"`
	Summary     string         `json:"summary"`
	Parameters  map[string]any `json:"parameters"`
	SeriesMeta  map[string]any `json:"series_meta"`
	Warnings    []string       `json:"warnings"`
	Values      any            `json:"values"`
	SourceMeta  map[string]any `json:"source_meta,omitempty"`
}
