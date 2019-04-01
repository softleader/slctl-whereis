package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
	"gopkg.in/resty.v1"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	longDesc = `
查看當日公司員工在哪兒

	$ slctl whereis

可以使用員工姓名(模糊查詢)過濾資料

	$ slctl whereis matt

傳入 '--from' 或 '--to' 可以用日期區間過濾
日期格式為年月日, 支援格式可參考: https://github.com/araddon/dateparse
同時也支援少數自然語言, 如: 'today', 'yesterday', 'tomorrow'

	$ slctl whereis -f yesterday
	$ slctl whereis matt -f 20181201 -t 20181203

查詢結果預設顯示第一頁, 每頁顯示 20 筆資料
可以傳入 '--page' 指定頁數或傳入 '--size' 指定一頁幾筆 (一頁筆數放很大則等於不分頁)

	$ slctl whereis -s 1000


傳入 '--grep'' 可以針對顯示的結果, 再做一次 regex 過濾, 類似 unix 系統的 'grep' 但可以跨系統使用

	# 查詢 2019 年起到當日的資料, 但只顯示每個星期一的紀錄
	$ slctl whereis matt -f 20190101 -t today --grep mon
`
)

var (
	api    = "https://support.softleader.com.tw/softleader-holiday"
	layout = "2006-01-02"
)

type whereisCmd struct {
	offline bool
	verbose bool
	token   string
	out     io.Writer
	cli     string
	version string
	name    string
	size    string
	page    string
	place   string
	from    string
	to      string
	grep    string
}

func main() {
	c := whereisCmd{}
	cmd := &cobra.Command{
		Use:   "slctl whereis NAME",
		Short: "find out where the SoftLeader member is",
		Long:  longDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if c.offline {
				return fmt.Errorf("can not run the command in offline mode")
			}
			if c.token = os.ExpandEnv(c.token); c.token == "" {
				return fmt.Errorf("require GitHub access token to run the command")
			}
			if len := len(args); len > 0 {
				if len > 1 {
					return errors.New("this command does not accept more than 1 arguments")
				}
				c.name = args[0]
			}
			return c.run()
		},
	}

	c.out = cmd.OutOrStdout()
	c.cli = os.Getenv("SL_CLI")
	c.version = os.Getenv("SL_VERSION")
	c.offline, _ = strconv.ParseBool(os.Getenv("SL_OFFLINE"))
	c.verbose, _ = strconv.ParseBool(os.Getenv("SL_VERBOSE"))

	f := cmd.Flags()
	f.BoolVarP(&c.offline, "offline", "o", c.offline, "work offline, Overrides $SL_OFFLINE")
	f.BoolVarP(&c.verbose, "verbose", "v", c.verbose, "enable verbose output, Overrides $SL_VERBOSE")
	f.StringVar(&c.token, "token", "$SL_TOKEN", "github access token. Overrides $SL_TOKEN")
	f.StringVarP(&c.size, "size", "s", "20", "determine output size")
	f.StringVarP(&c.page, "page", "p", "1", "determine output page")
	f.StringVarP(&c.from, "from", "f", time.Now().Format(layout), "filter the specified date from")
	f.StringVarP(&c.to, "to", "t", "", "filter the specified date to")
	f.StringVarP(&c.place, "place", "P", "", "specify the place")
	f.StringVar(&c.grep, "grep", ".*", "specify the grep regex pattern")

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func (c *whereisCmd) run() (err error) {
	resp, err := resty.R().
		SetQueryParams(c.queryParams()).
		SetAuthToken(c.token).
		SetHeader("User-Agent", fmt.Sprintf("%s/%s %s/%s", c.cli, c.version, "whereis", ver())).
		Get(fmt.Sprintf("%s/api/whereis", api))
	if c.verbose {
		fmt.Fprintf(c.out, "> %v %v\n", resp.Request.Method, resp.Request.URL)
		for k, v := range resp.Request.Header {
			fmt.Fprintf(c.out, "> %v: %v\n", k, strings.Join(v, ", "))
		}
		fmt.Fprintln(c.out, ">")
		fmt.Fprintf(c.out, "< Error: %v\n", err)
		fmt.Fprintf(c.out, "< Status Code: %v\n", resp.StatusCode())
		fmt.Fprintf(c.out, "< Status: %v\n", resp.Status())
		fmt.Fprintf(c.out, "< Time: %v\n", resp.Time())
		fmt.Fprintf(c.out, "< Received At: %v\n", resp.ReceivedAt())
		fmt.Fprintf(c.out, "%v\n", resp)
	}
	if err != nil {
		return
	}
	if !resp.IsSuccess() {
		return fmt.Errorf(`expected response status code 2xx, but got %v.
Use the '--verbose' flag to see the full stacktrace
`, resp.StatusCode())
	}

	g, err := regexp.Compile(fmt.Sprintf(`(?i)%s`, c.grep))
	if err != nil {
		return err
	}
	err = print(c.out, resp.Body(), g)
	return
}

func print(out io.Writer, data []byte, grep *regexp.Regexp) (err error) {
	w := whereis{}
	if err = json.Unmarshal(data, &w); err != nil {
		return fmt.Errorf("unable to unmarshal response: %s", err)
	}
	if len(w.Content) == 0 {
		fmt.Fprintf(out, "No search results")
	} else {
		fmt.Fprintf(out, "%s\n", w.summary())
		table := uitable.New()
		table.AddRow("PLACE", "NAME", "DATE", "WHERE TO")
		for _, c := range w.Content {
			row := uitable.NewRow(c.place(), c.name(), c.date(), c.whereTo())
			if grep.MatchString(row.String()) {
				table.Rows = append(table.Rows, row)
			}
		}
		fmt.Fprintln(out, table)
	}
	return
}

func (c *whereisCmd) queryParams() (qp map[string]string) {
	qp = make(map[string]string)
	if v := c.name; v != "" {
		qp["n"] = v
	}
	qp["l"] = c.limit()
	if v := c.place; v != "" {
		qp["p"] = c.place
	}
	qp["f"] = parse(c.from).Format(layout)
	if v := c.to; v != "" {
		qp["t"] = parse(v).Format(layout)
	}
	return
}

func (c *whereisCmd) limit() string {
	return fmt.Sprintf("%s/%s", c.size, c.page)
}
