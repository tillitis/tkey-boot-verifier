// SPDX-FileCopyrightText: 2025 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

package main

import (
	"encoding/hex"
	"flag"
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

func writeChunk(tk *tkeyclient.TillitisKey, chunk []byte) {
	tx, err := tkeyclient.NewFrameBuf(cmdUpdateAppChunk, 0x01)
	if err != nil {
		panic(err)
	}

	copy(tx[2:], chunk)

	if err = tk.Write(tx); err != nil {
		fmt.Fprintf(os.Stderr, "Write: %v", err)
		os.Exit(1)
	}
}
func main() {
	update_app1 := flag.String("update-app1", "", "Install app in flash slot 1")
	flag.Parse()

	devPath, err := tkeyclient.DetectSerialPort(true)
	if err != nil {
		panic(err)
	}

	tk := tkeyclient.New()

	if err = tk.Connect(devPath, tkeyclient.WithSpeed(tkeyclient.SerialSpeed)); err != nil {
		fmt.Printf("Could not open %s: %v\n", devPath, err)
		os.Exit(1)
	}

	if *update_app1 != "" {
		// Init
		tx, err := tkeyclient.NewFrameBuf(cmdUpdateAppInit, 0x01)
		if err != nil {
			panic(err)
		}

		digest, err := hex.DecodeString("953fc88fc7612006046322c6a199b959d3b4b2eadf711f71b2f8100bd8789ec2")
		if err != nil {
			panic(err)
		}

		sig, err := hex.DecodeString("079f4900f093e9aced9464628eb7954585b027215b0b7fbf1dd77f3bae431e7601adb1e54dea855ad6b2f8732838e6c42f4394814bd66cb4828527f92b2abc0b")
		if err != nil {
			panic(err)
		}

		appBin1, err := os.ReadFile(*update_app1)
		if err != nil {
			fmt.Printf("Failed to read file: %v\n", err)
			os.Exit(1)
		}

		size := len(appBin1)
		tx[2] = byte(size)
		tx[3] = byte(size >> 8)
		tx[4] = byte(size >> 16)
		tx[5] = byte(size >> 24)
		copy(tx[6:], digest)
		copy(tx[38:], sig)

		tkeyclient.Dump("update app1 tx", tx)

		if err = tk.Write(tx); err != nil {
			fmt.Fprintf(os.Stderr, "Write: %v", err)
			os.Exit(1)
		}

		// For each 127 byte
		//   Upload chunk
		var buf []byte
		for _, b := range appBin1 {
			buf = append(buf, b)
			if len(buf) == 127 {
				tkeyclient.Dump("update app1 chunk tx", tx)
				writeChunk(tk, buf)
				buf = []byte{}
			}
		}
		if len(buf) != 0 {
			tkeyclient.Dump("update app1 chunk tx", tx)
			writeChunk(tk, buf)
		}
	} else {
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
}
