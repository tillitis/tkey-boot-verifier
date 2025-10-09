// SPDX-FileCopyrightText: 2025 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

package main

import (
	"crypto/ed25519"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/tillitis/tkeyclient"
	"golang.org/x/crypto/blake2s"
)

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

func startVerifier(tk *tkeyclient.TillitisKey, path string, appBin []byte, digest [blake2s.Size]byte, sig [ed25519.SignatureSize]byte) error {
	verBin, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	var secret []byte

	err = tk.LoadApp(verBin, secret)
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

func main() {
	app1Path := flag.String("update-app1", "", "Path to app to install in flash slot 1")
	flag.Parse()

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

	if *app1Path != "" {
		appBin1, err := os.ReadFile(*app1Path)
		if err != nil {
			fmt.Printf("couldn't read file: %v\n", err)
			os.Exit(1)
		}

		app1Digest := blake2s.Sum256(appBin1)
		app1Sig := [ed25519.SignatureSize]byte(
			ed25519.Sign(privateKey, app1Digest[:]))

		if err := updateApp(tk, appBin1, app1Digest, app1Sig); err != nil {
			fmt.Printf("couldn't update app slot 1: %v\n", err)
			os.Exit(1)
		}

	} else {
		// Start verifier, then another app
		appBin, err := os.ReadFile("signer.bin")
		if err != nil {
			fmt.Printf("couldn't read file: %v\n", err)
			os.Exit(1)
		}

		appDigest := blake2s.Sum256(appBin)
		appSig := [ed25519.SignatureSize]byte(
			ed25519.Sign(privateKey, appDigest[:]))

		if err := startVerifier(tk, "verifier/app.bin", appBin, appDigest, appSig); err != nil {
			fmt.Printf("couldn't load and start verifier: %v\n", err)
			os.Exit(1)
		}
	}
}
