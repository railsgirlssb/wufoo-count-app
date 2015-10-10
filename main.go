package main

import (
	"github.com/railsgirlssb/wufoo-count-app/Godeps/_workspace/src/github.com/go-martini/martini"
	"github.com/railsgirlssb/wufoo-count-app/Godeps/_workspace/src/github.com/martini-contrib/cors"
	"github.com/railsgirlssb/wufoo-count-app/Godeps/_workspace/src/github.com/martini-contrib/render"
	"github.com/railsgirlssb/wufoo-count-app/Godeps/_workspace/src/gopkg.in/resty.v0"

	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type WufooConfig struct {
	Account  string
	ApiKey   string
	Password string
	FormIds  []string
}

var wufooConfig WufooConfig

func count() (int, error) {
	count := 0

	type Result struct {
		EntryCount string `json:"EntryCount"`
	}

	for _, formId := range wufooConfig.FormIds {
		resp, err := resty.R().
			SetHeader("Accept", "application/json").
			SetBasicAuth(wufooConfig.ApiKey, wufooConfig.Password).
			Get(fmt.Sprintf("https://%s.wufoo.com/api/v3/forms/%s/entries/count.json", wufooConfig.Account, formId))
		if err != nil {
			return count, err
		}

		var result Result
		err = json.Unmarshal(resp.Body, &result)
		if err != nil {
			return count, err
		}
		entryCount, _ := strconv.Atoi(result.EntryCount)

		count += entryCount
	}

	return count, nil
}

func main() {
	port := os.Getenv("PORT")
	if len(port) < 1 {
		port = "8080"
	}

	wufooConfig.Account = os.Getenv("WUFOO_ACCOUNT")
	wufooConfig.ApiKey = os.Getenv("WUFOO_API_KEY")
	wufooConfig.Password = "any"
	wufooConfig.FormIds = strings.Split(os.Getenv("WUFOO_FORM_IDS"), ",")

	m := martini.Classic()
	m.Use(render.Renderer())
	m.Use(cors.Allow(&cors.Options{
		AllowAllOrigins: true,
	}))

	m.Get("/", func(r render.Render) {
		count, err := count()
		if err != nil {
			r.JSON(200, map[string]interface{}{"error": "can't fetch information"})
		} else {
			r.JSON(200, map[string]interface{}{"count": count})
		}
	})
	m.RunOnAddr(":" + port)
}
