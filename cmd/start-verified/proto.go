package main

import (
	"fmt"
	"os"

	"github.com/tillitis/tkeyclient"
)

type appCmd struct {
	code   byte
	name   string
	cmdLen tkeyclient.CmdLen
}

func (c appCmd) Code() byte {
	return c.code
}

func (c appCmd) CmdLen() tkeyclient.CmdLen {
	return c.cmdLen
}

func (c appCmd) Endpoint() tkeyclient.Endpoint {
	return tkeyclient.DestApp
}

func (c appCmd) String() string {
	return c.name
}

var (
	cmdVerify         = appCmd{0x01, "cmdVerify", tkeyclient.CmdLen128}
	cmdReset          = appCmd{0x02, "cmdReset", tkeyclient.CmdLen1}
	cmdUpdateAppInit  = appCmd{0x03, "cmdUpdateAppInit", tkeyclient.CmdLen128}
	cmdUpdateAppChunk = appCmd{0x04, "cmdUpdateAppChunk", tkeyclient.CmdLen128}
)

func reset(tk *tkeyclient.TillitisKey) {
	tx, err := tkeyclient.NewFrameBuf(cmdReset, 0x01)
	if err != nil {
		panic(err)
	}

	if err = tk.Write(tx); err != nil {
		fmt.Fprintf(os.Stderr, "Write: %v", err)
		os.Exit(1)
	}
}

func updateAppInit(tk *tkeyclient.TillitisKey, size int, digest []byte, sig []byte) error {
	tx, err := tkeyclient.NewFrameBuf(cmdUpdateAppInit, 0x01)
	if err != nil {
		return err
	}

	tx[2] = byte(size)
	tx[3] = byte(size >> 8)
	tx[4] = byte(size >> 16)
	tx[5] = byte(size >> 24)
	copy(tx[6:], digest)
	copy(tx[38:], sig)

	tkeyclient.Dump("update app1 tx", tx)

	if err = tk.Write(tx); err != nil {
		return err
	}

	return nil
}

func writeChunk(tk *tkeyclient.TillitisKey, chunk []byte) error {
	tx, err := tkeyclient.NewFrameBuf(cmdUpdateAppChunk, 0x01)
	if err != nil {
		return err
	}

	copy(tx[2:], chunk)

	tkeyclient.Dump("update app1 chunk tx", tx)

	if err = tk.Write(tx); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

// verify sends
// - framing header 1 byte
// - 0x01 (verify) 1 byte
// - digest 32 bytes
// - signature 64 bytes
func verify(tk *tkeyclient.TillitisKey, digest []byte, sig []byte) error {
	tx, err := tkeyclient.NewFrameBuf(cmdVerify, 0x01)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	copy(tx[2:], digest)
	copy(tx[34:], sig)

	tkeyclient.Dump("verify tx", tx)

	if err = tk.Write(tx); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}
