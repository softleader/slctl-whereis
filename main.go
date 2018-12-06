package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
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
	cmd := &cobra.Command{
		Use:   "whereis NAME",
		Short: "slctl whereis",
		Long:  "to find where the member is",
		RunE: func(cmd *cobra.Command, args []string) error {
			var ok bool
			if c.token, ok = os.LookupEnv("SL_TOKEN"); !ok {
				return fmt.Errorf("require GitHub access token to run the plugin")
			}
			c.verbose, _ = strconv.ParseBool(os.Getenv("SL_VERBOSE"))

			if len(args) > 0 {
				c.name = args[0]
			}
			return c.run()
		},
	}
	f := cmd.Flags()
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
	url := fmt.Sprintf("%s/api/whereis?%s", api, c.queryString())
	req, err := http.NewRequest("GET", url, &buf)
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
	json.NewDecoder(resp.Body).Decode(&w)

	fmt.Printf("%s\n", w.summary())
	table := uitable.New()
	table.AddRow("PLACE", "NAME", "DATE", "WHERE TO")
	for _, c := range w.Content {
		table.AddRow(c.place(), c.name(), c.date(), c.whereTo())
	}
	fmt.Println(table)
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
			qss = append(qss, k+"="+v)
		}
	}
	return strings.Join(qss, "&")
}

func (c *whereisCmd) limit() string {
	return fmt.Sprintf("%s/%s", c.size, c.page)
}
