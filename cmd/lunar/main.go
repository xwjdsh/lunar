package main

import (
	"fmt"
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
	"github.com/xwjdsh/lunar/config"
)

var _CST = time.FixedZone("CST", 3600*8)

func main() {
	h := alias.NewHandler(lunar.New())
	beforeFunc := func(c *cli.Context) error {
		fp := c.String("config")
		conf, err := config.Init(fp, false)
		if err != nil {
			return err
		}
		h.LoadAlias(conf.Aliases)
		return nil
	}
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
				Value:   path.Join(mustUserHomeDir(), ".config/lunar/lunar.yml"),
				Usage:   "Custom config path",
			},
			&cli.IntFlag{
				Name:    "year",
				Aliases: []string{"y"},
				Value:   time.Now().In(_CST).Year(),
				Usage:   "Target year",
			},
			&cli.BoolFlag{
				Name:    "reverse",
				Aliases: []string{"r"},
				Usage:   "Reverse mode, query date by lunar date",
			},
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
				Usage:  "Show alias date info",
				Before: beforeFunc,
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
				Before:  beforeFunc,
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
			{
				Name:    "config",
				Aliases: []string{"c"},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "default",
						Aliases: []string{"d"},
						Usage:   "Show default config",
					},
				},
				Usage: "Display config",
				Action: func(c *cli.Context) error {
					conf, err := config.Init(c.String("config"), c.Bool("default"))
					if err != nil {
						return err
					}
					data, err := conf.Marshal()
					if err != nil {
						return err
					}
					fmt.Println(string(data))
					return nil
				},
			},
		},
		Action: func(c *cli.Context) error {
			if err := beforeFunc(c); err != nil {
				return err
			}
			d := currentDate(c)
			if s := c.Args().First(); s != "" {
				t, err := time.Parse("0102", s)
				if err != nil {
					return err
				}
				d.Month, d.Day = int(t.Month()), t.Day()
			}

			results, err := h.WrapResults(getLunarResult(d, c.Bool("reverse")))
			if err != nil {
				return err
			}
			outputResults(results, c)
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
	now := currentDate(nil).Time()

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

		leapMonthStr := ""
		if r.LunarDate.IsLeapMonth {
			leapMonthStr = " (闰月)"
		}
		row := []string{
			r.Date.Time().Format(dateFormat),
			r.LunarDate.Time().Format(dateFormat) + leapMonthStr,
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
	header := []string{"阳历", "阴历", "星期", "距今", "节气", "别名", "标签"}
	table.SetHeader(header)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.AppendBulk(data)
	table.Render()
}

func getLunarResult(d lunar.Date, reverse bool) ([]*lunar.Result, error) {
	results := []*lunar.Result{}
	if reverse {
		r1, err := lunar.Calendar(lunar.NewLunarDate(d, false))
		if err == nil {
			results = append(results, r1)
		}

		if err != nil && err != lunar.ErrNotFound {
			return nil, err
		}

		r2, err := lunar.Calendar(lunar.NewLunarDate(d, true))
		if err == nil {
			results = append(results, r2)
		}

		if err != nil && err != lunar.ErrNotFound {
			return nil, err
		}
	} else {
		r, err := lunar.Calendar(d)
		if err == nil {
			results = []*lunar.Result{r}
		}
		if err != nil && err != lunar.ErrNotFound {
			return nil, err
		}
	}

	return results, nil
}

func currentDate(c *cli.Context) lunar.Date {
	d := lunar.DateByTime(time.Now().In(_CST))
	if c != nil {
		d.Year = c.Int("year")
	}

	return d
}

func mustUserHomeDir() string {
	d, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	return d
}
