package lunar

import (
	"testing"
	"time"
)

func TestCalendar(t *testing.T) {
	m := map[string]string{
		"20210720": "20210611",
		"19950101": "19941201",
		"20220201": "20220101",
		"20880919": "20880805",
	}

	for k, v := range m {
		req, _ := time.Parse("20060102", k)
		d, err := Calendar(req)
		if err != nil {
			t.Error(err)
		}
		if actual := d.LunarDate.Format("20060102"); actual != v {
			t.Errorf("calendar error, expected: %s, actual: %s", v, actual)
		}
	}
}
