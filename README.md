# lunar

`lunar` 是一个命令行下的阴历阳历查询转换工具 (支持时间段范围[1901~2100])。

![](https://raw.githubusercontent.com/xwjdsh/lunar/main/screenshot.png)

## 安装
### Homebrew
```
brew tap xwjdsh/tap
brew install xwjdsh/tap/lunar
```
### Go
> 仅支持 `go1.16` 及以上版本

if go_version == 1.16.x
```
go get -u github.com/xwjdsh/lunar/cmd/lunar
```
else
```
go install github.com/xwjdsh/lunar/cmd/lunar
```

### Docker
```
alias lunar='docker run -it --rm wendellsun/lunar'

# 挂载自定义配置
alias lunar='docker run -it -v $PATH_TO_YOUR_CONFIG:/root/.config/lunar/lunar.yml --rm wendellsun/lunar'
```

### Manual
从 [releases](https://github.com/xwjdsh/lunar/releases) 下载对应的可执行文件并将其放到 PATH 环境变量对应的路径中。

## 使用
```
> lunar -h
NAME:
   lunar - lunar is a command line tool for conversion between Gregorian calendar and lunar calendar.(1901~2100)

USAGE:
   lunar [global options] command [command options] [arguments...]

COMMANDS:
   alias, a        Show alias date info
   solar-term, st  Get solar term info
   config, c       Display config
   help, h         Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --format value, -f value  Output date format (default: "2006-01-02")
   --config value, -c value  Custom config path (default: "$HOME/.config/lunar/lunar.yml")
   --year value, -y value    Target year (default: $THIS_YEAR)
   --reverse, -r             Reverse mode, query date by lunar date (default: false)
   --help, -h                show help (default: false)
```

### 阳历转阴历
```
> # lunar -y 2022       # 指定年份，月日为今日
> # lunar -y 2022 0701  # 指定年月日
> lunar                 # 不带参数年月日为今日
```
|    阳历    |    阴历    |  星期  | 距今 | 节气 | 别名 | 标签 |
|  ----  | ----  |  ----  | ----  |  ----  | ----  |  ----  |
| 2022-01-26 | 2021-12-24 | 星期三 | 今天 |      |      |      |


### 阴历转阳历
```
> # lunar -r -y 2022      # 查询阴历，指定年份
> # lunar -r -y 2022 0701 # 查询阴历，指定年月日
> lunar -r                # 查询阴历，不带参数年月日为阴历今日
```
|    阳历    |    阴历    |  星期  |    距今    | 节气 | 别名 | 标签 |
|  ----  | ----  |  ----  | ----  |  ----  | ----  |  ----  |
| 2022-02-26 | 2022-01-26 | 星期六 | 还有 31 天 |      |      |      |

### 自定义配置别名
```
> # lunar config -d                            # 显示默认配置，默认加入了一些常见节日的别名
> lunar config -d > ~/.config/lunar/lunar.yml  # 导出默认配置，自定义修改
> # lunar config                               # 显示当前配置
```
例如修改为如下，
```yml
aliases:
    - name: xx的生日
      disable: false
      date:
        year: 0
        month: 5
        day: 7
      is_lunar_date: true
      leap_month_limit: 0
      tags:
        - birthday
```
|    阳历    |    阴历    |  星期  |    距今     | 节气 |   别名   |   标签   |
|  ----  | ----  |  ----  | ----  |  ----  | ----  |  ----  |
| 2022-06-05 | 2022-05-07 | 星期日 | 还有 130 天 |      | xx的生日 | birthday |


### 查询别名
```
> # lunar a 春节 中秋   # 查询指定别名
> # lunar -y 2022 a    # 指定年份
> lunar a              # 列出所有别名日期
```
|    阳历    |    阴历    |  星期  |     距今     | 节气 | 别名 |  标签   |
|  ----  | ----  |  ----  | ----  |  ----  | ----  |  ----  |
| 2022-01-01 | 2021-11-29 | 星期六 | 已过去 25 天 |      | 元旦 | holiday |
| 2022-01-10 | 2021-12-08 | 星期一 | 已过去 16 天 |      | 腊八 |         |
| 2022-02-01 | 2022-01-01 | 星期二 | 还有 6 天    |      | 春节 | holiday |
| 2022-02-15 | 2022-01-15 | 星期二 | 还有 20 天   |      | 元宵 |         |
| 2022-04-04 | 2022-03-04 | 星期一 | 还有 68 天   |      | 清明 | holiday |
| 2022-05-01 | 2022-04-01 | 星期日 | 还有 95 天   |      | 劳动 | holiday |
| 2022-06-03 | 2022-05-05 | 星期五 | 还有 128 天  |      | 端午 | holiday |
| 2022-08-04 | 2022-07-07 | 星期四 | 还有 190 天  |      | 七夕 |         |
| 2022-08-12 | 2022-07-15 | 星期五 | 还有 198 天  |      | 中元 |         |
| 2022-09-10 | 2022-08-15 | 星期六 | 还有 227 天  |      | 中秋 | holiday |
| 2022-10-01 | 2022-09-06 | 星期六 | 还有 248 天  |      | 国庆 | holiday |
| 2022-10-04 | 2022-09-09 | 星期二 | 还有 251 天  |      | 重阳 |         |
| 2022-12-30 | 2022-12-08 | 星期五 | 还有 338 天  |      | 腊八 |         |

### 查询标签
```
> lunar a -t birthday # 查询自定义标签
> lunar a -t holiday  # 查询标签
```
|    阳历    |    阴历    |  星期  |     距今     | 节气 | 别名 |  标签   |
|  ----  | ----  |  ----  | ----  |  ----  | ----  |  ----  |
| 2022-01-01 | 2021-11-29 | 星期六 | 已过去 25 天 |      | 元旦 | holiday |
| 2022-02-01 | 2022-01-01 | 星期二 | 还有 6 天    |      | 春节 | holiday |
| 2022-04-04 | 2022-03-04 | 星期一 | 还有 68 天   |      | 清明 | holiday |
| 2022-05-01 | 2022-04-01 | 星期日 | 还有 95 天   |      | 劳动 | holiday |
| 2022-06-03 | 2022-05-05 | 星期五 | 还有 128 天  |      | 端午 | holiday |
| 2022-09-10 | 2022-08-15 | 星期六 | 还有 227 天  |      | 中秋 | holiday |
| 2022-10-01 | 2022-09-06 | 星期六 | 还有 248 天  |      | 国庆 | holiday |


### 查询节气
```
> # lunar -y 2022 st # 指定年份
> # lunar st         # 查询所有节气
> lunar st 冬至    # 查询指定节气
```
|    阳历    |    阴历    |  星期  |    距今     | 节气 | 别名 | 标签 |
|  ----  | ----  |  ----  | ----  |  ----  | ----  |  ----  |
| 2022-12-22 | 2022-11-29 | 星期四 | 还有 330 天 | 冬至 |      |      |

## 协议
[MIT License](https://github.com/xwjdsh/lunar/blob/main/LICENSE)
