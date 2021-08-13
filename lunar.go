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
	Date       Date
	LunarDate  LunarDate
	Weekday    time.Weekday
	WeekdayRaw string
	SolarTerm  string
}

type DateType interface {
	IsLunarDate() bool
}

var (
	_ DateType = Date{}
	_ DateType = LunarDate{}
)

type Date struct {
	Year  int `json:"year"`
	Month int `json:"month"`
	Day   int `json:"day"`
}

func (d Date) IsLunarDate() bool {
	return false
}

type LunarDate struct {
	Date
	IsLeapMonth bool
}

func NewLunarDate(d Date, isLeapMonth bool) LunarDate {
	return LunarDate{Date: d, IsLeapMonth: isLeapMonth}
}

func (d LunarDate) IsLunarDate() bool {
	return true
}

func NewDate(y, m, d int) Date {
	return Date{Year: y, Month: m, Day: d}
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

func (d Date) String() string {
	return d.Time().Format("20060102")
}

func (d Date) Valid() bool {
	return d.Year != 0 && d.Month != 0 && d.Day != 0
}

func fileDateFormat(year int) string {
	format := "2006年1月2日"
	if year <= 2010 {
		format = "2006年01月02日"
	}

	return format
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

var defaultHandler = New()

type fileCache struct {
	results        []*Result
	dateCache      map[Date]*Result
	lunarDateCache map[LunarDate]*Result
}

type Handler struct {
	cacheMap map[int]*fileCache
}

func New() *Handler {
	return &Handler{
		cacheMap: map[int]*fileCache{},
	}
}

func GetSolarTerms(year int, names ...string) ([]*Result, error) {
	return defaultHandler.GetSolarTerms(year, names...)
}

func (h *Handler) GetSolarTerms(year int, names ...string) ([]*Result, error) {
	if len(names) == 0 {
		return h.getSolarTerms(year, nil)
	}
	nameMap := map[string]bool{}
	for _, name := range names {
		nameMap[name] = true
	}

	return h.getSolarTerms(year, func(r *Result) bool {
		return nameMap[r.SolarTerm]
	})
}

func (h *Handler) getSolarTerms(year int, filterFunc func(*Result) bool) ([]*Result, error) {
	var results []*Result
	for _, y := range []int{year, year + 1} {
		_, err := h.dateToLunarDate(NewDate(y, 1, 1))
		if err != nil {
			return nil, err
		}

		for _, r := range h.cacheMap[y].lunarDateCache {
			if r.SolarTerm != "" && r.LunarDate.Year == year {
				if filterFunc == nil || filterFunc(r) {
					results = append(results, r)
				}
			}
		}
	}

	return results, nil
}

func Calendar(dt DateType) (*Result, error) {
	return defaultHandler.Calendar(dt)
}

func (h *Handler) Calendar(dt DateType) (*Result, error) {
	var (
		r   *Result
		err error
	)
	if dt.IsLunarDate() {
		r, err = h.lunarDateToDate(dt.(LunarDate))
	} else {
		r, err = h.dateToLunarDate(dt.(Date))
	}

	return r, err
}

func (h *Handler) dateToLunarDate(d Date) (*Result, error) {
	if loaded, r, _ := h.queryCache(d.Year, d); loaded && r != nil {
		return r, nil
	}

	var lastResult *Result
	if loaded, _, lr := h.queryCache(d.Year-1, d); loaded {
		lastResult = lr
	} else {
		f, err := loadFileFunc(fmt.Sprintf("T%dc.txt", d.Year-1))
		if err != nil {
			return nil, err
		}
		defer f.Close()

		_, lr, err := h.find(f, d, d.Year-1, NewLunarDate(NewDate(d.Year-2, 0, 0), false), false)
		if err != nil && err != ErrNotFound {
			return nil, err
		}
		lastResult = lr
	}

	f, err := loadFileFunc(fmt.Sprintf("T%dc.txt", d.Year))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r, _, err := h.find(f, d, d.Year, lastResult.LunarDate, true)
	return r, err
}

func (h *Handler) lunarDateToDate(d LunarDate) (*Result, error) {
	var lastResult *Result
	if fileLoaded, r, lr := h.queryCache(d.Year, d); fileLoaded {
		if r != nil {
			return r, nil
		}
		lastResult = lr
	} else {
		f, err := loadFileFunc(fmt.Sprintf("T%dc.txt", d.Year))
		if err != nil {
			return nil, err
		}
		defer f.Close()

		r, lr, err := h.find(f, d, d.Year, NewLunarDate(NewDate(d.Year-1, 0, 0), false), false)
		if err == nil {
			return r, nil
		}
		if err != ErrNotFound {
			return nil, err
		}
		lastResult = lr
	}

	fileLoaded, r, _ := h.queryCache(d.Year+1, d)
	if fileLoaded && r != nil {
		return r, nil
	}

	if !fileLoaded {
		f1, err := loadFileFunc(fmt.Sprintf("T%dc.txt", d.Year+1))
		if err != nil {
			return nil, err
		}
		defer f1.Close()

		r, _, err := h.find(f1, d, d.Year+1, lastResult.LunarDate, true)
		return r, err
	}

	return nil, ErrNotFound
}

func (h *Handler) find(rd io.Reader, dt DateType, fileYear int, lastLunarDate LunarDate, saveCache bool) (*Result, *Result, error) {
	r, err := prepareReader(rd)
	if err != nil {
		return nil, nil, err
	}

	isLunarDate := dt.IsLunarDate()
	lunarYear, lunarMonth := lastLunarDate.Year, lastLunarDate.Month
	isLeapMonth := lastLunarDate.IsLeapMonth

	var (
		result     *Result
		lastResult *Result
	)
	for {
		line, err := r.ReadString('\n')
		if len(line) == 0 && err != nil {
			if err == io.EOF {
				break
			}
			return nil, nil, err
		}

		res, err := h.parseLine(line, fileYear, lunarYear, lunarMonth, isLeapMonth)
		if res == nil && err == nil {
			continue
		}

		if err != nil {
			return nil, nil, err
		}

		if saveCache {
			h.cache(res, fileYear)
		}

		lastResult = res
		isLeapMonth = res.LunarDate.IsLeapMonth
		lunarYear, lunarMonth = res.LunarDate.Year, res.LunarDate.Month
		if (!isLunarDate && res.Date == dt.(Date)) ||
			(isLunarDate && res.LunarDate == dt.(LunarDate)) {
			result = res
		}
	}

	if result == nil {
		return nil, lastResult, ErrNotFound
	}

	return result, lastResult, nil
}

func (h *Handler) parseLine(line string, fileYear int, lunarYear, lunarMonth int, isLeapMonth bool) (*Result, error) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return nil, nil
	}

	rs := []rune(fields[1])
	isMonth := false
	if rs[len(rs)-1] == rune('月') {
		isMonth = true
		rs = rs[:len(rs)-1]
	}

	if isMonth {
		isLeapMonth = false
		if rs[0] == rune('閏') {
			isLeapMonth = true
			rs = rs[1:]
		}
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

	t, err := time.Parse(fileDateFormat(fileYear), fields[0])
	if err != nil {
		return nil, fmt.Errorf("lunar: parse time error: %w", err)
	}

	weekday := []rune(fields[2])
	r := &Result{
		Date:       DateByTime(t),
		LunarDate:  NewLunarDate(NewDate(lunarYear, lunarMonth, lunarDay), isLeapMonth),
		WeekdayRaw: fields[2],
		Weekday:    time.Weekday(lunarMap[weekday[len(weekday)-1]]),
	}
	if len(fields) > 3 {
		r.SolarTerm = fields[3]
	}

	return r, nil
}

func (h *Handler) cache(r *Result, fileYear int) {
	c, ok := h.cacheMap[fileYear]
	if !ok {
		c = &fileCache{
			results:        []*Result{},
			dateCache:      map[Date]*Result{},
			lunarDateCache: map[LunarDate]*Result{},
		}
		h.cacheMap[fileYear] = c
	}

	c.results = append(c.results, r)
	c.dateCache[r.Date] = r
	c.lunarDateCache[r.LunarDate] = r
}

func (h *Handler) queryCache(fileYear int, dt DateType) (bool, *Result, *Result) {
	isLunarDate := dt.IsLunarDate()
	c, loaded := h.cacheMap[fileYear]
	if !loaded {
		return false, nil, nil
	}

	var r *Result
	if isLunarDate {
		r = c.lunarDateCache[dt.(LunarDate)]
	} else {
		r = c.dateCache[dt.(Date)]
	}

	return true, r, c.results[len(c.results)-1]
}

func prepareReader(rd io.Reader) (*bufio.Reader, error) {
	r := bufio.NewReader(rd)

	// skip first three lines
	for i := 0; i < 3; i++ {
		line, err := r.ReadString('\n')
		if len(line) == 0 && err != nil {
			return nil, err
		}
	}

	return r, nil
}
