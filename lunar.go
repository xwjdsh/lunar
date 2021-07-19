package lunar

import (
	"bufio"
	"embed"
	"fmt"
	"io"
	"strings"
	"time"
)

/*
cd ./files && curl -O https://www.hko.gov.hk/tc/gts/time/calendar/text/files/T\[1901-2100\]c.txt && \
	find . -type f -exec sh -c 'iconv -f big5 -t utf-8 -c {} > {}.utf8' \; -exec mv "{}".utf8 "{}" \; && cd ..
*/

//go:embed files/*
var files embed.FS

type Date struct {
	Date      time.Time
	LunarDate time.Time
	Weekday   time.Weekday
	SolarTerm string
}

var lunarMap = map[string]int{
	"初": 0,
	"正": 1,
	"二": 2,
	"廿": 2,
	"三": 3,
	"四": 4,
	"五": 5,
	"六": 6,
	"七": 7,
	"八": 8,
	"九": 9,
	"十": 10,
}

var weekdayMap = map[string]time.Weekday{
	"星期一": time.Monday,
	"星期二": time.Tuesday,
	"星期三": time.Wednesday,
	"星期四": time.Thursday,
	"星期五": time.Friday,
	"星期六": time.Saturday,
	"星期天": time.Sunday,
}

var numberMap = map[string]int{}

func Calendar(t time.Time) (*Date, error) {
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	f, err := files.Open(fmt.Sprintf("files/T%dc.txt", t.Year()))
	if err != nil {
		return nil, err
	}

	defer f.Close()
	return find(f, t)
}

func find(rd io.Reader, t time.Time) (*Date, error) {
	target := t.Format("2006年1月2日")
	r := bufio.NewReader(rd)

	// skip first three lines
	for i := 0; i < 3; i++ {
		line, err := r.ReadString('\n')
		if len(line) == 0 && err != nil {
			return nil, err
		}
	}

	for {
		line, err := r.ReadString('\n')
		if len(line) == 0 && err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		res := strings.Fields(line)
		if res[0] != target {
			continue
		}

		d := &Date{
			Date:    t,
			Weekday: weekdayMap[res[2]],
		}
		if len(res) > 3 {
			d.SolarTerm = res[3]
		}
		return d, nil
	}

	return nil, nil
}
