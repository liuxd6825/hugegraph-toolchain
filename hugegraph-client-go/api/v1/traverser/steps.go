package traverser

type Direction string

const (
	Out  Direction = "OUT"
	In   Direction = "IN"
	Both Direction = "BOTH"
)

type Steps struct {
	Direction   Direction    `json:"direction,omitempty"`
	EdgeSteps   []EdgeStep   `json:"edge_steps,omitempty"`
	VertexSteps []VertexStep `json:"vertex_steps,omitempty"`
	MaxDegree   *int         `json:"max_degree,omitempty"`
	SkipDegree  *int         `json:"skip_degree,omitempty"`
}

type EdgeStep struct {
	Label      string         `json:"label,omitempty"`
	Properties map[string]any `json:"properties,omitempty"`
}
type VertexStep struct {
	Label      string         `json:"label,omitempty"`
	Properties map[string]any `json:"properties,omitempty"`
}

type Step struct {
	Direction  Direction      `json:"direction,omitempty"`
	Properties map[string]any `json:"properties,omitempty"`
}
