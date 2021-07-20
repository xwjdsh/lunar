# lunar

```
> # today to lunar date
> go run ./cmd/lunar
2021-06-11

> # specify output format
> go run ./cmd/lunar -f "2006年01月02日"
2021年06月11日

> # specify month and day, this year
> go run ./cmd/lunar 0803
2021-06-25

> # specify year, month and day
> go run ./cmd/lunar 20100303
2010-01-18
```
