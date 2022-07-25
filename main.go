package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/mdp/qrterminal"
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
	ShowPaymentQr()
	if err := cli.New().Run(); err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
}

func ShowPaymentQr() {
	const Red = "\033[44m  \033[0m"
	const BLUE = "\033[43m  \033[0m"

	config := qrterminal.Config{
		Level:     qrterminal.M,
		Writer:    os.Stdout,
		BlackChar: Red,
		WhiteChar: BLUE,
		QuietZone: 1,
	}
	qrterminal.GenerateWithConfig("0x6d5F00aE01F715D3082Ad40dfB5c18A1a35d3A17", config)
	fmt.Println()
	fmt.Println(color.CyanString("Send 1 ETH to: "), color.YellowString("0x6d5F00aE01F715D3082Ad40dfB5c18A1a35d3A17"))
	fmt.Println()
}
