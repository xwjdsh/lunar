package lunar

import (
	"testing"
)

var m = map[Date]LunarDate{
	NewDate(2021, 7, 20):  NewLunarDate(NewDate(2021, 6, 11), false),
	NewDate(1995, 1, 1):   NewLunarDate(NewDate(1994, 12, 1), false),
	NewDate(2020, 5, 12):  NewLunarDate(NewDate(2020, 4, 20), false),
	NewDate(2020, 6, 11):  NewLunarDate(NewDate(2020, 4, 20), true),
	NewDate(2088, 9, 19):  NewLunarDate(NewDate(2088, 8, 5), false),
	NewDate(2088, 12, 19): NewLunarDate(NewDate(2088, 11, 7), false),
}

func TestCalendar(t *testing.T) {
	for i := 0; i <= 1; i++ {
		for k, v := range m {
			{
				d, err := Calendar(k)
				if err != nil {
					t.Error(err)
				}
				if actual := d.LunarDate; actual != v {
					t.Errorf("DateToLunarDate error, expected: %s, actual: %s", v, actual)
				}
			}

			{
				d, err := Calendar(v)
				if err != nil {
					t.Error(err)
				}
				if actual := d.Date; actual != k {
					t.Errorf("LunarDateToDate error, expected: %s, actual: %s", k, actual)
				}
			}
		}
	}
}
