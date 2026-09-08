package frontmatterparams

type LinkType struct {
	Referrer string `json:"referrer,omitempty"`
	Target   string `json:"target,omitempty"`

	Metadata                    *LinkType_Metadata `json:"metadata,omitempty"`
	TargetResource              *LinkType_Resource `json:"targetResource,omitempty"`
	TargetOriginResource        *LinkType_Resource `json:"targetOriginResource,omitempty"`
	TargetOriginFaviconResource *LinkType_Resource `json:"targetOriginFaviconResource,omitempty"`
	Diagnostics                 []string           `json:"diagnostics,omitempty"`
}

type LinkType_Metadata struct {
	Name                    string                  `json:"name,omitempty"`
	Description             string                  `json:"description,omitempty"`
	FeaturedImageThumbnails MediaType_ThumbnailList `json:"featuredImageThumbnails,omitempty"`
	Author                  *LinkType_Author        `json:"author,omitempty"`
	Origin                  *LinkType_Origin        `json:"origin,omitempty"`
}

type LinkType_Author struct {
	Name           string                  `json:"name,omitempty"`
	IconThumbnails MediaType_ThumbnailList `json:"iconThumbnails,omitempty"`
}

type LinkType_Origin struct {
	Name           string                  `json:"name,omitempty"`
	IconThumbnails MediaType_ThumbnailList `json:"iconThumbnails,omitempty"`
}

type LinkType_Resource struct {
	Time      string             `json:"time,omitempty"`
	URL       string             `json:"url,omitempty"`
	LandedURL string             `json:"landedUrl,omitempty"`
	MediaType string             `json:"mediaType,omitempty"`
	Size      int64              `json:"size,omitempty"`
	Digest    string             `json:"digest,omitempty"`
	Error     string             `json:"error,omitempty"`
	Profiles  []LinkType_Profile `json:"profiles,omitempty"`
}

type LinkType_Profile struct {
	Kind  string `json:"kind,omitempty"`
	Data  string `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}
