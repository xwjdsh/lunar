package main

import (
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"

	"github.com/xwjdsh/lunar"
)

var CST = time.FixedZone("CST", 3600*8)

func main() {
	app := &cli.App{
		Name:  "lunar",
		Usage: "lunar is a command line tool for conversion between Gregorian calendar and lunar calendar.(1901~2100)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"f"},
				Value:   "2006-01-02",
				Usage:   "output date format",
			},
			&cli.IntFlag{
				Name:    "year",
				Aliases: []string{"y"},
				Value:   time.Now().In(CST).Year(),
				Usage:   "target year",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "holidays",
				Aliases: []string{"h"},
				Usage:   "show holidays date info",
				Action: func(c *cli.Context) error {
					var results []*lunar.Result
					d := currentDate(c)
					results, err := lunar.Holidays(d.Year)
					if err != nil {
						return err
					}

					outputResults(results, c)
					return nil
				},
			},
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
				Usage: "show alias date info",
				Action: func(c *cli.Context) error {
					d := currentDate(c)
					var (
						results []*lunar.Result
						err     error
					)
					if c.Args().Len() >= 1 {
						results, err = lunar.GetAliases(d.Year, c.Args().Slice()...)
					} else {
						results, err = lunar.GetAliases(d.Year)
					}
					if err != nil {
						return err
					}

					outputResults(results, c)
					return nil
				},
			},
			{
				Name:    "reverse",
				Aliases: []string{"r"},
				Usage:   "reverse mode, query date by lunar date",
				Action: func(c *cli.Context) error {
					return queryAndDisplay(c, true)
				},
			},
			{
				Name:    "solar-term",
				Aliases: []string{"st"},
				Usage:   "get solar term info by name",
				Action: func(c *cli.Context) error {
					d := currentDate(c)
					var (
						rs  []*lunar.Result
						err error
					)
					if c.Args().Len() >= 1 {
						var r *lunar.Result
						r, err = lunar.GetSolarTerm(c.Args().First(), d.Year)
						if err == nil {
							rs = append(rs, r)
						}
					} else {
						rs, err = lunar.GetSolarTerms(d.Year)
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
			return queryAndDisplay(c, false)
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func outputResults(rs []*lunar.Result, c *cli.Context) {
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
	for i, r := range rs {
		row := []string{
			r.Date.Time().Format(dateFormat),
			r.LunarDate.Time().Format(dateFormat),
			r.WeekdayRaw,
			r.SolarTerm,
		}

		aliases := []string{}
		for _, a := range r.Aliases {
			aliases = append(aliases, a.Name)
		}
		row = append(row, strings.Join(aliases, ","))
		data[i] = row
	}

	table := tablewriter.NewWriter(os.Stdout)
	header := []string{"公历", "农历", "星期", "节气", "别名"}
	table.SetHeader(header)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.AppendBulk(data)
	table.Render()
}

func getLunarResult(d lunar.Date, reverse bool) (*lunar.Result, lunar.Date, error) {
	var (
		result     *lunar.Result
		err        error
		resultDate lunar.Date
	)
	if reverse {
		result, err = lunar.LunarDateToDate(d)
		if err == nil {
			resultDate = result.Date
		}
	} else {
		result, err = lunar.DateToLunarDate(d)
		if err == nil {
			resultDate = result.LunarDate
		}
	}

	return result, resultDate, err
}

func currentDate(c *cli.Context) lunar.Date {
	d := lunar.DateByTime(time.Now().In(CST))
	d.Year = c.Int("year")

	return d
}

func queryAndDisplay(c *cli.Context, reverse bool) error {
	d := currentDate(c)
	if s := c.Args().First(); s != "" {
		t, err := time.Parse("0102", s)
		if err != nil {
			return err
		}
		d.Month, d.Day = int(t.Month()), t.Day()
	}

	result, _, err := getLunarResult(d, reverse)
	if err != nil {
		return err
	}
	outputResults([]*lunar.Result{result}, c)
	return nil
}
