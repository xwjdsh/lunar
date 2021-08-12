package alias

import "github.com/xwjdsh/lunar"

type Alias struct {
	Name    string           `json:"name"`
	Disable bool             `json:"disable"`
	Dates   []lunar.DateType `json:"dates"`
	Tags    []string         `json:"tags"`
}

func New(name string, ds []lunar.DateType, tags ...string) *Alias {
	return &Alias{
		Name:  name,
		Dates: ds,
		Tags:  tags,
	}
}

type LeapMonthLimitType int

const (
	LeapMonthNoLimit LeapMonthLimitType = iota
	LeapMonthOnly
	LeapMonthOnlyNot
)

const holidayTag = "holiday"

var holidayTags = []string{holidayTag}
var commonAliases = []*Alias{
	New("春节", getLunarDates(lunar.NewDate(0, 1, 1), LeapMonthOnlyNot), holidayTag),
	New("元旦", getDates(lunar.NewDate(0, 1, 1)), holidayTag),
	New("元宵", getLunarDates(lunar.NewDate(0, 1, 15), LeapMonthOnlyNot)),
	New("清明", getDates(lunar.NewDate(0, 4, 4)), holidayTag),
	New("劳动", getDates(lunar.NewDate(0, 5, 1)), holidayTag),
	New("端午", getLunarDates(lunar.NewDate(0, 5, 5), LeapMonthOnlyNot), holidayTag),
	New("七夕", getLunarDates(lunar.NewDate(0, 7, 7), LeapMonthOnlyNot)),
	New("中元", getLunarDates(lunar.NewDate(0, 7, 15), LeapMonthOnlyNot)),
	New("中秋", getLunarDates(lunar.NewDate(0, 8, 15), LeapMonthOnlyNot), holidayTag),
	New("重阳", getLunarDates(lunar.NewDate(0, 9, 9), LeapMonthOnlyNot)),
	New("国庆", getDates(lunar.NewDate(0, 10, 1)), holidayTag),
	New("下元", getLunarDates(lunar.NewDate(0, 10, 15), LeapMonthOnlyNot)),
	New("腊八", getLunarDates(lunar.NewDate(0, 12, 8), LeapMonthOnlyNot)),
}

type Result struct {
	Aliases []Alias
	*lunar.Result
}

type Handler struct {
	*lunar.Handler
	aliasMap       map[string]*Alias
	dateToAliasMap map[lunar.DateType][]*Alias
}

func NewHandler(h *lunar.Handler) *Handler {
	handler := &Handler{
		Handler:        h,
		aliasMap:       map[string]*Alias{},
		dateToAliasMap: map[lunar.DateType][]*Alias{},
	}
	for _, a := range commonAliases {
		handler.aliasMap[a.Name] = a
	}

	handler.refreshDateMap()
	return handler
}

func (h *Handler) refreshDateMap() {
	h.dateToAliasMap = map[lunar.DateType][]*Alias{}
	for _, a := range h.aliasMap {
		for _, dt := range a.Dates {
			h.dateToAliasMap[dt] = append(h.dateToAliasMap[dt], a)
		}
	}
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
		results []*Result
		dm      = map[lunar.Date]bool{}
	)

	for _, a := range h.aliasMap {
		if filterFunc != nil && !filterFunc(a) {
			continue
		}

		rs, err := h.getAliasResult(a, year)
		if err != nil {
			return nil, err
		}
		for _, r := range rs {
			if dm[r.Date] {
				// eg. 2001-10-01, 既是国庆也是中秋
				continue
			}

			results = append(results, r)
			dm[r.Date] = true
		}
	}

	return results, nil
}

func (h *Handler) getAliasResult(a *Alias, year int) ([]*Result, error) {
	results := []*Result{}
	for _, dt := range a.Dates {
		if !dt.IsLunarDate() {
			d := dt.(lunar.Date)
			d.Year = year
			r, err := h.WrapResult(h.DateToLunarDate(d))
			if err != nil {
				if err == lunar.ErrNotFound {
					continue
				}
				return nil, err
			}
			results = append(results, r)
			continue
		}

		for _, y := range []int{year, year - 1} {
			d := dt.(lunar.LunarDate)
			d.Year = y
			r, err := h.LunarDateToDate(d)
			if err != nil {
				return nil, err
			}
			if r.Date.Year == year {
				results = append(results, h.resultWithAliases(r))
			}
		}
	}

	return results, nil
}

func (h *Handler) WrapResults(rs []*lunar.Result, err error) ([]*Result, error) {
	if err != nil {
		return nil, err
	}

	var nrs []*Result
	for _, r := range rs {
		nrs = append(nrs, h.resultWithAliases(r))
	}

	return nrs, nil
}

func (h *Handler) WrapResult(r *lunar.Result, err error) (*Result, error) {
	if err != nil {
		return nil, err
	}

	return h.resultWithAliases(r), nil
}

func (h *Handler) resultWithAliases(r *lunar.Result) *Result {
	d := r.Date
	d.Year = 0
	nr := &Result{Result: r}
	if as, ok := h.dateToAliasMap[d]; ok {
		for _, a := range as {
			nr.Aliases = append(nr.Aliases, *a)
		}
	}

	d1 := r.LunarDate
	d1.Year = 0
	if as, ok := h.dateToAliasMap[d1]; ok {
		for _, a := range as {
			nr.Aliases = append(nr.Aliases, *a)
		}
	}

	return nr
}

func (h *Handler) LoadCustomAlias(as []*Alias) error {
	for _, a := range as {
		if a.Disable {
			delete(h.aliasMap, a.Name)
		} else {
			h.aliasMap[a.Name] = a
		}
	}

	h.refreshDateMap()
	return nil
}

func getLunarDates(d lunar.Date, leapMonthType LeapMonthLimitType) []lunar.DateType {
	var results []lunar.DateType
	switch leapMonthType {
	case LeapMonthNoLimit:
		results = []lunar.DateType{lunar.NewLunarDate(d, true), lunar.NewLunarDate(d, false)}
	case LeapMonthOnly:
		results = []lunar.DateType{lunar.NewLunarDate(d, true)}
	case LeapMonthOnlyNot:
		results = []lunar.DateType{lunar.NewLunarDate(d, false)}
	}

	return results
}

func getDates(ds ...lunar.Date) []lunar.DateType {
	result := []lunar.DateType{}
	for _, d := range ds {
		result = append(result, d)
	}

	return result
}
