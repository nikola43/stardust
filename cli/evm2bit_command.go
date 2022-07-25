package cli

import (
	"context"
	"fmt"
)

type Evm2BitCommand struct {
	args *Args
}

const (
	evm2bitCommand     = "evm2bit"
	evm2bitDescription = "derive BTC public and private key from ETH private key"
)

func newEvm2BitCommand() Command {
	return Command{
		Name:        evm2bitCommand,
		Description: evm2bitDescription,
		Exec:        &Evm2BitCommand{},
	}
}

func (c *Evm2BitCommand) evm2bitMasterWallet() error {
	fmt.Println("evm2bit master wallet")
	return nil
}

func (c *Evm2BitCommand) evm2bitBTCWallet() error {
	fmt.Println("evm2bit BTC wallet")
	return nil
}

func (c *Evm2BitCommand) ExecCommand(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return ErrorFromString(fmt.Sprintf("%s: no subcommand passed", evm2bitCommand))
	}
	c.args = &Args{args}

	return ErrorFromString(fmt.Sprintf("%s: invalid subcommand passed", evm2bitCommand))
}
