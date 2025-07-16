package frontmatterparams

import "github.com/dpb587/dpb587.me/tools/content/hugoutil"

type TimeRange struct {
	From *hugoutil.FrontmatterTime `json:"from,omitempty"`
	Thru *hugoutil.FrontmatterTime `json:"thru,omitempty"`
}
