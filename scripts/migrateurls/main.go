package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

// https://remark42.com/docs/backup/url-migration/

var reMetaRefresh = regexp.MustCompile(`<meta http-equiv="refresh" content="0; url=([^"]+)" />`)

func main() {
	s := bufio.NewScanner(os.Stdin)

	for s.Scan() {
		oldurl := s.Text()

		err := func() error {
			newurl, err := url.Parse(strings.ReplaceAll(oldurl, "http://", "https://"))
			if err != nil {
				return fmt.Errorf("parsing url: %v", err)
			}

			localurl := *newurl
			localurl.Scheme = "http"
			localurl.Host = "localhost:1313"

			res, err := http.Get(localurl.String())
			if err != nil {
				return fmt.Errorf("getting url: %v", err)
			} else if res.StatusCode != http.StatusOK {
				return fmt.Errorf("unexpected status code: %d", res.StatusCode)
			}

			buf, err := ioutil.ReadAll(res.Body)
			if err != nil {
				return fmt.Errorf("reading response body: %v", err)
			}

			if m := reMetaRefresh.FindStringSubmatch(string(buf)); len(m) > 0 {
				parsedRedirect, err := url.Parse(m[1])
				if err != nil {
					return fmt.Errorf("parsing meta redirect: %v", err)
				}

				nexturl := newurl.ResolveReference(parsedRedirect)
				nexturl.Scheme = newurl.Scheme
				nexturl.Host = newurl.Host
				newurl = nexturl
			}

			newurlString := newurl.String()

			if oldurl == newurlString {
				return nil
			}

			fmt.Printf("%s %s\n", oldurl, newurlString)

			return nil
		}()
		if err != nil {
			fmt.Fprintf(os.Stderr, "# error (%s): %s\n", oldurl, err)
		}
	}
}
