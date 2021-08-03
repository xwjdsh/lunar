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

type Alias struct {
	Name        string
	Date        Date
	IsLunarDate bool
	Tags        []string
}

func NewAlias(name string, d Date, isLunarDate bool, tags ...string) *Alias {
	return &Alias{
		Name:        name,
		Date:        d,
		IsLunarDate: isLunarDate,
		Tags:        tags,
	}
}

const holidayTag = "holiday"

var holidayTags = []string{holidayTag}
var commonAliases = []*Alias{
	NewAlias("春节", NewDate(0, 1, 1), true, holidayTag),
	NewAlias("元旦", NewDate(0, 1, 1), false, holidayTag),
	NewAlias("元宵", NewDate(0, 1, 15), true),
	NewAlias("清明", NewDate(0, 4, 4), false, holidayTag),
	NewAlias("劳动", NewDate(0, 5, 1), false, holidayTag),
	NewAlias("端午", NewDate(0, 5, 5), true, holidayTag),
	NewAlias("七夕", NewDate(0, 7, 7), true),
	NewAlias("中元", NewDate(0, 7, 15), true),
	NewAlias("中秋", NewDate(0, 8, 15), true, holidayTag),
	NewAlias("重阳", NewDate(0, 9, 9), true),
	NewAlias("国庆", NewDate(0, 10, 1), false, holidayTag),
	NewAlias("下元", NewDate(0, 10, 15), true),
	NewAlias("腊八", NewDate(0, 12, 8), true),
}

var (
	aliasMap       = map[string]*Alias{}
	dateToAliasMap = map[dateWithLunar]*Alias{}
)

type dateWithLunar struct {
	isLunar bool
	Date
}

func init() {
	for _, a := range commonAliases {
		aliasMap[a.Name] = a
		dateToAliasMap[dateWithLunar{Date: a.Date, isLunar: a.IsLunarDate}] = a
	}
}

type Result struct {
	Aliases    []Alias
	Date       Date
	LunarDate  Date
	Weekday    time.Weekday
	WeekdayRaw string
	SolarTerm  string
}

