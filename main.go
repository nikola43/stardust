package main

import (
	"fmt"
	"os"

	"github.com/nikola43/stardust/cli"
	"github.com/nikola43/stardust/config"
)

var (
	configFile string
	update     string
	etcd       bool
	cfg        *config.Config
)

func main() {

	if err := cli.New().Run(); err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	os.Exit(0)
}
