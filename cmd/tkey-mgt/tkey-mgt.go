// SPDX-FileCopyrightText: 2025 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

package main

import (
	"crypto/ed25519"
	_ "embed"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/tillitis/tkey-sign-cli/signify"
	"github.com/tillitis/tkeyclient"
	"golang.org/x/crypto/blake2s"
)

// nolint:typecheck // Avoid lint error when the embedding file is missing.
//
//go:embed verifier.bin
var verifierBinary []byte

var expectClose = true

func verifyAppSignature(tk *tkeyclient.TillitisKey, bin []byte, sig [ed25519.SignatureSize]byte) error {
	pubkey, err := getPubkey(tk)
	if err != nil {
		return err
	}

	digest := blake2s.Sum256(bin)
	if !ed25519.Verify(pubkey[:], digest[:], sig[:]) {
		return fmt.Errorf("app signature invalid")
	}

	return nil
}

func updateApp1(tk *tkeyclient.TillitisKey, bin []byte, sig [ed25519.SignatureSize]byte) error {
	err := reset(tk, fwResetTypeStartFlash0, verifierResetDstCmdMode)
	if err != nil {
		return err
	}

	if expectClose {
		waitUntilPortClosed(tk)
		reconnect(tk)
	} else {
		time.Sleep(1000 * time.Millisecond)
	}

	err = verifyAppSignature(tk, bin, sig)
	if err != nil {
		return err
	}

	digest := blake2s.Sum256(bin)

	if err := updateAppInit(tk, len(bin), digest, sig); err != nil {
		return err
	}

	// For each 127 byte
	//   Upload chunk
	var buf []byte
	for _, b := range bin {
		buf = append(buf, b)
		if len(buf) == 127 {
			if err := writeChunk(tk, buf); err != nil {
				return err
			}

			buf = []byte{}
		}
	}

	if len(buf) != 0 {
		if err := writeChunk(tk, buf); err != nil {
			return err
		}
	}

	return nil
}

func startVerifier(tk *tkeyclient.TillitisKey, appBin []byte, sig [ed25519.SignatureSize]byte) error {
	var err error
	var secret []byte

	err = reset(tk, fwResetTypeStartClient, verifierResetDstCmdMode)
	if err != nil {
		return err
	}

	if expectClose {
		waitUntilPortClosed(tk)
		reconnect(tk)
	} else {
		time.Sleep(1000 * time.Millisecond)
	}

	err = tk.LoadApp(verifierBinary, secret)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	err = verifyAppSignature(tk, appBin, sig)
	if err != nil {
		return err
	}

	digest := blake2s.Sum256(appBin)

	err = verify(tk, digest, sig)
	if err != nil {
		return err
	}

	if expectClose {
		waitUntilPortClosed(tk)
		reconnect(tk)
	} else {
		time.Sleep(1000 * time.Millisecond)
	}

	err = tk.LoadApp(appBin, []byte{})
	if err != nil {
		fmt.Printf("%v", err)
	}

	return nil
}

func waitUntilPortClosed(tk *tkeyclient.TillitisKey) {
	_, _, _ = tk.ReadFrame(rspVerify, 0x01)
	_ = tk.Close()
}

func reconnect(tk *tkeyclient.TillitisKey) {
	time.Sleep(2000 * time.Millisecond)

	devPath, err := tkeyclient.DetectSerialPort(true)
	if err != nil {
		fmt.Printf("couldn't find any TKeys\n")
		os.Exit(1)
	}

	if err = tk.Connect(devPath, tkeyclient.WithSpeed(tkeyclient.SerialSpeed)); err != nil {
		fmt.Printf("Could not open %s: %v\n", devPath, err)
		os.Exit(1)
	}
}

func usage() {
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), "%s -cmd boot -app path -sig path\n%s -cmd install -app path -sig path\n", os.Args[0], os.Args[0])
	flag.PrintDefaults()
}

func main() {
	cmd := flag.String("cmd", "", "Command")
	appPath := flag.String("app", "", "Path to app")
	sigPath := flag.String("sig", "", "Path to signature")
	port := flag.String("port", "", "TKey serial port")
	noExpectClose := flag.Bool("no-expect-close", false, "Do not expect serial port to disappear when TKey resets")
	flag.Usage = usage

	flag.Parse()

	expectClose = !*noExpectClose

	if *appPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	if *sigPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Both commands need
	appBin, err := os.ReadFile(*appPath)
	if err != nil {
		fmt.Printf("couldn't read file: %v\n", err)
		os.Exit(1)
	}

	var appSig signify.Signature

	if err := appSig.FromFile(*sigPath); err != nil {
		fmt.Printf("couldn't read file: %v\n", err)
		os.Exit(1)
	}

	tkeyclient.SilenceLogging()

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
	case "install":
		if err := updateApp1(tk, appBin, appSig.Sig); err != nil {
			fmt.Printf("couldn't update app slot 1: %v\n", err)
			exit(1)
		}

	case "boot":
		if err := startVerifier(tk, appBin, appSig.Sig); err != nil {
			fmt.Printf("couldn't load and start verifier: %v\n", err)
			exit(1)
		}
	default:
		flag.Usage()
		exit(1)
	}
}
