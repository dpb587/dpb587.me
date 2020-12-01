package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

func main() {
	if err := mainErr(); err != nil {
		panic(err)
	}
}

func mainErr() error {
	contentdir := "../../content/photo"

	err := filepath.Walk(contentdir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrapf(err, "walking %s", path)
		} else if info.IsDir() {
			return nil
		} else if !strings.HasSuffix(path, ".md") {
			return nil
		} else if path == "_index.md" {
			return nil
		}

		dat, err := read(path)
		if err != nil {
			return errors.Wrapf(err, "reading %s", path)
		}

		if _, ok := dat["image"]; ok {
			// already migrated
			return nil
		}

		sizes, ok := dat["sizes"].(map[interface{}]interface{})
		if !ok {
			return nil
		}

		var images []map[string]interface{}

		for _, key := range []interface{}{1920, "1920", 1280, "1280"} {
			sizeU, ok := sizes[key]
			if !ok {
				continue
			}

			size, ok := sizeU.(map[interface{}]interface{})
			if !ok {
				return fmt.Errorf("unexpected size type: %T", sizeU)
			}

			image := map[string]interface{}{}

			if v, ok := size["source"]; ok {
				image["url"] = fmt.Sprintf("https://s3.dualstack.us-east-1.amazonaws.com/dpb587-website-us-east-1/%s", v)
			} else {
				var urlkey string

				switch keyT := key.(type) {
				case int:
					urlkey = strconv.Itoa(keyT)
				case string:
					urlkey = keyT
				default:
					panic(fmt.Sprintf("unexpected key type: %T", key))
				}

				image["url"] = fmt.Sprintf("%s%s~%s.jpg", "https://s3.dualstack.us-east-1.amazonaws.com/dpb587-website-us-east-1/asset/gallery", strings.TrimSuffix(strings.TrimPrefix(path, contentdir), ".md"), urlkey)
			}

			if v, ok := size["width"]; ok {
				if vT, ok := v.(string); ok {
					v, err = strconv.Atoi(vT)
					if err != nil {
						panic(fmt.Sprintf("atoi: %v", v))
					}
				}

				image["width"] = v
			}

			if v, ok := size["height"]; ok {
				if vT, ok := v.(string); ok {
					v, err = strconv.Atoi(vT)
					if err != nil {
						panic(fmt.Sprintf("atoi: %v", v))
					}
				}

				image["height"] = v
			}

			if v, ok := size["integrity"]; ok {
				ids := []string{}

				for _, vv := range v.([]interface{}) {
					vvSplit := strings.SplitN(vv.(string), "-", 2)
					if vvSplit[0] != "sha512" {
						panic(fmt.Sprintf("unexpected algo: %s", vvSplit[0]))
					}

					vvBuf, err := base64.StdEncoding.DecodeString(vvSplit[1])
					if err != nil {
						panic(fmt.Sprintf("bad b64 decode: %s", err))
					}

					ids = append(ids, fmt.Sprintf("ni://sha-512;%s", base64.RawURLEncoding.EncodeToString(vvBuf)))
				}

				image["id"] = ids
			}

			images = append(images, image)

			break
		}

		delete(dat, "sizes")

		if len(images) == 0 {
			// no change
			return nil
		} else if len(images) != 1 {
			panic(fmt.Sprintf("multiple images for %s", path))
		}

		dat["image"] = images[0]

		err = write(path, dat)
		if err != nil {
			return errors.Wrap(err, "writing")
		}

		return nil
	})
	if err != nil {
		return errors.Wrap(err, "walking")
	}

	return nil
}

func write(path string, dat map[string]interface{}) error {
	content := dat["_content"].(string)
	delete(dat, "_content")

	buf, err := yaml.Marshal(dat)
	if err != nil {
		return errors.Wrap(err, "marshalling")
	}

	err = ioutil.WriteFile(path, []byte(strings.TrimSuffix(fmt.Sprintf("---\n%s---\n\n%s\n", buf, strings.TrimSpace(content)), "\n\n")), 0755)
	if err != nil {
		return errors.Wrap(err, "writing")
	}

	return nil
}

func read(path string) (map[string]interface{}, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "reading")
	}

	pieces := strings.SplitN(string(buf), "---\n", 3)

	var dat map[string]interface{}

	err = yaml.Unmarshal([]byte(pieces[1]), &dat)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshalling")
	}

	dat["_content"] = pieces[2]

	return dat, nil
}
