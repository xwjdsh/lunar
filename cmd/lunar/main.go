package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/xwjdsh/lunar"
)

var (
	format = flag.String("f", "2006-01-02", "time format")
)

func main() {
	flag.Parse()

	t := time.Now()
	args := flag.Args()
	if len(args) > 0 {
		t1, err := time.Parse("0102", args[0])
		if err != nil {
			t1, err = time.Parse("20060102", args[0])
		}
		if err != nil {
			log.Fatal(err)
		}

		if t1.Year() == 0 {
			t1 = time.Date(t.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, t.Location())
		}
		t = t1
	}

	d, err := lunar.Calendar(t)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(d.LunarDate.Format(*format))
}
