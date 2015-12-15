package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/codegangsta/cli"
	"github.com/ingeniosity/rdb"
)

func init() {
	app.Commands = append(app.Commands, cli.Command{
		Name:   "options",
		Usage:  "produce predefined set of options for various purposes",
		Action: optionsBulk,
	})
}

func optionsBulk(c *cli.Context) {
	dbOptions := rdb.NewDefaultOptions()
	defaultFlags.setOptions(dbOptions, c)
	out, err := json.MarshalIndent(DefaultOptions, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(out))
}
