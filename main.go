package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
	"net/http"
	"net/url"
	"os"
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
`
)

var (
	api    = "http://support.softleader.com.tw/softleader-holiday"
	layout = "2006-01-02"
)

type whereisCmd struct {
	verbose bool
	token   string
	name    string
	size    string
	page    string
	place   string
	from    string
	to      string
}

func main() {
	c := whereisCmd{}
	c.verbose, _ = strconv.ParseBool(os.Getenv("SL_VERBOSE"))

	cmd := &cobra.Command{
		Use:   "slctl whereis NAME",
		Short: "find out where SoftLeader the member is",
		Long:  longDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if offline, _ := strconv.ParseBool(os.Getenv("SL_OFFLINE")); offline {
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
	f := cmd.Flags()
	f.BoolVarP(&c.verbose, "verbose", "v", c.verbose, "enable verbose output, Overrides $SL_VERBOSE")
	f.StringVar(&c.token, "token", "$SL_TOKEN", "github access token. Overrides $SL_TOKEN")
	f.StringVarP(&c.size, "size", "s", "20", "determine output size")
	f.StringVarP(&c.page, "page", "p", "1", "determine output page")
	f.StringVarP(&c.from, "from", "f", time.Now().Format(layout), "filter the specified date from")
	f.StringVarP(&c.to, "to", "t", "", "filter the specified date to")
	f.StringVarP(&c.place, "place", "P", "", "specified the place")

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func (c *whereisCmd) run() (err error) {
	var buf bytes.Buffer
	s := fmt.Sprintf("%s/api/whereis?%s", api, c.queryString())
	req, err := http.NewRequest("GET", s, &buf)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	if c.verbose {
		fmt.Printf("%s %s\n", req.Method, req.URL)
		fmt.Printf("Header: %v\n", req.Header)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	w := whereis{}
	if err = json.NewDecoder(resp.Body).Decode(&w); err != nil {
		return err
	}
	if len(w.Content) == 0 {
		fmt.Printf("No search results")
	} else {
		fmt.Printf("%s\n", w.summary())
		table := uitable.New()
		table.AddRow("PLACE", "NAME", "DATE", "WHERE TO")
		for _, c := range w.Content {
			table.AddRow(c.place(), c.name(), c.date(), c.whereTo())
		}
		fmt.Println(table)
	}
	return
}

func (c *whereisCmd) queryString() string {
	qs := make(map[string]string)
	qs["n"] = c.name
	qs["l"] = c.limit()
	qs["p"] = c.place
	qs["f"] = parse(c.from).Format(layout)
	qs["t"] = parse(c.to).Format(layout)

	var qss []string
	for k, v := range qs {
		if v != "" {
			qss = append(qss, k+"="+url.QueryEscape(v))
		}
	}
	return strings.Join(qss, "&")
}

func (c *whereisCmd) limit() string {
	return fmt.Sprintf("%s/%s", c.size, c.page)
}
