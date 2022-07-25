package cli

import (
	"context"
	"fmt"
)

type Bit2EvmCommand struct {
	args *Args
}

const (
	bit2evmCommand     = "bit2evm"
	bit2evmDescription = "generate ETH ublic and private key from BTC private WIF"
)

func newBit2EvmCommand() Command {
	return Command{
		Name:        bit2evmCommand,
		Description: bit2evmDescription,
		Exec:        &Bit2EvmCommand{},
	}
}

func (c *Bit2EvmCommand) bit2evmMasterWallet() error {
	fmt.Println("bit2evm master wallet")
	return nil
}

func (c *Bit2EvmCommand) bit2evmBTCWallet() error {
	fmt.Println("bit2evm BTC wallet")
	return nil
}

func (c *Bit2EvmCommand) ExecCommand(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return ErrorFromString(fmt.Sprintf("%s: no subcommand passed", bit2evmCommand))
	}
	c.args = &Args{args}

	return ErrorFromString(fmt.Sprintf("%s: invalid subcommand passed", bit2evmCommand))
}
