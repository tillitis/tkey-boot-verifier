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

	"github.com/tillitis/tkeyclient"
	"golang.org/x/crypto/blake2s"
)

// nolint:typecheck // Avoid lint error when the embedding file is missing.
//
//go:embed verifier.bin
var verifierBinary []byte

func updateApp(tk *tkeyclient.TillitisKey, appBin1 []byte, digest [blake2s.Size]byte, sig [ed25519.SignatureSize]byte) error {
	if err := updateAppInit(tk, len(appBin1), digest, sig); err != nil {
		return err
	}

	// For each 127 byte
	//   Upload chunk
	var buf []byte
	for _, b := range appBin1 {
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

func startVerifier(tk *tkeyclient.TillitisKey, appBin []byte, digest [blake2s.Size]byte, sig [ed25519.SignatureSize]byte) error {
	var secret []byte

	err := tk.LoadApp(verifierBinary, secret)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if verify(tk, digest, sig) != nil {
		return err
	}

	// Wait for TKey to reset
	time.Sleep(500 * time.Millisecond)

	devPath, err := tkeyclient.DetectSerialPort(true)
	if err != nil {
		fmt.Printf("couldn't find any TKeys\n")
		os.Exit(1)
	}

	if err = tk.Connect(devPath, tkeyclient.WithSpeed(tkeyclient.SerialSpeed)); err != nil {
		fmt.Printf("Could not open %s: %v\n", devPath, err)
		os.Exit(1)
	}

	// load appBin (using USS?)
	err = tk.LoadApp(appBin, []byte{})
	if err != nil {
		fmt.Printf("%v", err)
	}

	return nil
}

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "%s -cmd boot -app path\n%s -cmd install -app path\n", os.Args[0], os.Args[0])
	flag.PrintDefaults()
}

func main() {
	cmd := flag.String("cmd", "", "Command")
	appPath := flag.String("app", "", "Path to app")
	flag.Usage = usage

	flag.Parse()

	if *appPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Both commands need
	appBin, err := os.ReadFile(*appPath)
	if err != nil {
		fmt.Printf("couldn't read file: %v\n", err)
		os.Exit(1)
	}

	seed := []byte{
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
	}

	privateKey := ed25519.NewKeyFromSeed(seed)

	tkeyclient.SilenceLogging()

	devPath, err := tkeyclient.DetectSerialPort(true)
	if err != nil {
		fmt.Printf("couldn't find any TKeys\n")
		os.Exit(1)
	}

	tk := tkeyclient.New()
	if err = tk.Connect(devPath, tkeyclient.WithSpeed(tkeyclient.SerialSpeed)); err != nil {
		fmt.Printf("Could not open %s: %v\n", devPath, err)
		os.Exit(1)
	}

	switch *cmd {
	case "install":
		app1Digest := blake2s.Sum256(appBin)
		app1Sig := [ed25519.SignatureSize]byte(
			ed25519.Sign(privateKey, app1Digest[:]))

		if err := updateApp(tk, appBin, app1Digest, app1Sig); err != nil {
			fmt.Printf("couldn't update app slot 1: %v\n", err)
			os.Exit(1)
		}

	case "boot":
		// Start verifier, then another app
		appDigest := blake2s.Sum256(appBin)
		appSig := [ed25519.SignatureSize]byte(
			ed25519.Sign(privateKey, appDigest[:]))

		if err := startVerifier(tk, appBin, appDigest, appSig); err != nil {
			fmt.Printf("couldn't load and start verifier: %v\n", err)
			os.Exit(1)
		}
	default:
		flag.Usage()
		os.Exit(1)
	}
}
