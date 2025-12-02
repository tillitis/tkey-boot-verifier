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

	"tkey-mgt/sigfile"

	"golang.org/x/crypto/blake2s"
)

func usage() {
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), "%s -m FILE -s seckey\n\n", os.Args[0])
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Sign message in FILE and write the result to file.sig.\n")
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Signatures are produced by Ed25519-signing the Blake2s digest of message.\n\n")
	flag.PrintDefaults()
}

type signature struct {
	Alg    [2]byte
	KeyNum [8]byte
	Sig    [64]byte
}

func main() {
	messagePath := flag.String("m", "", "File containing message to sign")
	seedPath := flag.String("s", "", "File containing private key seed in hex")
	flag.Usage = usage

	flag.Parse()

	if *messagePath == "" {
		flag.Usage()
		os.Exit(1)
	}

	if *seedPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	message, err := os.ReadFile(*messagePath)
	if err != nil {
		fmt.Printf("couldn't read file: %v\n", err)
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

	digest := blake2s.Sum256(message)
	rawSig := [ed25519.SignatureSize]byte(
		ed25519.Sign(privateKey, digest[:]))

	sig := signature{
		Alg:    [2]byte{'E', 'b'},
		KeyNum: [8]byte{1, 7},
		Sig:    [64]byte{},
	}

	copy(sig.Sig[:], rawSig[:])

	err = sigfile.WriteBase64(*messagePath+".sig", sig, "", true)
	if err != nil {
		fmt.Printf("Couldn't store signature: %v", err)
		os.Exit(1)
	}
}
