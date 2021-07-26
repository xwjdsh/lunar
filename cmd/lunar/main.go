package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/xwjdsh/lunar"
)

var (
	CST            = time.FixedZone("CST", 3600*8)
	format         = flag.String("f", "2006-01-02", "time format")
	queryLunarDate = flag.Bool("q", false, "query lunar date, lunar date to date")
)

func main() {
	flag.Parse()

	args := flag.Args()
	s := ""
	if len(args) > 0 {
		s = args[0]
	}

	d, err := parseDate(s)
	if err != nil {
		log.Fatal(err)
	}

	var result *lunar.Result
	var resultDate lunar.Date
	if *queryLunarDate {
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

func parseDate(s string) (lunar.Date, error) {
	d := lunar.DateByTime(time.Now().In(CST))
	if s == "" {
		return d, nil
	}

	t, err := time.Parse("0102", s)
	if err == nil {
		d.Month, d.Day = int(t.Month()), t.Day()
		return d, nil
	}

	t, err = time.Parse("20060102", s)
	if err != nil {
		return lunar.Date{}, err
	}

	return lunar.DateByTime(t), nil
}
