package main

import (
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

func main() {
	contentPaths, err := filepath.Glob("content/photo/*/*.md")
	if err != nil {
		panic(err)
	}

	for _, contentPath := range contentPaths {
		fmt.Println(contentPath)

		contentBytes, err := ioutil.ReadFile(contentPath)
		if err != nil {
			panic(err)
		}

		contentParts := strings.SplitN(string(contentBytes), "---\n", 3)

		var frontmatter map[string]interface{}

		err = yaml.Unmarshal([]byte(contentParts[1]), &frontmatter)
		if err != nil {
			panic(err)
		}

		galleriesValue, ok := frontmatter["galleries"]
		if !ok {
			fmt.Println("> no galleries")
			continue
		}

		galleries, ok := galleriesValue.([]interface{})
		if !ok {
			fmt.Println("> no galleries array")
			continue
		}

		contentParts[1] = fmt.Sprintf("%saliases:\n- /gallery/%s/%s\n", contentParts[1], galleries[0].(string), strings.Replace(path.Base(contentPath), ".md", ".html", 1))

		err = ioutil.WriteFile(contentPath, []byte(strings.Join(contentParts, "---\n")), 0644)
		if err != nil {
			panic(err)
		}
	}
}
