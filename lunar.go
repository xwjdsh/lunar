package lunar

import (
	"bufio"
	"embed"
	"errors"
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

var ErrNotFound = errors.New("lunar: date not found")

type Date struct {
	Date                            time.Time
	lunarYear, lunarMonth, lunarDay int
	LunarDate                       time.Time
	Weekday                         time.Weekday
	SolarTerm                       string
}

var lunarMap = map[rune]int{
	'天': 0,
	'初': 0,
	'正': 1,
	'一': 1,
	'二': 2,
	'廿': 2,
	'三': 3,
	'四': 4,
	'五': 5,
	'六': 6,
	'七': 7,
	'八': 8,
	'九': 9,
	'十': 10,
}

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
	format := "2006年1月2日"
	if t.Year() <= 2010 {
		format = "2006年01月02日"
	}
	target := t.Format(format)
	r := bufio.NewReader(rd)

	// skip first three lines
	for i := 0; i < 3; i++ {
		line, err := r.ReadString('\n')
		if len(line) == 0 && err != nil {
			return nil, err
		}
	}

	lunarYear := t.Year() - 1
	lunarMonth := 0
	var beforeResult *Date
	for {
		line, err := r.ReadString('\n')
		if len(line) == 0 && err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		rs := []rune(fields[1])
		if rs[0] == rune('閏') {
			rs = rs[1:]
		}

		isMonth := false
		if rs[len(rs)-1] == rune('月') {
			isMonth = true
			rs = rs[:len(rs)-1]
		}

		lastChar := rs[len(rs)-1]
		unitDigit := lunarMap[lastChar]
		if lastChar == '正' {
			lunarYear++
		}

		tensDigit := 0
		if len(rs) > 1 {
			tensDigit = lunarMap[rs[0]]
			if tensDigit == 10 {
				tensDigit = 1
			}
			if tensDigit != 0 && unitDigit == 10 {
				tensDigit--
			}
		}

		lunarDay := tensDigit*10 + unitDigit
		if isMonth {
			lunarMonth = lunarDay
			lunarDay = 1
		}

		if beforeResult != nil && isMonth {
			lunarMonth--
			if lunarMonth == 0 {
				lunarMonth = 12
			}
			beforeResult.LunarDate = time.Date(beforeResult.lunarYear, time.Month(lunarMonth), beforeResult.lunarDay, 0, 0, 0, 0, t.Location())
			return beforeResult, nil
		}

		if fields[0] != target {
			continue
		}

		weekday := []rune(fields[2])
		d := &Date{
			Date:       t,
			lunarYear:  lunarYear,
			lunarMonth: lunarMonth,
			lunarDay:   lunarDay,
			Weekday:    time.Weekday(lunarMap[weekday[len(weekday)-1]]),
		}
		if len(fields) > 3 {
			d.SolarTerm = fields[3]
		}

		if lunarMonth == 0 {
			beforeResult = d
			continue
		}
		d.LunarDate = time.Date(d.lunarYear, time.Month(d.lunarMonth), d.lunarDay, 0, 0, 0, 0, t.Location())
		return d, nil
	}

	return nil, ErrNotFound
}
