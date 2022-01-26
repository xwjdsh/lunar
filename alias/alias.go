package alias

import (
	"github.com/xwjdsh/lunar"
	"github.com/xwjdsh/lunar/config"
)

// Alias represents a date alias
type Alias struct {
	Name  string
	Dates []lunar.DateType
	Tags  []string
}

// ConvertAlias convert config.Alias to Alias
func ConvertAlias(c *config.Alias) *Alias {
	var dts []lunar.DateType
	date := lunar.Date(c.Date)
	if c.IsLunarDate {
		dts = getLunarDates(date, c.LeapMonthLimit)
	} else {
		dts = getDates(date)
	}

	return New(c.Name, dts, c.Tags...)
}

// New returns a new Alias
func New(name string, ds []lunar.DateType, tags ...string) *Alias {
	return &Alias{
		Name:  name,
		Dates: ds,
		Tags:  tags,
	}
}

// Result wraps lunar.Result with aliases
type Result struct {
	Aliases []Alias
	*lunar.Result
}

// Handler alias handler
type Handler struct {
	*lunar.Handler
	aliasMap       map[string]*Alias
	dateToAliasMap map[lunar.DateType][]*Alias
}

// NewHandler returns a new Handler
func NewHandler(h *lunar.Handler) *Handler {
	return &Handler{
		Handler:        h,
		aliasMap:       map[string]*Alias{},
		dateToAliasMap: map[lunar.DateType][]*Alias{},
	}
}

// GetAliasesByTag get alias results by tag
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

// GetAliases query aliases
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
			r, err := h.WrapResult(h.Calendar(d))
			if err != nil {
				if err == lunar.ErrNotFound {
					continue
				}
				return nil, err
			}
			results = append(results, r)
			continue
		}

		d := dt.(lunar.LunarDate)
		for _, y := range []int{year, year - 1} {
			d.Year = y
			r, err := h.Calendar(d)
			if err != nil {
				if err == lunar.ErrNotFound {
					continue
				}
				return nil, err
			}
			if r.Date.Year == year {
				results = append(results, h.resultWithAliases(r))
			}
		}
	}

	return results, nil
}

// WrapResults wrap results with alias info
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

// WrapResult wrap result with alias info
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

// LoadAlias load alias config
func (h *Handler) LoadAlias(cs []*config.Alias) {
	h.aliasMap = map[string]*Alias{}
	for _, c := range cs {
		if !c.Disable {
			h.aliasMap[c.Name] = ConvertAlias(c)
		}
	}

	h.dateToAliasMap = map[lunar.DateType][]*Alias{}
	for _, a := range h.aliasMap {
		for _, dt := range a.Dates {
			h.dateToAliasMap[dt] = append(h.dateToAliasMap[dt], a)
		}
	}
}

func getLunarDates(d lunar.Date, leapMonthType config.LeapMonthLimitType) []lunar.DateType {
	var results []lunar.DateType
	switch leapMonthType {
	case config.LeapMonthNoLimit:
		results = []lunar.DateType{lunar.NewLunarDate(d, true), lunar.NewLunarDate(d, false)}
	case config.LeapMonthOnly:
		results = []lunar.DateType{lunar.NewLunarDate(d, true)}
	case config.LeapMonthOnlyNot:
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
