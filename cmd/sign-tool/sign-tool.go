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

	"github.com/tillitis/tkey-sign-cli/distfile"
	"github.com/tillitis/tkey-sign-cli/signify"
	"golang.org/x/crypto/blake2s"
)

func usage() {
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), `Sign files, create distribution message or certificate, and export Ed25519 public keys.

Sign a BLAKE2s digest of file and write a detached signature (default output: FILE.sig)
  %[1]s -m FILE -s SEEDFILE [-o OUTFILE]

Export a distribution message as JSON (default output: APP_msg.json)
  %[1]s -d -m APP [-n NAME] [-i] [-o OUTFILE]

Create a signed distribution certificate as JSON (default output: APP_cert.json)
  %[1]s -d -m APP -s SEEDFILE [-n NAME] [-i] [-o OUTFILE]

Derive a public key from a seed and write it in Signify format
  %[1]s -p PUBFILE -s SEEDFILE

Derive a public key from a seed and write it as 32 raw bytes
  %[1]s -P OUTFILE -s SEEDFILE

Extract the 32 raw public key bytes from a Signify public key file
  %[1]s -P OUTFILE -p PUBFILE

`, os.Args[0])

	flag.PrintDefaults()
}

func main() {
	messagePath := flag.String("m", "", "Input `file`. With -d, this is the application binary.")
	sigPath := flag.String("o", "", "Output path for the signature or JSON `file`.")
	pubkeyPath := flag.String("p", "", "Signify public key `file` to read or write.")
	binPubkeyPath := flag.String("P", "", "Output `file` for the 32 raw public key bytes.")
	seedPath := flag.String("s", "", "`File` containing a hex-encoded 32-byte Ed25519 seed.")
	name := flag.String("n", "", "`name` to include in the distribution message, ASCII. Only with -d.")
	distFlag := flag.Bool("d", false, "Create distribution message as JSON. With -s, create a signed distribution certificate.")
	installable := flag.Bool("i", false, "Mark the application as installable in the distribution message. Only with -d.")
	flag.Usage = usage

	flag.Parse()

	noFileArgs := *messagePath == "" && *pubkeyPath == "" && *binPubkeyPath == ""
	tooManyFileArgs := *messagePath != "" && *pubkeyPath != ""
	if noFileArgs || tooManyFileArgs {
		flag.Usage()
		os.Exit(1)
	}

	if *distFlag {

		app, err := os.ReadFile(*messagePath)
		if err != nil {
			fmt.Printf("couldn't read file: %v\n", err)
			os.Exit(1)
		}

		message := distfile.Message{}
		message.AppDigest = blake2s.Sum256(app)

		if *name != "" {
			message.Name = *name
		}

		if *installable {
			message.Installable = true
		}

		// No private key, just export message
		if *seedPath == "" {
			err := message.Validate()
			if err != nil {
				fmt.Printf("message invalid: %v\n", err)
				os.Exit(1)

			}
			b, err := message.ToBuffer()
			if err != nil {
				fmt.Printf("couldn't serialize: %v\n", err)
				os.Exit(1)
			}

			path := *messagePath + "_msg.json"

			if *sigPath != "" {
				path = *sigPath
			}
			err = writeFile(path, b, true)
			if err != nil {
				fmt.Printf("couldn't write file: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		}
		b, err := message.SigningDigest()
		if err != nil {
			fmt.Printf("creating signing digest failed: %v\n", err)
			os.Exit(1)
		}

		privateKey, err := readPrivate(*seedPath)
		if err != nil {
			fmt.Printf("reading private key: %v\n", err)
			os.Exit(1)
		}

		rawSig := [ed25519.SignatureSize]byte(
			ed25519.Sign(privateKey, b[:]))

		di := distfile.DistInfo{}
		di.Version = 1
		di.Message = message
		di.Signature.Alg = distfile.B2sEd
		copy(di.Signature.Sig[:], rawSig[:])
		di.PubKey.Alg = distfile.Ed
		copy(di.PubKey.Key[:], privateKey.Public().(ed25519.PublicKey))

		err = di.Validate()
		if err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}

		path := *messagePath + "_cert.json"
		if *sigPath != "" {
			path = *sigPath
		}

		b, err = di.ToBuffer()
		if err != nil {
			fmt.Printf("couldn't serialize: %v\n", err)
			os.Exit(1)
		}

		err = writeFile(path, b, true)
		if err != nil {
			fmt.Printf("couldn't write file: %v\n", err)
			os.Exit(1)
		}

		os.Exit(0)
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

// writeFile writes buf into file filename. If overwrite is true it
// overwrites any existing file, otherwise it returns an error.
func writeFile(filename string, buf []byte, overwrite bool) error {
	var f *os.File

	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o666)
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

	defer f.Close()

	if _, err := f.Write(buf); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}
