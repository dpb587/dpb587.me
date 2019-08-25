package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

// mkdir -p ~/Projects/src/github.com/dpb587/tsg/internal/dpb/assets/content/galleries
// cp -rp content/post ~/Projects/src/github.com/dpb587/tsg/internal/dpb/assets/content/posts
// cp -rp content/photo/* ~/Projects/src/github.com/dpb587/tsg/internal/dpb/assets/content/galleries/
// for g in $( cd content/galleries ; ls ); do p=content/galleries/$g/_index.md ; [ ! -e "$p" ] || cp $p ~/Projects/src/github.com/dpb587/tsg/internal/dpb/assets/content/galleries/$g/ ; done
// rm ~/Projects/src/github.com/dpb587/tsg/internal/dpb/assets/content/posts/_index.md

func main() {
	repoDir := os.Args[1]

	files, err := filepath.Glob(path.Join(repoDir, "content/galleries/*/_index.md"))
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		err = migrateGallery(repoDir, file)
		if err != nil {
			panic(err)
		}
	}

	err = filepath.Walk(path.Join(repoDir, "content/photo"), func(p string, i os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrapf(err, "walking %s", p)
		} else if i.IsDir() {
			return nil
		}

		return migratePhoto(p)
	})
	if err != nil {
		panic(err)
	}

	err = filepath.Walk(path.Join(repoDir, "content/post"), func(p string, i os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrapf(err, "walking %s", p)
		} else if i.IsDir() {
			return nil
		}

		return migratePost(p)
	})
	if err != nil {
		panic(err)
	}
}

func migrateGallery(repoDir string, p string) error {
	fmt.Fprintf(os.Stderr, "GALLERY: %s\n", p)

	b, err := ioutil.ReadFile(p)
	if err != nil {
		return errors.Wrap(err, "reading file")
	}

	sp := strings.SplitN(string(b), "---\n", 3)

	var frontmatter map[string]interface{}

	err = yaml.Unmarshal([]byte(sp[1]), &frontmatter)
	if err != nil {
		return errors.Wrap(err, "unmarshalling frontmatter")
	}

	galleryName := path.Base(path.Dir(p))

	newfrontmatter := map[string]interface{}{
		"@context":       "http://schema.org",
		"@type":          "Collection",
		"additionalType": "http://schema.org/ItemList",
	}

	for k, v := range frontmatter {
		if k == "title" {
			k = "name"
		} else if k == "date" || k == "date_end" {
			continue
		} else if k == "slug" {
			continue
		} else if k == "aliases" {
			k = "url"
		} else if k == "highlight_photo" {
			k = "image"
			v = map[string]interface{}{
				"@id": fmt.Sprintf("https://dpb587.github.io/dpb587.me/content/galleries/%s/%s.md", galleryName, v),
			}
		}

		newfrontmatter[k] = v
	}

	{
		var itemListElement []map[string]string

		var stdout bytes.Buffer
		cmd := exec.Command("bash", "-c", fmt.Sprintf("cd %s ; grep ordering * | sort -nk2", path.Join(repoDir, "content/photo", galleryName)))
		cmd.Stdout = &stdout

		err := cmd.Run()
		if err != nil {
			return errors.Wrap(err, "loading photo order")
		}

		for _, itemline := range strings.Split(stdout.String(), "\n") {
			sp := strings.SplitN(itemline, ":", 2)

			if len(sp[0]) == 0 {
				continue
			}

			itemListElement = append(
				itemListElement,
				map[string]string{
					"@id": fmt.Sprintf("https://dpb587.github.io/dpb587.me/content/galleries/%s/%s", galleryName, sp[0]),
				},
			)
		}

		newfrontmatter["itemListElement"] = itemListElement
	}

	b, err = yaml.Marshal(newfrontmatter)
	if err != nil {
		return errors.Wrap(err, "marshalling new frontmatter")
	}

	sp[1] = string(b)

	err = ioutil.WriteFile(p, []byte(strings.Join(sp, "---\n")), 0755)
	if err != nil {
		return errors.Wrap(err, "writing file")
	}

	return nil
}

func migratePost(p string) error {
	fmt.Fprintf(os.Stderr, "POST: %s\n", p)

	b, err := ioutil.ReadFile(p)
	if err != nil {
		return errors.Wrap(err, "reading file")
	}

	sp := strings.SplitN(string(b), "---\n", 3)

	var frontmatter map[string]interface{}

	err = yaml.Unmarshal([]byte(sp[1]), &frontmatter)
	if err != nil {
		return errors.Wrap(err, "unmarshalling frontmatter")
	}

	newfrontmatter := map[string]interface{}{
		"@context": "http://schema.org",
		"@type":    "BlogPosting",
	}

	for k, v := range frontmatter {
		if k == "title" {
			k = "name"
		} else if k == "date" {
			k = "datePublished"
		} else if k == "tags" {
			k = "keywords"
		} else if k == "aliases" {
			k = "url"
		} else if k == "code" {
			k = "mentions"
		} else if k == "primary_image" {
			k = "image"
		}

		newfrontmatter[k] = v
	}

	b, err = yaml.Marshal(newfrontmatter)
	if err != nil {
		return errors.Wrap(err, "marshalling new frontmatter")
	}

	sp[1] = string(b)

	err = ioutil.WriteFile(p, []byte(strings.Join(sp, "---\n")), 0755)
	if err != nil {
		return errors.Wrap(err, "writing file")
	}

	return nil
}

