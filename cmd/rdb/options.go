package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/codegangsta/cli"
)

func init() {
	app.Commands = append(app.Commands, cli.Command{
		Name:   "options",
		Usage:  "produce predefined set of options for various purposes",
		Action: optionsBulk,
	})
}

func optionsBulk(c *cli.Context) {
	DefaultOptions.Update(c)
	out, err := json.MarshalIndent(DefaultOptions, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(out))
}
