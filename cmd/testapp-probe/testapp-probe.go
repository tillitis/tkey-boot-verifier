// SPDX-FileCopyrightText: 2025 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"

	"github.com/tillitis/tkeyclient"
)

func usage() {
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Probe test app\n\n")
	flag.PrintDefaults()
}

func main() {
	var err error

	cmd := flag.String("cmd", "", "Command. One of: reset")
	port := flag.String("port", "", "TKey serial port")
	fwResType := flag.Int("fw-reset-type", 0, "Firmware reset type. Integer")
	verifierResetDst := flag.Int("verifier-reset-dst", 0, "Verifier reset dst. Integer")

	flag.Usage = usage

	flag.Parse()

	if *cmd == "" {
		flag.Usage()
		os.Exit(1)
	}

	devPath := *port
	if devPath == "" {
		devPath, err = tkeyclient.DetectSerialPort(true)
		if err != nil {
			fmt.Printf("couldn't find any TKeys\n")
			os.Exit(1)
		}
	}

	tk := tkeyclient.New()

	if err = tk.Connect(devPath, tkeyclient.WithSpeed(tkeyclient.SerialSpeed)); err != nil {
		fmt.Printf("Could not open %s: %v\n", devPath, err)
		os.Exit(1)
	}

	defer func() { _ = tk.Close() }()

	exit := func(code int) {
		_ = tk.Close()
		os.Exit(code)
	}

	switch *cmd {
	case "reset":
		rstType, err := fwResetTypeFromInt(*fwResType)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			exit(1)
		}

		dst, err := resetDstFromInt(*verifierResetDst)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			exit(1)
		}

		err = reset(tk, rstType, dst)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "unknown command %s\n", *cmd)
		exit(1)
	}
}
