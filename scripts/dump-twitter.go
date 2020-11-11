package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

func main() {
	config := oauth1.NewConfig(os.Getenv("TWITTER_CONSUMER_KEY"), os.Getenv("TWITTER_CONSUMER_SECRET_KEY"))
	token := oauth1.NewToken(os.Getenv("TWITTER_ACCESS_KEY"), os.Getenv("TWITTER_ACCESS_SECRET_KEY"))
	httpClient := config.Client(oauth1.NoContext, token)

	client := twitter.NewClient(httpClient)

	// Home Timeline
	tweets, _, err := client.Timelines.UserTimeline(&twitter.UserTimelineParams{
		ScreenName: "dpb587",
		Count:      1000,
	})
	if err != nil {
		panic(errors.Wrap(err, "getting tweets"))
	}

	for tweetIdx, tweet := range tweets {
		fmt.Printf("%d\n", tweet.ID)

		tweetPath := fmt.Sprintf("content/tweet/%d.md", tweet.ID)

		_, err := os.Stat(tweetPath)
		if err != nil {
			if os.IsNotExist(err) {
				// okay
				// continue
			} else {
				panic(errors.Wrap(err, "checking tweet.md"))
			}
		}

		if tweet.InReplyToStatusID > 0 {
			continue
		}

		buf, err := yaml.Marshal(map[string]interface{}{
			"date":    tweet.CreatedAt,
			"api_1_1": tweet,
		})
		if err != nil {
			panic(errors.Wrapf(err, "marshaling tweet %d", tweetIdx))
		}

		text := tweet.Text

		for _, entity := range tweet.Entities.Media {
			text = strings.Replace(text, entity.URLEntity.URL, " ", -1)
		}

		tweetContent := fmt.Sprintf("---\n%s---\n\n%s\n", buf, strings.TrimSpace(text))

		err = ioutil.WriteFile(tweetPath, []byte(tweetContent), 0755)
		if err != nil {
			panic(errors.Wrapf(err, "writing tweet %d", tweetIdx))
		}
	}
}
