package config

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v3"
)

const holidayTag = "holiday"

// Date date
type Date struct {
	Year  int `yaml:"year"`
	Month int `yaml:"month"`
	Day   int `yaml:"day"`
}

func NewDate(y, m, d int) Date {
	return Date{
		Year:  y,
		Month: m,
		Day:   d,
	}
}

// LeapMonthLimitType leap month limit type
type LeapMonthLimitType int

const (
	// LeapMonthOnlyNot only not leap month
	LeapMonthOnlyNot LeapMonthLimitType = iota
	// LeapMonthOnly only leap month
	LeapMonthOnly
	// LeapMonthNoLimit no limit of leap month
	LeapMonthNoLimit
)

// Config custom config
type Config struct {
	Aliases []*Alias `yaml:"aliases"`
}

func (c *Config) Marshal() ([]byte, error) {
	return yaml.Marshal(c)
}

// Alias alias config
type Alias struct {
	Name           string             `yaml:"name"`
	Disable        bool               `yaml:"disable"`
	Date           Date               `yaml:"date"`
	IsLunarDate    bool               `yaml:"is_lunar_date"`
	LeapMonthLimit LeapMonthLimitType `yaml:"leap_month_limit"`
	Tags           []string           `yaml:"tags"`
}

// NewAlias return a new Alias instance
func NewAlias(name string, date Date, isLunarDate bool, lm LeapMonthLimitType, tags ...string) *Alias {
	return &Alias{
		Name:           name,
		Date:           date,
		IsLunarDate:    isLunarDate,
		LeapMonthLimit: lm,
		Tags:           tags,
	}
}

// Init init config
func Init(fp string, useDefault bool) (*Config, error) {
	c := defaultConfig()
	if useDefault {
		return c, nil
	}
	data, err := ioutil.ReadFile(fp)
	if err != nil {
		if os.IsNotExist(err) {
			return c, nil
		}
		return nil, err
	}
	if err := yaml.Unmarshal(data, c); err != nil {
		return nil, err
	}
	return c, nil
}

func defaultConfig() *Config {
	return &Config{
		Aliases: []*Alias{
			NewAlias("春节", NewDate(0, 1, 1), true, LeapMonthOnlyNot, holidayTag),
			NewAlias("元旦", NewDate(0, 1, 1), false, LeapMonthNoLimit, holidayTag),
			NewAlias("元宵", NewDate(0, 1, 15), true, LeapMonthOnlyNot),
			NewAlias("清明", NewDate(0, 4, 4), false, LeapMonthNoLimit, holidayTag),
			NewAlias("劳动", NewDate(0, 5, 1), false, LeapMonthNoLimit, holidayTag),
			NewAlias("端午", NewDate(0, 5, 5), true, LeapMonthOnlyNot, holidayTag),
			NewAlias("七夕", NewDate(0, 7, 7), true, LeapMonthOnlyNot),
			NewAlias("中元", NewDate(0, 7, 15), true, LeapMonthOnlyNot),
			NewAlias("中秋", NewDate(0, 8, 15), true, LeapMonthOnlyNot, holidayTag),
			NewAlias("重阳", NewDate(0, 9, 9), true, LeapMonthOnlyNot),
			NewAlias("国庆", NewDate(0, 10, 1), false, LeapMonthNoLimit, holidayTag),
			NewAlias("下元", NewDate(0, 10, 15), true, LeapMonthOnlyNot),
			NewAlias("腊八", NewDate(0, 12, 8), true, LeapMonthOnlyNot),
		},
	}
}
