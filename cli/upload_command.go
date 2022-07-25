package cli

import (
	"context"
	"fmt"
)

type UploadCommand struct {
	args *Args
}

const (
	uploadCommand     = "upload"
	uploadDescription = "derive BTC public and private key from ETH private key"
)

func newUploadCommand() Command {
	return Command{
		Name:        uploadCommand,
		Description: uploadDescription,
		Exec:        &UploadCommand{},
	}
}

func (c *UploadCommand) uploadMasterWallet() error {
	fmt.Println("upload master wallet")
	return nil
}

func (c *UploadCommand) uploadBTCWallet() error {
	fmt.Println("upload BTC wallet")
	return nil
}

func (c *UploadCommand) ExecCommand(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return ErrorFromString(fmt.Sprintf("%s: no subcommand passed", uploadCommand))
	}
	c.args = &Args{args}

	return ErrorFromString(fmt.Sprintf("%s: invalid subcommand passed", uploadCommand))
}
