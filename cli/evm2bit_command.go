package cli

import (
	"context"
	"fmt"
	"log"

	"github.com/fatih/color"
	wallet "github.com/nikola43/stardust/wallet"
)

type Evm2BitCommand struct {
	args *Args
}

const (
	evm2bitCommand     = "evm2bit"
	evm2bitDescription = "generate BTC public and private key from ETH private key"
)

func newEvm2BitCommand() Command {
	return Command{
		Name:        evm2bitCommand,
		Description: evm2bitDescription,
		Exec:        &Evm2BitCommand{},
	}
}

func (c *Evm2BitCommand) ExecCommand(ctx context.Context, args []string) error {
	c.args = &Args{args}
	if len(args) == 0 {
		return ErrorFromString(fmt.Sprintf("file not found"))
	}

	w, err := wallet.GenerateETHWalletFromPlainPrivateKey(args[1])
	if err != nil {
		log.Fatal(err)
	}

	btcDerivedPrivateKey := HashValue(w.PrivateKey)
	btcDerivedPublicKey, err := GenerateAddressFromPlainPrivateKey(btcDerivedPrivateKey)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(color.CyanString("BTC Derived Public Key: "), color.YellowString(btcDerivedPublicKey.Hex()))
	fmt.Println(color.CyanString("BTC Derived Private Key: "), color.YellowString(btcDerivedPrivateKey))
	fmt.Println()

	return nil
}
