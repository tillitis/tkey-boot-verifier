// SPDX-FileCopyrightText: 2025 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

package main

import (
	"bytes"
	"crypto/ed25519"
	_ "embed"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"os"

	"github.com/tillitis/tkey-sign-cli/signify"
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

		pub, err := signify.NewPubKey(privateKey.Public().(ed25519.PublicKey))
		if err != nil {
			fmt.Printf("Couldn't create pubkey: %v\n", err)
			os.Exit(1)
		}

		sig, err := signify.NewSignature(signify.B2sEd, pub.KeyNum, rawSig[:])
		if err != nil {
			fmt.Printf("Couldn't create signature: %v\n", err)
			os.Exit(1)
		}

		path := *messagePath + ".sig"
		if *sigPath != "" {
			path = *sigPath
		}

		err = sig.ToFile(path, "", true)
		if err != nil {
			fmt.Printf("Couldn't store signature: %v", err)
			os.Exit(1)
		}
	} else if *binPubkeyPath != "" {
		// Write only the public key part as a binary file

		if *seedPath == "" {
			// Export binary form of pubkey from Signify pubkey
			pub := signify.PubKey{}
			err := pub.FromFile(*pubkeyPath)
			if err != nil {
				fmt.Printf("Couldn't read pubkey: %v\n", err)
				os.Exit(1)
			}

			err = WriteBinary(*binPubkeyPath, pub.Key, true)
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

			err = WriteBinary(*binPubkeyPath, privateKey.Public().(ed25519.PublicKey), true)
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

		pub, err := signify.NewPubKey(privateKey.Public().(ed25519.PublicKey))
		if err != nil {
			fmt.Printf("Couldn't create pubkey: %v\n", err)
			os.Exit(1)
		}

		err = pub.ToFile(*pubkeyPath, "", true)
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

func WriteBinary(filename string, data any, overwrite bool) error {
	var buf bytes.Buffer

	err := binary.Write(&buf, binary.BigEndian, data)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	var f *os.File

	f, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o666)
	if err != nil {
		if os.IsExist(err) && overwrite {
			f, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o666)
			if err != nil {
				return fmt.Errorf("%w", err)
			}
		} else {
			return fmt.Errorf("%w", err)
		}
	}
	defer func() { _ = f.Close() }()

	_, err = f.Write([]byte(buf.Bytes()))
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}
