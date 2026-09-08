package frontmatterparams

type Topics map[string]TopicParam

type TopicParam struct {
	Rel  string         `json:"rel,omitempty"`
	Meta map[string]any `json:"meta,omitempty"`
}
