package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

func main() {
	fromdir := os.Args[1]
	todir := os.Args[2]

	idx, err := read(filepath.Join(fromdir, "_index.md"))
	if err != nil {
		panic(errors.Wrap(err, "reading _index.md"))
	}

	err = write(filepath.Join(todir, "_index.md"), map[interface{}]interface{}{
		"title":           idx["name"],
		"date":            strings.SplitN(idx["temporalCoverage"].(string), "/", 2)[0],
		"date_end":        strings.SplitN(idx["temporalCoverage"].(string), "/", 2)[1],
		"highlight_photo": strings.TrimSuffix(path.Base(idx["image"].(map[interface{}]interface{})["@id"].(string)), ".md"),
	})
	if err != nil {
		panic(errors.Wrap(err, "writing _index.md"))
	}

	for itemIdx, item := range idx["itemListElement"].([]interface{}) {
		item := item.(map[interface{}]interface{})

		itemName := path.Base(item["@id"].(string))
		itm, err := read(filepath.Join(fromdir, itemName))
		if err != nil {
			panic(errors.Wrapf(err, "reading %s", itemName))
		}

		itmOutExif := map[string]interface{}{}

		for _, prop := range itm["associatedMedia"].(map[interface{}]interface{})["exifData"].([]interface{}) {
			prop := prop.(map[interface{}]interface{})
			itmOutExif[prop["identifier"].(string)] = prop["value"]
		}

		itmOutSizes := map[string]map[string]int{}

		for _, tn := range itm["associatedMedia"].(map[interface{}]interface{})["thumbnail"].([]interface{}) {
			tn := tn.(map[interface{}]interface{})
			itmOutSizes[strings.SplitN(strings.TrimSuffix(tn["contentUrl"].(string), ".jpg"), "~", 2)[1]] = map[string]int{
				"height": tn["height"].(int),
				"width":  tn["width"].(int),
			}
		}

		err = write(filepath.Join(todir, itemName), map[interface{}]interface{}{
			"date":     itm["associatedMedia"].(map[interface{}]interface{})["dateCreated"],
			"exif":     itmOutExif,
			"ordering": itemIdx,
			"title":    itm["associatedMedia"].(map[interface{}]interface{})["name"],
			"sizes":    itmOutSizes,
		})
		if err != nil {
			panic(errors.Wrapf(err, "writing %s", itemName))
		}
	}
}

func write(path string, dat interface{}) error {
	buf, err := yaml.Marshal(dat)
	if err != nil {
		return errors.Wrap(err, "marshalling")
	}

	err = ioutil.WriteFile(path, []byte(fmt.Sprintf("---\n%s---\n", buf)), 0755)
	if err != nil {
		return errors.Wrap(err, "writing")
	}

	return nil
}

func read(path string) (map[interface{}]interface{}, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "reading")
	}

	pieces := strings.SplitN(string(buf), "---\n", 3)

	var dat map[interface{}]interface{}

	err = yaml.Unmarshal([]byte(pieces[1]), &dat)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshalling")
	}

	return dat, nil
}
