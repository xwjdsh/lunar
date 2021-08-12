package alias

import "github.com/xwjdsh/lunar"

type Alias struct {
	Name        string     `json:"name"`
	Disable     bool       `json:"disable"`
	Date        lunar.Date `json:"date"`
	IsLunarDate bool       `json:"is_lunar_date"`
	Tags        []string   `json:"tags"`
}

func New(name string, d lunar.Date, isLunarDate bool, tags ...string) *Alias {
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
	New("春节", lunar.NewDate(0, 1, 1), true, holidayTag),
	New("元旦", lunar.NewDate(0, 1, 1), false, holidayTag),
	New("元宵", lunar.NewDate(0, 1, 15), true),
	New("清明", lunar.NewDate(0, 4, 4), false, holidayTag),
	New("劳动", lunar.NewDate(0, 5, 1), false, holidayTag),
	New("端午", lunar.NewDate(0, 5, 5), true, holidayTag),
	New("七夕", lunar.NewDate(0, 7, 7), true),
	New("中元", lunar.NewDate(0, 7, 15), true),
	New("中秋", lunar.NewDate(0, 8, 15), true, holidayTag),
	New("重阳", lunar.NewDate(0, 9, 9), true),
	New("国庆", lunar.NewDate(0, 10, 1), false, holidayTag),
	New("下元", lunar.NewDate(0, 10, 15), true),
	New("腊八", lunar.NewDate(0, 12, 8), true),
}

type dateWithLunar struct {
	isLunar bool
	lunar.Date
}

type Result struct {
	Aliases []Alias
	*lunar.Result
}

type Handler struct {
	*lunar.Handler
	aliasMap       map[string]*Alias
	dateToAliasMap map[dateWithLunar][]*Alias
}

func NewHandler(h *lunar.Handler) *Handler {
	handler := &Handler{
		Handler:        h,
		aliasMap:       map[string]*Alias{},
		dateToAliasMap: map[dateWithLunar][]*Alias{},
	}
	for _, a := range commonAliases {
		handler.aliasMap[a.Name] = a
	}

	handler.refreshDateMap()
	return handler
}

func (h *Handler) refreshDateMap() {
	h.dateToAliasMap = map[dateWithLunar][]*Alias{}
	for _, a := range h.aliasMap {
		dl := dateWithLunar{Date: a.Date, isLunar: a.IsLunarDate}
		h.dateToAliasMap[dl] = append(h.dateToAliasMap[dl], a)
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
		rs []*Result
		dm = map[lunar.Date]bool{}
	)

	for _, a := range h.aliasMap {
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

func (h *Handler) getAliasResult(a *Alias, year int) (*Result, error) {
	d := a.Date
	if !a.IsLunarDate {
		d.Year = year
		return h.WrapResult(h.DateToLunarDate(d))
	}

	for _, y := range []int{year, year - 1} {
		d.Year = y
		r, err := h.LunarDateToDate(lunar.LunarDate{Date: d})
		if err != nil {
			return nil, err
		}
		if r.Date.Year == year {
			return h.resultWithAliases(r), nil
		}
	}

	return nil, lunar.ErrNotFound
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
	d := lunar.NewDate(0, r.Date.Month, r.Date.Day)
	nr := &Result{Result: r}
	if as, ok := h.dateToAliasMap[dateWithLunar{Date: d, isLunar: false}]; ok {
		for _, a := range as {
			nr.Aliases = append(nr.Aliases, *a)
		}
	}

	d = lunar.NewDate(0, r.LunarDate.Month, r.LunarDate.Day)
	if as, ok := h.dateToAliasMap[dateWithLunar{Date: d, isLunar: true}]; ok {
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
