package lunar

import (
	"bufio"
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

var (
	ErrNotFound  = errors.New("lunar: date not found")
	loadFileFunc func(string) (io.ReadCloser, error)
)

type Result struct {
	Date      Date
	LunarDate Date
	Weekday   time.Weekday
	SolarTerm string
}

type Date struct {
	Year  int
	Month int
	Day   int
}

func DateByTime(t time.Time) Date {
	year, month, day := t.Date()
	return Date{
		Year:  year,
		Month: int(month),
		Day:   day,
	}
}

func (d Date) Time() time.Time {
	return time.Date(d.Year, time.Month(d.Month), d.Day, 0, 0, 0, 0, time.UTC)
}

func (d Date) Equal(d1 Date) bool {
	return d.Year == d1.Year && d.Month == d1.Month && d.Day == d1.Day
}

func (d Date) String() string {
	return d.Time().Format("20060102")
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

var defaultHandler = new(Handler)

type Handler struct {
	cacheEnabled   bool
	dateCache      map[int]map[string]*Result
	lunarDateCache map[int]map[string]*Result
}

func New() *Handler {
	return &Handler{
		dateCache:      map[int]map[string]*Result{},
		lunarDateCache: map[int]map[string]*Result{},
	}
}

func DateToLunarDate(d Date) (*Result, error) {
	return defaultHandler.DateToLunarDate(d)
}

func (h *Handler) DateToLunarDate(d Date) (*Result, error) {
	fileName := fmt.Sprintf("T%dc.txt", d.Year)
	f, err := loadFileFunc(fileName)
	if err != nil {
		return nil, err
	}

	defer f.Close()
	return h.find(f, d)
}

func LunarDateToDate(d Date) (*Result, error) {
	return defaultHandler.LunarDateToDate(d)
}

func (h *Handler) LunarDateToDate(d Date) (*Result, error) {
	fileName := fmt.Sprintf("T%dc.txt", d.Year)
	f, err := loadFileFunc(fileName)
	if err != nil {
		return nil, err
	}

	defer f.Close()
	r, lunarMonth, err := h.query(f, d, d.Year-1, 0)
	if err == nil {
		return r, nil
	}
	if err != ErrNotFound {
		return nil, err
	}

	f1, err := loadFileFunc(fmt.Sprintf("T%dc.txt", d.Year+1))
	if err != nil {
		return nil, err
	}

	defer f1.Close()
	r, _, err = h.query(f1, d, d.Year, lunarMonth)
	return r, err

}

func (h *Handler) find(rd io.Reader, d Date) (*Result, error) {
	format := "2006年1月2日"
	if d.Year <= 2010 {
		format = "2006年01月02日"
	}
	target := d.Time().Format(format)
	r := bufio.NewReader(rd)

	// skip first three lines
	for i := 0; i < 3; i++ {
		line, err := r.ReadString('\n')
		if len(line) == 0 && err != nil {
			return nil, err
		}
	}

	lunarYear := d.Year - 1
	lunarMonth := 0
	var result *Result
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

		if result != nil && isMonth {
			lunarMonth--
			if lunarMonth == 0 {
				lunarYear--
				lunarMonth = 12
			}

			result.LunarDate = Date{lunarYear, lunarMonth, lunarDay}
			return result, nil
		}

		if fields[0] != target {
			continue
		}

		weekday := []rune(fields[2])
		result = &Result{
			Date:    d,
			Weekday: time.Weekday(lunarMap[weekday[len(weekday)-1]]),
		}
		if len(fields) > 3 {
			result.SolarTerm = fields[3]
		}

		if lunarMonth == 0 {
			continue
		}
		result.LunarDate = Date{lunarYear, lunarMonth, lunarDay}
		return result, nil
	}

	return nil, ErrNotFound
}

func (h *Handler) query(rd io.Reader, d Date, lunarYear, lunarMonth int) (*Result, int, error) {
	fileName := fmt.Sprintf("T%dc.txt", d.Year)
	f, err := loadFileFunc(fileName)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()

	r := bufio.NewReader(rd)

	// skip first three lines
	for i := 0; i < 3; i++ {
		line, err := r.ReadString('\n')
		if len(line) == 0 && err != nil {
			return nil, 0, err
		}
	}

	var unknownMonthResults []*Result
	var result *Result
	for {
		line, err := r.ReadString('\n')
		if len(line) == 0 && err != nil {
			if err == io.EOF {
				break
			}
			return nil, 0, err
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

		if isMonth && len(unknownMonthResults) > 0 {
			tmpLunarMonth := lunarMonth - 1
			if tmpLunarMonth == 0 {
				tmpLunarMonth = 12
			}

			for _, v := range unknownMonthResults {
				v.LunarDate.Month = tmpLunarMonth
				if v.LunarDate.Equal(d) {
					result = v
					if !h.cacheEnabled {
						return result, lunarMonth, nil
					}
				}
			}
			unknownMonthResults = nil

		}

		format := "2006年1月2日"
		if d.Year <= 2010 {
			format = "2006年01月02日"
		}

		t, err := time.Parse(format, fields[0])
		if err != nil {
			return nil, 0, fmt.Errorf("lunar: parse time error: %w", err)
		}

		weekday := []rune(fields[2])
		res := &Result{
			Date:      DateByTime(t),
			LunarDate: Date{lunarYear, lunarMonth, lunarDay},
			Weekday:   time.Weekday(lunarMap[weekday[len(weekday)-1]]),
		}
		if len(fields) > 3 {
			res.SolarTerm = fields[3]
		}

		if lunarMonth == 0 {
			unknownMonthResults = append(unknownMonthResults, res)
			continue
		}

		if res.LunarDate.Equal(d) {
			result = res
			if !h.cacheEnabled {
				return result, lunarMonth, nil
			}
		}
	}

	if result != nil {
		return result, lunarMonth, nil
	}

	return nil, lunarMonth, ErrNotFound
}
