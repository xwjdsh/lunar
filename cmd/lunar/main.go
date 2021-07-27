package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/xwjdsh/lunar"
)

var (
	CST         = time.FixedZone("CST", 3600*8)
	format      = flag.String("f", "2006-01-02", "time format")
	year        = flag.Int("y", 0, "year")
	reverseMode = flag.Bool("r", false, "reverse mode, find date by lunar date")
)

type alias struct {
	d           lunar.Date
	isLunarDate bool
	isHoliday   bool
}

func newAlias(d lunar.Date, isLunarDate, isHoliday bool) *alias {
	return &alias{
		d:           d,
		isLunarDate: isLunarDate,
		isHoliday:   isHoliday,
	}
}

var aliasMap = map[string]*alias{
	"春节": newAlias(lunar.NewDate(0, 1, 1), true, true),
	"元旦": newAlias(lunar.NewDate(0, 1, 1), false, true),
	"元宵": newAlias(lunar.NewDate(0, 1, 15), true, false),
	"清明": newAlias(lunar.NewDate(0, 4, 4), false, true),
	"劳动": newAlias(lunar.NewDate(0, 5, 4), false, true),
	"端午": newAlias(lunar.NewDate(0, 5, 5), true, true),
	"七夕": newAlias(lunar.NewDate(0, 7, 7), true, false),
	"中元": newAlias(lunar.NewDate(0, 7, 15), true, false),
	"中秋": newAlias(lunar.NewDate(0, 8, 15), true, true),
	"重阳": newAlias(lunar.NewDate(0, 9, 9), true, false),
	"国庆": newAlias(lunar.NewDate(0, 10, 1), false, true),
	"下元": newAlias(lunar.NewDate(0, 10, 15), true, false),
	"腊八": newAlias(lunar.NewDate(0, 12, 8), true, false),
}

func main() {
	flag.Parse()

	d := lunar.DateByTime(time.Now().In(CST))
	if *year != 0 {
		d.Year = *year
	}

	if args := flag.Args(); len(args) > 0 {
		s := args[0]
		if v, ok := aliasMap[s]; ok {
			d.Month, d.Day = v.d.Month, v.d.Day
		} else {
			t, err := time.Parse("0102", s)
			if err != nil {
				log.Fatal(err)
			}
			d.Month, d.Day = int(t.Month()), t.Day()
		}
	}

	var (
		result     *lunar.Result
		err        error
		resultDate lunar.Date
	)
	if *reverseMode {
		result, err = lunar.LunarDateToDate(d)
		resultDate = result.Date
	} else {
		result, err = lunar.DateToLunarDate(d)
		resultDate = result.LunarDate
	}
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resultDate.Time().Format(*format))
}
