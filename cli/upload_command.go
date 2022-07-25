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
	c.args = &Args{args}
	if len(args) == 0 {
		return ErrorFromString(fmt.Sprintf("file not found"))
	}

	if len(args) < 2 {
		return ErrorFromString(fmt.Sprintf("server ip"))
	}

	if len(args) < 3 {
		return ErrorFromString(fmt.Sprintf("server port"))
	}

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

	conn, err := net.Dial("udp", args[1]+":"+args[2])
	if err != nil {
		fmt.Println(err)
	}

	if conn == nil {
		return ErrorFromString(fmt.Sprintf("nil connection"))
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
