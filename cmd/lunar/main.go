package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"

	"github.com/xwjdsh/lunar"
	"github.com/xwjdsh/lunar/alias"
)

var CST = time.FixedZone("CST", 3600*8)

type Config struct {
	Aliases []*alias.Alias `json:"aliases"`
}

func main() {
	h := alias.NewHandler(lunar.New())
	app := &cli.App{
		Name:  "lunar",
		Usage: "lunar is a command line tool for conversion between Gregorian calendar and lunar calendar.(1901~2100)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"f"},
				Value:   "2006-01-02",
				Usage:   "Output date format",
			},
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   path.Join(mustUserHomeDir(), ".config/lunar/lunar.json"),
				Usage:   "Custom config path",
			},
			&cli.IntFlag{
				Name:    "year",
				Aliases: []string{"y"},
				Value:   time.Now().In(CST).Year(),
				Usage:   "Target year",
			},
			&cli.BoolFlag{
				Name:    "reverse",
				Aliases: []string{"r"},
				Usage:   "Reverse mode, query date by lunar date",
			},
		},
		Before: func(c *cli.Context) error {
			fp := c.String("config")
			data, err := ioutil.ReadFile(fp)
			if err != nil {
				if os.IsNotExist(err) {
					return nil
				}
				return err
			}
			conf := &Config{}
			if err := json.Unmarshal(data, conf); err != nil {
				return err
			}
			return h.LoadCustomAlias(conf.Aliases)
		},
		Commands: []*cli.Command{
			{
				Name:    "alias",
				Aliases: []string{"a"},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "tag",
						Aliases: []string{"t"},
						Usage:   "query by tag",
					},
				},
				Usage: "Show alias date info",
				Action: func(c *cli.Context) error {
					d := currentDate(c)
					var (
						results []*alias.Result
						err     error
					)
					if tag := c.String("tag"); tag != "" {
						results, err = h.GetAliasesByTag(d.Year, tag)
					} else {
						if c.Args().Len() >= 1 {
							results, err = h.GetAliases(d.Year, c.Args().Slice()...)
						} else {
							results, err = h.GetAliases(d.Year)
						}
					}
					if err != nil {
						return err
					}

					outputResults(results, c)
					return nil
				},
			},
			{
				Name:    "solar-term",
				Aliases: []string{"st"},
				Usage:   "Get solar term info",
				Action: func(c *cli.Context) error {
					d := currentDate(c)
					var (
						rs  []*alias.Result
						err error
					)
					if c.Args().Len() >= 1 {
						rs, err = h.WrapResults(h.GetSolarTerms(d.Year, c.Args().Slice()...))
					} else {
						rs, err = h.WrapResults(h.GetSolarTerms(d.Year))
					}
					if err != nil {
						return err
					}

					outputResults(rs, c)
					return nil
				},
			},
		},
		Action: func(c *cli.Context) error {
			d := currentDate(c)
			if s := c.Args().First(); s != "" {
				t, err := time.Parse("0102", s)
				if err != nil {
					return err
				}
				d.Month, d.Day = int(t.Month()), t.Day()
			}

			result, err := h.WrapResult(getLunarResult(d, c.Bool("reverse")))
			if err != nil {
				return err
			}
			outputResults([]*alias.Result{result}, c)
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func outputResults(rs []*alias.Result, c *cli.Context) {
	dateFormat := c.String("format")
	sort.Slice(rs, func(i, j int) bool {
		di, dj := rs[i].Date, rs[j].Date
		if di.Year != dj.Year {
			return di.Year < dj.Year
		}
		if di.Month != dj.Month {
			return di.Month < dj.Month
		}

		return di.Day < dj.Day
	})

	data := make([][]string, len(rs))
	now := time.Now()

	for i, r := range rs {
		// calc timedelta
		var timedeltaStr string
		timedelta := int(now.Sub(r.Date.Time()).Hours() / 24)
		switch {
		case timedelta < 0:
			timedeltaStr = fmt.Sprintf("还有 %d 天", -timedelta)
		case timedelta == 0:
			timedeltaStr = "今天"
		case timedelta > 0:
			timedeltaStr = fmt.Sprintf("已过去 %d 天", timedelta)
		}

		row := []string{
			r.Date.Time().Format(dateFormat),
			r.LunarDate.Time().Format(dateFormat),
			r.WeekdayRaw,
			timedeltaStr,
			r.SolarTerm,
		}

		aliases := []string{}
		tagMap := map[string]bool{}
		tags := []string{}
		for _, a := range r.Aliases {
			aliases = append(aliases, a.Name)
			for _, t := range a.Tags {
				if !tagMap[t] {
					tagMap[t] = true
					tags = append(tags, t)
				}
			}
		}
		row = append(row, strings.Join(aliases, ","))
		row = append(row, strings.Join(tags, ","))
		data[i] = row
	}

	table := tablewriter.NewWriter(os.Stdout)
	header := []string{"公历", "农历", "星期", "距今", "节气", "别名", "标签"}
	table.SetHeader(header)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.AppendBulk(data)
	table.Render()
}

func getLunarResult(d lunar.Date, reverse bool) (*lunar.Result, error) {
	var (
		result *lunar.Result
		err    error
	)
	if reverse {
		result, err = lunar.LunarDateToDate(d)
	} else {
		result, err = lunar.DateToLunarDate(d)
	}

	return result, err
}

func currentDate(c *cli.Context) lunar.Date {
	d := lunar.DateByTime(time.Now().In(CST))
	d.Year = c.Int("year")

	return d
}

func mustUserHomeDir() string {
	d, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	return d
}
