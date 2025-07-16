package hugoutil

import (
	"encoding/json"
	"fmt"
	"time"
)

type FrontmatterTime struct {
	layout string
	time   time.Time
}

func NewFrontmatterTime(layout string, t time.Time) FrontmatterTime {
	return FrontmatterTime{
		layout: layout,
		time:   t,
	}
}

func (t FrontmatterTime) Time() time.Time {
	return t.time
}

func (t *FrontmatterTime) UnmarshalJSON(data []byte) error {
	var str string

	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("unmarshal time: %v", err)
	}

	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02",
	} {
		parsedTime, err := time.Parse(layout, str)
		if err != nil {
			continue
		}

		*t = FrontmatterTime{
			layout: layout,
			time:   parsedTime,
		}

		return nil
	}

	return fmt.Errorf("unrecognized time format: %s", str)
}

func (t FrontmatterTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.time.Format(t.layout))
}
