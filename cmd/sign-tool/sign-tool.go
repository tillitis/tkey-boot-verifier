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
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Write pubkey in binary form from Signify format to FILE.\n")
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), "%s -P FILE -p key.pub\n\n", os.Args[0])

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
	pubkeyPath := flag.String("p", "", "Public key file to read from or write to")
	binPubkeyPath := flag.String("P", "", "File to write binary pubkey to")
	seedPath := flag.String("s", "", "File containing private key seed in hex")
	flag.Usage = usage

	flag.Parse()

	noFileArgs := *messagePath == "" && *pubkeyPath == "" && *binPubkeyPath == ""
	tooManyFileArgs := *messagePath != "" && *pubkeyPath != ""
	if noFileArgs || tooManyFileArgs {
		flag.Usage()
		os.Exit(1)
	}

	if *messagePath != "" {
		privateKey, err := readPrivate(*seedPath)
		if err != nil {
			fmt.Printf("reading private key: %v", err)
		}

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
	} else if *binPubkeyPath != "" {
		// Write only the public key part as a binary file

		if *seedPath == "" {
			// Export binary form of pubkey from Signify pubkey
			pub, err := sigfile.ReadKey(*pubkeyPath)
			if err != nil {
				fmt.Printf("Couldn't read pubkey: %v\n", err)
				os.Exit(1)
			}

			err = sigfile.WriteBinary(*binPubkeyPath, pub.Key, true)
			if err != nil {
				fmt.Printf("Couldn't store pubkey: %v\n", err)
				os.Exit(1)
			}
		} else {
			// Export binary form derived from private key
			privateKey, err := readPrivate(*seedPath)
			if err != nil {
				fmt.Printf("reading private key: %v\n", err)
				os.Exit(1)
			}

			err = sigfile.WriteBinary(*binPubkeyPath, privateKey.Public().(ed25519.PublicKey), true)
			if err != nil {
				fmt.Printf("Couldn't store pubkey: %v\n", err)
				os.Exit(1)
			}
		}
	} else if *pubkeyPath != "" {
		privateKey, err := readPrivate(*seedPath)
		if err != nil {
			fmt.Printf("reading private key: %v\n", err)
			os.Exit(1)
		}

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
	}
}

func readPrivate(seedPath string) (ed25519.PrivateKey, error) {
	seedHex, err := os.ReadFile(seedPath)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	if len(seedHex) < 64 {
		return nil, fmt.Errorf("expected hex seed length: 64, got %d", len(seedHex))
	}

	var seed [32]byte
	seedLen, err := hex.Decode(seed[:], seedHex[:64])
	if err != nil {
		return nil, fmt.Errorf("invalid seed: %s", seed)
	}
	if seedLen != 32 {
		return nil, fmt.Errorf("expected seed length: 32, got %d", seedLen)
	}

	return ed25519.NewKeyFromSeed(seed[:]), nil
}