type Date struct {
	Year  int
	Month int
	Day   int
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

func (d Date) Equal(d1 Date) bool {
	return d == d1
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
	dateCache      map[Date]*Result
	lunarDateCache map[Date]*Result
}

type Handler struct {
	cacheMap map[int]*fileCache
}

func New() *Handler {
	return &Handler{
		cacheMap: map[int]*fileCache{},
	}
}

func Holidays(year int) ([]*Result, error) {
	return defaultHandler.Holidays(year)
}

func (h *Handler) Holidays(year int) ([]*Result, error) {
	return h.getAliases(year, func(a *Alias) bool {
		for _, t := range a.Tags {
			if t == holidayTag {
				return true
			}
		}
		return false
	})
}

func GetSolarTerm(name string, year int) (*Result, error) {
	return defaultHandler.GetSolarTerm(name, year)
}

func (h *Handler) GetSolarTerm(name string, year int) (*Result, error) {
	rs, err := h.getSolarTerms(year, func(r *Result) bool { return r.SolarTerm == name })
	if err != nil {
		return nil, err
	}

	if len(rs) == 0 {
		return nil, ErrNotFound
	}

	return rs[0], nil
}

func GetSolarTerms(year int) ([]*Result, error) {
	return defaultHandler.GetSolarTerms(year)
}

func (h *Handler) GetSolarTerms(year int) ([]*Result, error) {
	return h.getSolarTerms(year, nil)
}

func (h *Handler) getSolarTerms(year int, filterFunc func(*Result) bool) ([]*Result, error) {
	var results []*Result
	for _, y := range []int{year, year + 1} {
		_, err := h.DateToLunarDate(NewDate(y, 1, 1))
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

func GetAliasesByTag(year int, tag string) ([]*Result, error) {
	return defaultHandler.GetAliasesByTag(year, tag)
}

func (h *Handler) GetAliasesByTag(year int, tag string) ([]*Result, error) {
	return h.getAliases(year, func(a *Alias) bool {
		for _, t := range a.Tags {
			if t == tag {
				return true
			}
		}
		return false
	})
}

func GetAliases(year int, names ...string) ([]*Result, error) {
	return defaultHandler.GetAliases(year, names...)
}

func (h *Handler) GetAliases(year int, names ...string) ([]*Result, error) {
	if len(names) == 0 {
		return h.getAliases(year, nil)
	}

	nameMap := map[string]bool{}
	for _, name := range names {
		nameMap[name] = true
	}

	return h.getAliases(year, func(a *Alias) bool {
		return nameMap[a.Name]
	})
}

func (h *Handler) getAliases(year int, filterFunc func(*Alias) bool) ([]*Result, error) {
	var (
		rs []*Result
		dm = map[Date]bool{}
	)

	for _, a := range commonAliases {
		if filterFunc != nil && !filterFunc(a) {
			continue
		}

		r, err := h.getAliasResult(a, year)
		if err != nil {
			return nil, err
		}
		if dm[r.Date] {
			// eg. 2001-10-01, 既是国庆也是中秋
			continue
		}

		rs = append(rs, r)
		dm[r.Date] = true
	}

	return rs, nil
}

func DateToLunarDate(d Date) (*Result, error) {
	return defaultHandler.DateToLunarDate(d)
}

func (h *Handler) DateToLunarDate(d Date) (*Result, error) {
	loaded, r := h.queryCache(d.Year, d, true)
	if loaded && r != nil {
		return r, nil
	}

	f, err := loadFileFunc(fmt.Sprintf("T%dc.txt", d.Year))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	lunarMonth := 0
	return h.find(f, d, true, d.Year, d.Year-1, &lunarMonth)
}

func LunarDateToDate(d Date) (*Result, error) {
	return defaultHandler.LunarDateToDate(d)
}

func (h *Handler) LunarDateToDate(d Date) (*Result, error) {
	fileLoaded := false
	var r *Result
	fileLoaded, r = h.queryCache(d.Year, d, false)
	if fileLoaded && r != nil {
		return r, nil
	}

	lunarMonth := 0
	if !fileLoaded {
		f, err := loadFileFunc(fmt.Sprintf("T%dc.txt", d.Year))
		if err != nil {
			return nil, err
		}
		defer f.Close()

		r, err := h.find(f, d, false, d.Year, d.Year-1, &lunarMonth)
		if err == nil {
			return r, nil
		}
		if err != ErrNotFound {
			return nil, err
		}
	}

	fileLoaded, r = h.queryCache(d.Year+1, d, false)
	if fileLoaded && r != nil {
		return r, nil
	}

	if !fileLoaded {
		f1, err := loadFileFunc(fmt.Sprintf("T%dc.txt", d.Year+1))
		if err != nil {
			return nil, err
		}
		defer f1.Close()
		return h.find(f1, d, false, d.Year+1, d.Year, &lunarMonth)
	}

	return nil, ErrNotFound
}

func (h *Handler) find(rd io.Reader, d Date, dateToLunarDate bool, fileYear, lunarYear int, lunarMonth *int) (*Result, error) {
	r, err := prepareReader(rd)
	if err != nil {
		return nil, err
	}

	var result *Result
	unknownMonthResults := []*Result{}
	for {
		line, err := r.ReadString('\n')
		if len(line) == 0 && err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		res, newunknownMonthResults, err := h.parseLine(line, fileYear, lunarYear, *lunarMonth, unknownMonthResults)
		if res == nil && err == nil {
			continue
		}

		if err != nil {
			return nil, err
		}
		lunarYear, *lunarMonth = res.LunarDate.Year, res.LunarDate.Month

		if dateToLunarDate {
			if res.Date.Equal(d) {
				result = res
			}
		} else {
			if res.LunarDate.Equal(d) {
				result = res
			}

			if len(unknownMonthResults) > 0 && len(newunknownMonthResults) == 0 {
				for _, v := range unknownMonthResults {
					if v.LunarDate.Equal(d) {
						result = res
					}
				}
			}
		}

		unknownMonthResults = newunknownMonthResults
	}

	if result == nil || !result.LunarDate.Valid() {
		return nil, ErrNotFound
	}

	return result, nil
}

func (h *Handler) parseLine(line string, fileYear int, lunarYear, lunarMonth int, unknownMonthResults []*Result) (*Result, []*Result, error) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return nil, nil, nil
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

	newunknownMonthResults := unknownMonthResults
	if isMonth && len(unknownMonthResults) > 0 {
		tmpLunarMonth := lunarMonth - 1
		if tmpLunarMonth == 0 {
			tmpLunarMonth = 12
		}

		for _, v := range unknownMonthResults {
			v.LunarDate.Month = tmpLunarMonth
			h.cache(v, fileYear)
		}
		newunknownMonthResults = []*Result{}
	}

	t, err := time.Parse(fileDateFormat(fileYear), fields[0])
	if err != nil {
		return nil, nil, fmt.Errorf("lunar: parse time error: %w", err)
	}

	weekday := []rune(fields[2])
	r := &Result{
		Date:       DateByTime(t),
		LunarDate:  NewDate(lunarYear, lunarMonth, lunarDay),
		WeekdayRaw: fields[2],
		Weekday:    time.Weekday(lunarMap[weekday[len(weekday)-1]]),
	}
	if len(fields) > 3 {
		r.SolarTerm = fields[3]
	}

	if lunarMonth == 0 {
		newunknownMonthResults = append(unknownMonthResults, r)
	} else {
		h.cache(r, fileYear)
	}

	return r, newunknownMonthResults, nil
}

func (h *Handler) cache(r *Result, fileYear int) {
	d := NewDate(0, r.Date.Month, r.Date.Day)
	if a, ok := dateToAliasMap[dateWithLunar{Date: d, isLunar: false}]; ok {
		r.Aliases = append(r.Aliases, *a)
	}

	d = NewDate(0, r.LunarDate.Month, r.LunarDate.Day)
	if a, ok := dateToAliasMap[dateWithLunar{Date: d, isLunar: true}]; ok {
		r.Aliases = append(r.Aliases, *a)
	}

	c, ok := h.cacheMap[fileYear]
	if !ok {
		c = &fileCache{
			dateCache:      map[Date]*Result{},
			lunarDateCache: map[Date]*Result{},
		}
		h.cacheMap[fileYear] = c
	}

	c.dateCache[r.Date] = r
	c.lunarDateCache[r.LunarDate] = r
}

func (h *Handler) queryCache(fileYear int, d Date, dateToLunarDate bool) (bool, *Result) {
	c, loaded := h.cacheMap[fileYear]
	if !loaded {
		return false, nil
	}

	var r *Result
	if dateToLunarDate {
		r = c.dateCache[d]
	} else {
		r = c.lunarDateCache[d]
	}

	return true, r
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

func (h *Handler) getAliasResult(a *Alias, year int) (*Result, error) {
	d := a.Date
	if !a.IsLunarDate {
		d.Year = year
		return h.DateToLunarDate(d)
	}

	for _, y := range []int{year, year - 1} {
		d.Year = y
		r, err := h.LunarDateToDate(d)
		if err != nil {
			return nil, err
		}
		if r.Date.Year == year {
			return r, nil
		}
	}

	return nil, ErrNotFound
}
