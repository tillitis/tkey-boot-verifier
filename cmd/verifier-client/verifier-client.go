// SPDX-FileCopyrightText: 2025 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"flag"
	"fmt"
	"os"

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

func startVerifier(tk *tkeyclient.TillitisKey, path string, digest []byte, sig []byte) error {
	appBin1, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	var secret []byte

	err = tk.LoadApp(appBin1, secret)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if verify(tk, digest, sig) != nil {
		return err
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
		app1Digest, err := hex.DecodeString("953fc88fc7612006046322c6a199b959d3b4b2eadf711f71b2f8100bd8789ec2")
		if err != nil {
			panic(err)
		}
		app1Sig, err := hex.DecodeString("079f4900f093e9aced9464628eb7954585b027215b0b7fbf1dd77f3bae431e7601adb1e54dea855ad6b2f8732838e6c42f4394814bd66cb4828527f92b2abc0b")
		if err != nil {
			panic(err)
		}

		if err := startVerifier(tk, "verifier/app.bin", app1Digest, app1Sig); err != nil {
			fmt.Printf("couldn't load and start verifier: %v\n", err)
			os.Exit(1)
		}
	}
}
