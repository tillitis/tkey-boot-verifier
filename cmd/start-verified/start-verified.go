// SPDX-FileCopyrightText: 2025 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

package main

import (
	"encoding/hex"
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
	cmdVerify = appCmd{0x01, "cmdVerify", tkeyclient.CmdLen128}
	cmdReset  = appCmd{0x02, "cmdReset", tkeyclient.CmdLen1}
)

func main() {
	devPath, err := tkeyclient.DetectSerialPort(true)
	if err != nil {
		panic(err)
	}

	tk := tkeyclient.New()

	if err = tk.Connect(devPath, tkeyclient.WithSpeed(tkeyclient.SerialSpeed)); err != nil {
		fmt.Printf("Could not open %s: %v\n", devPath, err)
		os.Exit(1)
	}

	appBin1, err := os.ReadFile("verifier/app.bin")
	if err != nil {
		fmt.Printf("Failed to read file: %v\n", err)
		os.Exit(1)
	}

	var secret []byte

	err = tk.LoadApp(appBin1, secret)
	if err != nil {
		fmt.Printf("LoadAppFromFile failed: %v\n", err)
		os.Exit(1)
	}

	// Send:
	// - framing header 1 byte
	// - 0x01 (verify) 1 byte
	// - digest 32 bytes
	// - signature 64 bytes

	tx, err := tkeyclient.NewFrameBuf(cmdVerify, 0x01)
	if err != nil {
		panic(err)
	}

	// tx[0] framing header
	// tx[1] app header (0x01)

	// 	   Digest           : 953fc88fc7612006046322c6a199b959
	//                    d3b4b2eadf711f71b2f8100bd8789ec2

	// Signature        : 079f4900f093e9aced9464628eb79545
	//                    85b027215b0b7fbf1dd77f3bae431e76
	//                    01adb1e54dea855ad6b2f8732838e6c4
	//                    2f4394814bd66cb4828527f92b2abc0b

	digest, err := hex.DecodeString("953fc88fc7612006046322c6a199b959d3b4b2eadf711f71b2f8100bd8789ec2")
	if err != nil {
		panic(err)
	}

	sig, err := hex.DecodeString("079f4900f093e9aced9464628eb7954585b027215b0b7fbf1dd77f3bae431e7601adb1e54dea855ad6b2f8732838e6c42f4394814bd66cb4828527f92b2abc0b")
	if err != nil {
		panic(err)
	}

	copy(tx[2:], digest)
	copy(tx[34:], sig)

	tkeyclient.Dump("verify tx", tx)

	if err = tk.Write(tx); err != nil {
		fmt.Fprintf(os.Stderr, "Write: %v", err)
		os.Exit(1)
	}
}
