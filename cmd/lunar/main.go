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

func main() {
	flag.Parse()

	d := lunar.DateByTime(time.Now().In(CST))
	if *year != 0 {
		d.Year = *year
	}

	if args := flag.Args(); len(args) > 0 {
		t, err := time.Parse("0102", args[0])
		if err != nil {
			log.Fatal(err)
		}
		d.Month, d.Day = int(t.Month()), t.Day()
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
