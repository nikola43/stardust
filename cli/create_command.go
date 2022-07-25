package cli

import (
	"context"
	"fmt"
)

type CreateCommand struct {
	args *Args
}

const (
	createCommand     = "create"
	createDescription = "create blockchain wallet"

	createMasterWallet = "master-wallet"
	createBTCWallet    = "btc-wallet"
	createEthWallet    = "eth-wallet"
	createNetwork      = "network"
	createNode         = "node"
)

func newCreateCommand() Command {
	return Command{
		Name:        createCommand,
		Description: createDescription,
		Exec:        &CreateCommand{},
	}
}

func (c *CreateCommand) createMasterWallet() error {
	fmt.Println("create master wallet")
	return nil
}

func (c *CreateCommand) createBTCWallet() error {
	fmt.Println("create BTC wallet")
	return nil
}

func (c *CreateCommand) createEthWallet() error {
	fmt.Println("create ETH wallet")
	return nil
}

func (c *CreateCommand) createNetwork() error {
	networkTypeNumbder := c.args.pop()
	if networkTypeNumbder == "" {
		fmt.Println("create network")
		return nil
	}

	fmt.Printf("create network type with number: %s", networkTypeNumbder)
	return nil
}

func (c *CreateCommand) createNode() error {
	nodeHash := c.args.pop()
	if nodeHash == "" {
		return ErrorFromString(fmt.Sprintf("%s-node: invalid node hash", createCommand))
	}

	fmt.Println("create node")
	return nil
}

func (c *CreateCommand) ExecCommand(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return ErrorFromString(fmt.Sprintf("%s: no subcommand passed", createCommand))
	}
	c.args = &Args{args}

	subcommand := c.args.pop()
	switch subcommand {
	case createMasterWallet:
		return c.createMasterWallet()
	case createBTCWallet:
		return c.createBTCWallet()
	case createEthWallet:
		return c.createEthWallet()
	case createNode:
		return c.createNode()
	case createNetwork:
		return c.createNetwork()
	}

	return ErrorFromString(fmt.Sprintf("%s: invalid subcommand passed", createCommand))
}
