package frontmatterparams

type Nav struct {
	Type      *Nav_Type      `json:"type,omitempty"`
	Place     *Nav_Place     `json:"place,omitempty"`
	PlacePark *Nav_PlacePark `json:"placePark,omitempty"`
	Tag       *Nav_Tag       `json:"tag,omitempty"`
}

type Nav_Type map[string]bool

type Nav_Place map[string]bool

type Nav_PlacePark map[string]bool

type Nav_Tag map[string]bool