func migratePhoto(p string) error {
	fmt.Fprintf(os.Stderr, "PHOTO: %s\n", p)

	b, err := ioutil.ReadFile(p)
	if err != nil {
		return errors.Wrap(err, "reading file")
	}

	sp := strings.SplitN(string(b), "---\n", 3)

	var frontmatter map[string]interface{}

	err = yaml.Unmarshal([]byte(sp[1]), &frontmatter)
	if err != nil {
		return errors.Wrap(err, "unmarshalling frontmatter")
	}

	dateCreated, err := time.Parse("2006-01-02 15:04:05", frontmatter["date"].(string))
	if err != nil {
		return errors.Wrap(err, "parsing date")
	}

	newfrontmatter := map[string]interface{}{
		"@context": "http://schema.org",
		"@type":    "Photograph",
		"associatedMedia": map[string]interface{}{
			"@type":       "ImageObject",
			"dateCreated": dateCreated.Format(time.RFC3339),
		},
	}

	if strings.HasPrefix(frontmatter["title"].(string), "IMG") {
		newfrontmatter["associatedMedia"].(map[string]interface{})["name"] = frontmatter["title"].(string)
	} else {
		newfrontmatter["name"] = frontmatter["title"].(string)
	}

	if location, ok := frontmatter["location"].(map[interface{}]interface{}); ok {
		newlocationgeo := map[string]interface{}{
			"@type": "GeoCoordinates",
		}

		if v, ok := location["latitude"]; ok {
			newlocationgeo["latitude"] = v.(float64)
		}

		if v, ok := location["longitude"]; ok {
			newlocationgeo["longitude"] = v.(float64)
		}

		if len(newlocationgeo) > 1 {
			newfrontmatter["locationCreated"] = map[string]interface{}{
				"@type": "Place",
				"geo":   newlocationgeo,
			}
		}
	}

	if exif, ok := frontmatter["exif"].(map[interface{}]interface{}); ok {
		newexif := []map[string]interface{}{}

		if v, ok := exif["aperture"]; ok && v != nil {
			newexif = append(
				newexif,
				map[string]interface{}{
					"@type":      "PropertyValue",
					"identifier": "aperture",
					"value":      v.(string),
				},
			)
		}

		if v, ok := exif["exposure"]; ok && v != nil {
			newexif = append(
				newexif,
				map[string]interface{}{
					"@type":      "PropertyValue",
					"identifier": "exposure",
					"value":      v.(string),
				},
			)
		}

		if v, ok := exif["make"]; ok && v != nil {
			newexif = append(
				newexif,
				map[string]interface{}{
					"@type":      "PropertyValue",
					"identifier": "make",
					"value":      v.(string),
				},
			)
		}

		if v, ok := exif["model"]; ok && v != nil {
			newexif = append(
				newexif,
				map[string]interface{}{
					"@type":      "PropertyValue",
					"identifier": "model",
					"value":      v.(string),
				},
			)
		}

		if len(exif) > 0 {
			newfrontmatter["associatedMedia"].(map[string]interface{})["exifData"] = newexif
		}
	}

	galleryName := frontmatter["galleries"].([]interface{})[0].(string)
	baseFileName := strings.TrimSuffix(path.Base(p), ".md")

	{
		newthumbnail := []map[string]interface{}{}

		for tname, tsize := range frontmatter["sizes"].(map[interface{}]interface{}) {
			var tnamestr string

			if v, ok := tname.(int); ok {
				tnamestr = fmt.Sprintf("%d", v)
			} else {
				tnamestr = tname.(string)
			}

			newthumbnail = append(
				newthumbnail,
				map[string]interface{}{
					"@type":          "ImageObject",
					"contentUrl":     fmt.Sprintf("https://dpb587-website-us-east-1.s3.amazonaws.com/asset/gallery/%s/%s~%s.jpg", galleryName, baseFileName, tnamestr),
					"encodingFormat": "image/jpeg",
					"height":         tsize.(map[interface{}]interface{})["height"].(int),
					"width":          tsize.(map[interface{}]interface{})["width"].(int),
				},
			)
		}

		newfrontmatter["associatedMedia"].(map[string]interface{})["thumbnail"] = newthumbnail
	}

	if v, ok := frontmatter["aliases"].([]interface{}); ok {
		if len(v) > 0 {
			newfrontmatter["url"] = v
		}
	}

	b, err = yaml.Marshal(newfrontmatter)
	if err != nil {
		return errors.Wrap(err, "marshalling new frontmatter")
	}

	sp[1] = string(b)

	err = ioutil.WriteFile(p, []byte(strings.Join(sp, "---\n")), 0755)
	if err != nil {
		return errors.Wrap(err, "writing file")
	}

	return nil
}
