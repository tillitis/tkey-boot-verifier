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
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), "%s signs a BLAKE2s digest of the contents of a file or exports the public key.\n\n", os.Args[0])
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Sign message in FILE and write the result to FILE.sig (default):\n")
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), "%s -m FILE -s seckey [-o SIGFILE]\n\n", os.Args[0])

	_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Write pubkey (in binary form with -P) generated from seckey to FILE.\n")
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), "%s -p|P FILE -s seckey\n\n", os.Args[0])

	flag.PrintDefaults()
}

type signature struct {
	Alg    [2]byte
	KeyNum [8]byte
	Sig    [64]byte
}

type pubKey struct {
	Alg    [2]byte
	KeyNum [8]byte
	Key    [ed25519.PublicKeySize]byte
}

func main() {
	messagePath := flag.String("m", "", "File containing message to sign")
	sigPath := flag.String("o", "", "File to write signature to. Default: <message-file>.sig")
	pubkeyPath := flag.String("p", "", "File to write pubkey to")
	binPubkeyPath := flag.String("P", "", "File to write pubkey to")
	seedPath := flag.String("s", "", "File containing private key seed in hex")
	flag.Usage = usage

	flag.Parse()

	noFileArgs := *messagePath == "" && *pubkeyPath == "" && *binPubkeyPath == ""
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
		rawSig := [ed25519.SignatureSize]byte(
			ed25519.Sign(privateKey, digest[:]))

		sig := signature{
			Alg:    [2]byte{'E', 'b'},
			KeyNum: [8]byte{1, 7},
			Sig:    [64]byte{},
		}

		copy(sig.Sig[:], rawSig[:])

		path := *messagePath + ".sig"
		if *sigPath != "" {
			path = *sigPath
		}

		err = sigfile.WriteBase64(path, sig, "", true)
		if err != nil {
			fmt.Printf("Couldn't store signature: %v", err)
			os.Exit(1)
		}
	} else if *pubkeyPath != "" {
		pub := pubKey{
			Alg:    [2]byte{'E', 'b'},
			KeyNum: [8]byte{1, 7},
		}
		copy(pub.Key[:], privateKey.Public().(ed25519.PublicKey))

		err = sigfile.WriteBase64(*pubkeyPath, pub, "", true)
		if err != nil {
			fmt.Printf("Couldn't store pubkey: %v\n", err)
			os.Exit(1)
		}
	} else if *binPubkeyPath != "" {
		// Write only the public key part as a binary file

		err = sigfile.WriteBinary(*binPubkeyPath, privateKey.Public().(ed25519.PublicKey), true)
		if err != nil {
			fmt.Printf("Couldn't store pubkey: %v\n", err)
			os.Exit(1)
		}
	}
}
