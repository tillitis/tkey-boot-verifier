// SPDX-FileCopyrightText: 2025 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

package main

import (
	"crypto/ed25519"
	_ "embed"
	"encoding/hex"
	"flag"
	"fmt"
	"os"

	"github.com/tillitis/tkey-sign-cli/signify"
	"golang.org/x/crypto/blake2s"
)

func usage() {
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), "%s [-m|-p] FILE -s seckey\n\n", os.Args[0])
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Sign message in FILE and write the result to file.sig.\n")
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Or, write pubkey generated from seckey to FILE.\n")
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Signatures and pubkeys are produced by Ed25519-signing the Blake2s digest of message.\n\n")
	flag.PrintDefaults()
}

func main() {
	messagePath := flag.String("m", "", "File containing message to sign")
	pubkeyPath := flag.String("p", "", "File to write pubkey to")
	seedPath := flag.String("s", "", "File containing private key seed in hex")
	flag.Usage = usage

	flag.Parse()

	noFileArgs := *messagePath == "" && *pubkeyPath == ""
	tooManyFileArgs := *messagePath != "" && *pubkeyPath != ""
	if noFileArgs || tooManyFileArgs {
		flag.Usage()
		os.Exit(1)
	}

	if *seedPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	seedHex, err := os.ReadFile(*seedPath)
	if err != nil {
		fmt.Printf("couldn't read file: %v\n", err)
		os.Exit(1)
	}
	if len(seedHex) < 64 {
		fmt.Printf("Expected seed length: 64, got %d\n", len(seedHex))
		os.Exit(1)
	}

	var seed [32]byte
	seedLen, err := hex.Decode(seed[:], seedHex[:64])
	if err != nil {
		fmt.Printf("Invalid seed: %s\n", seed)
		os.Exit(1)
	}
	if seedLen != 32 {
		fmt.Printf("Expected seed length: 32, got %d\n", seedLen)
		os.Exit(1)
	}

	privateKey := ed25519.NewKeyFromSeed(seed[:])

	if *messagePath != "" {
		message, err := os.ReadFile(*messagePath)
		if err != nil {
			fmt.Printf("couldn't read file: %v\n", err)
			os.Exit(1)
		}

		digest := blake2s.Sum256(message)
		sig, err := signify.NewSignature(signify.B2sEd, ed25519.Sign(privateKey, digest[:]))
		if err != nil {
			fmt.Printf("couldn't convert signature")
			os.Exit(1)
		}

		if err := sig.ToFile(*messagePath+".sig", "", true); err != nil {
			fmt.Printf("Couldn't store signature: %v", err)
			os.Exit(1)
		}
	} else if *pubkeyPath != "" {
		pub, err := signify.NewPubKey(privateKey.Public().(ed25519.PublicKey))
		if err != nil {
			fmt.Printf("couldn't convert to signify pubkey: %v\n", err)
			os.Exit(1)
		}

		if err := pub.ToFile(*pubkeyPath, "", true); err != nil {
			fmt.Printf("Couldn't store pubkey: %v\n", err)
			os.Exit(1)
		}
	}
}
