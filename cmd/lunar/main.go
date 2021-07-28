package main

import (
	"log"
	"os"
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
				Value:   0,
				Usage:   "target year",
			},
			&cli.BoolFlag{
				Name:    "reverse",
				Aliases: []string{"r"},
				Value:   false,
				Usage:   "reverse mode, query by lunar date",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "holidays",
				Aliases: []string{"h"},
				Usage:   "show holidays date info",
				Action: func(c *cli.Context) error {
					var results []*lunar.Result
					d := currentDate(c.Int("year"))
					results, err := lunar.Holidays(d.Year)
					if err != nil {
						return err
					}

					outputResults(results, c.String("format"))
					return nil
				},
			},
			{
				Name:    "aliases",
				Aliases: []string{"h"},
				Usage:   "show aliases date info",
				Action: func(c *cli.Context) error {
					var results []*lunar.Result
					d := currentDate(c.Int("year"))
					results, err := lunar.Aliases(d.Year)
					if err != nil {
						return err
					}

					outputResults(results, c.String("format"))
					return nil
				},
			},
		},
		Action: func(c *cli.Context) error {
			d := currentDate(c.Int("year"))
			if s := c.Args().First(); s != "" {
				if v, ok := lunar.GetAlias(s); ok {
					d.Month, d.Day = v.Date.Month, v.Date.Day
				} else {
					t, err := time.Parse("0102", s)
					if err != nil {
						log.Fatal(err)
					}
					d.Month, d.Day = int(t.Month()), t.Day()
				}
			}

			result, _, err := getLunarResult(d, c.Bool("reverse"))
			if err != nil {
				return err
			}
			outputResults([]*lunar.Result{result}, c.String("format"))
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func outputResults(rs []*lunar.Result, dateFormat string) {
	data := make([][]string, len(rs))
	showAliases := false
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
		if len(aliases) > 0 {
			row = append(row, strings.Join(aliases, ","))
			showAliases = true
		}
		data[i] = row
	}

	table := tablewriter.NewWriter(os.Stdout)
	header := []string{"公历", "农历", "星期", "节气"}
	if showAliases {
		header = append(header, "别名")
	}
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

func currentDate(year int) lunar.Date {
	d := lunar.DateByTime(time.Now().In(CST))
	if year != 0 {
		d.Year = year
	}

	return d
}
