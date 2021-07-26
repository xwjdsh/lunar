package lunar

import (
	"testing"
	"time"
)

var m = map[string]string{
	"20210720": "20210611",
	"19950101": "19941201",
	"20220201": "20220101",
	"20880919": "20880805",
}

func TestDateToLunarDate(t *testing.T) {
	for k, v := range m {
		req, _ := time.Parse("20060102", k)
		d, err := DateToLunarDate(DateByTime(req))
		if err != nil {
			t.Error(err)
		}
		if actual := d.LunarDate.Time().Format("20060102"); actual != v {
			t.Errorf("DateToLunarDate error, expected: %s, actual: %s", v, actual)
		}
	}
}

func TestLunarDateToDate(t *testing.T) {
	for k, v := range m {
		req, _ := time.Parse("20060102", v)
		d, err := LunarDateToDate(DateByTime(req))
		if err != nil {
			t.Error(err)
		}
		if actual := d.Date.Time().Format("20060102"); actual != k {
			t.Errorf("LunarDateToDate error, expected: %s, actual: %s", k, actual)
		}
	}
}
