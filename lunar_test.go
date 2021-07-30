package lunar

import (
	"testing"
)

var m = map[Date]Date{
	{2021, 7, 20}:  {2021, 6, 11},
	{1995, 1, 1}:   {1994, 12, 1},
	{2022, 2, 1}:   {2022, 1, 1},
	{2088, 9, 19}:  {2088, 8, 5},
	{2088, 12, 19}: {2088, 11, 7},
}

func TestDateToLunarDate(t *testing.T) {
	for i := 0; i <= 1; i++ {
		for k, v := range m {
			d, err := DateToLunarDate(k)
			if err != nil {
				t.Error(err)
			}
			if actual := d.LunarDate; actual != v {
				t.Errorf("DateToLunarDate error, expected: %s, actual: %s", v, actual)
			}
		}
	}
}

func BenchmarkDateToLunarDate(b *testing.B) {
	h := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for k, v := range m {
			r, _ := h.DateToLunarDate(k)
			if r.LunarDate != v {
				b.FailNow()
			}
		}
	}
}

func TestLunarDateToDate(t *testing.T) {
	for k, v := range m {
		d, err := LunarDateToDate(v)
		if err != nil {
			t.Error(err)
		}
		if actual := d.Date; actual != k {
			t.Errorf("LunarDateToDate error, expected: %s, actual: %s", k, actual)
		}
	}
}

func BenchmarkLunarDateToDate(b *testing.B) {
	h := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for k, v := range m {
			r, _ := h.LunarDateToDate(v)
			if r.Date != k {
				b.FailNow()
			}
		}
	}
}
