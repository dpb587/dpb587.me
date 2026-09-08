package content

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/dpb587/dpb587.me/tools/content/frontmatterparams"
	"github.com/dpb587/dpb587.me/tools/content/hugoutil"
	"go.yaml.in/yaml/v4"
)

type Document struct {
	Frontmatter *Content_Frontmatter
	Body        []byte
}

func (c *Document) ReadFrom(r io.Reader) (int64, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return int64(len(buf)), fmt.Errorf("read content: %v", err)
	}

	parts := bytes.SplitN(buf, []byte("---\n"), 3)
	if len(parts) > 1 && len(parts[0]) == 0 {
		var remarshal map[string]any

		err = yaml.Unmarshal(parts[1], &remarshal)
		if err != nil {
			return int64(len(buf)), fmt.Errorf("unmarshal frontmatter: %v", err)
		}

		remarshalBytes, err := json.Marshal(remarshal)
		if err != nil {
			return int64(len(buf)), fmt.Errorf("marshal frontmatter to json: %v", err)
		}

		c.Frontmatter = &Content_Frontmatter{}

		err = json.Unmarshal(remarshalBytes, &c.Frontmatter)
		if err != nil {
			return int64(len(buf)), fmt.Errorf("unmarshal frontmatter to struct: %v", err)
		}

		if len(parts) > 2 {
			c.Body = parts[2]
		}
	} else {
		c.Body = buf
	}

	return int64(len(buf)), nil
}

func (c Document) WriteTo(w io.Writer) (int64, error) {
	buf, err := json.Marshal(c.Frontmatter)
	if err != nil {
		return 0, fmt.Errorf("marshal frontmatter: %v", err)
	}

	if len(buf) > 4 {
		var remarshal map[string]any

		err = json.Unmarshal(buf, &remarshal)
		if err != nil {
			return 0, fmt.Errorf("unmarshal frontmatter: %v", err)
		}

		_, err = w.Write([]byte("---\n"))
		if err != nil {
			return 0, fmt.Errorf("write separator: %v", err)
		}

		frontmatterBytes := &bytes.Buffer{}

		ym, err := yaml.NewDumper(frontmatterBytes)
		if err != nil {
			return 0, fmt.Errorf("create yaml dumper: %v", err)
		}

		err = ym.Dump(remarshal)
		if err != nil {
			return 0, fmt.Errorf("write frontmatter: %v", err)
		}

		_, err = w.Write(frontmatterBytes.Bytes())
		if err != nil {
			return 0, fmt.Errorf("write frontmatter: %v", err)
		}

		_, err = w.Write([]byte("---\n"))
		if err != nil {
			return 0, fmt.Errorf("write separator: %v", err)
		}

		if len(c.Body) > 0 {
			_, err = w.Write([]byte("\n"))
			if err != nil {
				return 0, fmt.Errorf("write newline before body: %v", err)
			}
		}
	}

	if len(c.Body) > 0 {
		_, err = w.Write(c.Body)
		if err != nil {
			return 0, fmt.Errorf("write body: %v", err)
		}
	}

	return 0, nil
}

//

type Content_Frontmatter struct {
	Date        *hugoutil.FrontmatterTime   `json:"date,omitempty"`
	Layout      *string                     `json:"layout,omitempty"`
	Params      *Content_Frontmatter_Params `json:"params,omitempty"`
	PublishDate *hugoutil.FrontmatterTime   `json:"publishDate,omitempty"`
	Slug        *string                     `json:"slug,omitempty"`
	Title       *string                     `json:"title,omitempty"`
	Type        *string                     `json:"type,omitempty"`
}

//

type Content_Frontmatter_Params struct {
	LinkType  *frontmatterparams.LinkType  `json:"linkType,omitempty"`
	MediaType *frontmatterparams.MediaType `json:"mediaType,omitempty"`
	RouteType *frontmatterparams.RouteType `json:"routeType,omitempty"`
	TimeRange *frontmatterparams.TimeRange `json:"timeRange,omitempty"`
	Topics    *frontmatterparams.Topics    `json:"topics,omitempty"`
}

func (p *Content_Frontmatter_Params) SetTopic(k string, v frontmatterparams.TopicParam) {
	if p.Topics == nil {
		p.Topics = &frontmatterparams.Topics{}
	}

	(*p.Topics)[k] = v
}

func (p *Content_Frontmatter_Params) SetTopicPlace(k string, v frontmatterparams.TopicParam) {
	p.SetTopic("places/"+k, v)
}
