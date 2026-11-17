// SPDX-FileCopyrightText: 2025 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

package main

import (
	"bufio"
	"bytes"
	"crypto/ed25519"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/tillitis/tkey-sign-cli/distfile"
	"github.com/tillitis/tkey-sign-cli/signify"
	"github.com/tillitis/tkeyclient"
	"golang.org/x/crypto/blake2s"
)

// nolint:typecheck // Avoid lint error when the embedding file is missing.
//
//go:embed verifier.bin
var verifierBinary []byte

func verifyAppSignature(bin []byte, distinfo distfile.DistInfo) error {
	// Verify app and digest is equal
	digest := blake2s.Sum256(bin)
	if !bytes.Equal(digest[:], distinfo.Message.AppDigest[:]) {
		return errors.New("App and appDigest in cert file does not match")
	}

	message, err := distinfo.Message.SigningDigest()
	if err != nil {
		return err
	}

	if !ed25519.Verify(distinfo.PubKey.Key[:], message, distinfo.Signature.Sig[:]) {
		return fmt.Errorf("app signature invalid")
	}

	return nil
}

func reconnect(tk *tkeyclient.TillitisKey) error {
	err := tk.WaitClosed()
	if err != nil {
		return fmt.Errorf("expected port close: %w", err)
	}

	err = tk.Reconnect()
	if err != nil {
		return fmt.Errorf("couldn't reconnect: %w", err)
	}

	return nil
}

func userConfirm() bool {
	fmt.Printf("\nAny data stored by the app(s) will be erased and cannot be restored.")

	fmt.Printf("\nAre you sure? Press y to continue or anything else to abort.\n")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')

	// Trim whitespace to remove any remaining Windows line ending
	input = strings.TrimSpace(input)

	if input != "y" {
		return false
	}
	return true
}

func eraseAll(tk *tkeyclient.TillitisKey) error {
	err := tk.Reset(tkeyclient.RstTypeStartFlash0, tkeyclient.VerifierCmdMode)
	if err != nil {
		return err
	}

	err = reconnect(tk)
	if err != nil {
		return err
	}

	if !userConfirm() {
		return fmt.Errorf("User aborted")
	}

	fmt.Printf("Your TKey will begin to blink yellow.\n")
	fmt.Printf("Confirm the erase operation by touching the TKey touch sensor three times.\n")
	fmt.Printf("If you want to abort then wait for the process to timeout.\n")

	err = eraseAreas(tk)
	if err != nil {
		return err
	}

	fmt.Printf("\nAll data erased\n")

	return nil
}

func updateApp1(tk *tkeyclient.TillitisKey, bin []byte, distinfo distfile.DistInfo) error {
	err := tk.Reset(tkeyclient.RstTypeStartFlash0, tkeyclient.VerifierCmdMode)
	if err != nil {
		return err
	}

	err = reconnect(tk)
	if err != nil {
		return err
	}

	pubkey, err := getPubkey(tk)
	if err != nil {
		return err
	}

	if distinfo.PubKey.Key != pubkey {
		return errors.New("supplied pubkey does not match pyubkey stored on device")
	}

	err = verifyAppSignature(bin, distinfo)
	if err != nil {
		return err
	}

	fmt.Printf("Your TKey will begin to blink yellow.\n")
	fmt.Printf("Any installed app will be replaced. To confirm the installation, touch the TKey three times.\n")
	fmt.Printf("If you want to abort then wait for the process to timeout.\n")

	digest := blake2s.Sum256(bin)

	if err := updateAppInit(tk, len(bin), digest, distinfo.Signature.Sig, distinfo.Message.Name); err != nil {
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

	fmt.Printf("\nApp installed\n")

	return nil
}

func startVerifier(tk *tkeyclient.TillitisKey, appBin []byte, distinfo distfile.DistInfo) error {
	var err error
	var secret []byte

	err = verifyAppSignature(appBin, distinfo)
	if err != nil {
		return err
	}

	err = tk.Reset(tkeyclient.RstTypeStartClient, tkeyclient.VerifierCmdMode)
	if err != nil {
		return err
	}

	err = reconnect(tk)
	if err != nil {
		return err
	}

	err = tk.LoadApp(verifierBinary, secret)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	err = reconnect(tk)
	if err != nil {
		return err
	}

	err = setPubkey(tk, distinfo.PubKey.Key)
	if err != nil {
		return fmt.Errorf("couldn't set pubkey: %w", err)
	}

	digest := blake2s.Sum256(appBin)

	err = verify(tk, digest, distinfo.Signature.Sig, distinfo.Message.Name, distinfo.Message.Installable)
	if err != nil {
		return err
	}

	err = reconnect(tk)
	if err != nil {
		return err
	}

	err = tk.LoadApp(appBin, []byte{})
	if err != nil {
		fmt.Printf("%v", err)
	}

	return nil
}

func installPubkey(tk *tkeyclient.TillitisKey, pubkey [32]byte, sig [64]byte) error {
	err := tk.Reset(tkeyclient.RstTypeStartFlash0, tkeyclient.VerifierCmdMode)
	if err != nil {
		return err
	}

	err = reconnect(tk)
	if err != nil {
		return err
	}

	currentPubkey, err := getPubkey(tk)
	if err != nil {
		return err
	}

	if bytes.Equal(currentPubkey[:], pubkey[:]) {
		return errors.New("pubkey already installed")
	}

	fmt.Printf("Your TKey will begin to blink yellow.\n")
	fmt.Printf("Confirm the pubkey update by tapping the TKey touch sensor.\n")
	fmt.Printf("If you want to abort then wait for the process to timeout.\n")

	err = storePubkey(tk, pubkey, sig)
	if err != nil {
		return err
	}

	readbackPubkey, err := getPubkey(tk)
	if err != nil {
		return err
	}

	if !bytes.Equal(readbackPubkey[:], pubkey[:]) {
		return errors.New("something went wrong, pubkey not installed")
	}

	fmt.Printf("\nPubkey updated\n")

	err = tk.Reset(tkeyclient.RstTypeStartFlash0, tkeyclient.VerifierBootSlot1)
	if err != nil {
		return err
	}

	// Wait until closed to make sure the reset command was delivered before the program exits
	err = tk.WaitClosed()
	if err != nil {
		return err
	}

	return nil
}

func usage() {
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Boot and install applications on a TKey, install public keys,\nand erase storage areas.\n\n")

	_, _ = fmt.Fprintf(flag.CommandLine.Output(), "%s -cmd boot -app APPFILE -cert CERTFILE\n", os.Args[0])
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), "%s -cmd install -app APPFILE -cert CERTFILE\n", os.Args[0])
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), "%s -cmd install-pubkey -pub PUBFILE -sig SIGFILE\n", os.Args[0])
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), "%s -cmd erase-areas\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	var err error

	cmd := flag.String("cmd", "", "`command` to perform")
	appPath := flag.String("app", "", "Path to the application binary `file`")
	distinfoPath := flag.String("cert", "", "Path to distribution certificate `file`")
	sigPath := flag.String("sig", "", "Path to signature `file`")
	pubPath := flag.String("pub", "", "Path to pubkey `file`")
	port := flag.String("port", "", "TKey serial `port`")
	noExpectClose := flag.Bool("no-expect-close", false, "Do not expect serial port to disappear during resets")
	flag.Usage = usage

	flag.Parse()

	// tkeyclient.SilenceLogging()

	devPath := *port

	tkOpts := []func(*tkeyclient.TillitisKey){
		tkeyclient.WithSpeed(tkeyclient.SerialSpeed),
	}
	if *noExpectClose {
		tkOpts = append(tkOpts,
			tkeyclient.NotUSBDevice(),
			tkeyclient.NoRemoteClose())
	}

	tk := tkeyclient.New()
	if err = tk.Connect(devPath, tkOpts...); err != nil {
		fmt.Printf("Could not open %s: %v\n", devPath, err)
		os.Exit(1)
	}
	defer func() { _ = tk.Close() }()

	exit := func(code int) {
		_ = tk.Close()
		os.Exit(code)
	}

	if !*noExpectClose && !tk.CanRemoteClose {
		fmt.Printf("unsupported TKey model\n")
		os.Exit(1)
	}

	switch *cmd {
	case "erase-areas":
		if err := eraseAll(tk); err != nil {
			fmt.Printf("couldn't erase areas: %v\n", err)
			exit(1)
		}

	case "install":
		if *appPath == "" {
			flag.Usage()
			os.Exit(1)
		}

		if *distinfoPath == "" {
			flag.Usage()
			os.Exit(1)
		}

		appBin, err := os.ReadFile(*appPath)
		if err != nil {
			fmt.Printf("couldn't read file: %v\n", err)
			os.Exit(1)
		}

		distinfo := distfile.DistInfo{}
		err = distinfo.FromFile(*distinfoPath)
		if err != nil {
			fmt.Printf("couldn't read distinfo: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("distinfo: %v\n", distinfo)

		if distinfo.Signature.Alg != distfile.B2sEd {
			fmt.Printf("incompatible sig file, expected ed25519 signature over blake2s digest\n")
			os.Exit(1)
		}

		if err := updateApp1(tk, appBin, distinfo); err != nil {
			fmt.Printf("couldn't update app slot 1: %v\n", err)
			exit(1)
		}

	case "boot":
		if *appPath == "" || *distinfoPath == "" {
			flag.Usage()
			os.Exit(1)
		}

		appBin, err := os.ReadFile(*appPath)
		if err != nil {
			fmt.Printf("couldn't read file: %v\n", err)
			os.Exit(1)
		}

		distinfo := distfile.DistInfo{}
		err = distinfo.FromFile(*distinfoPath)
		if err != nil {
			fmt.Printf("couldn't read distinfo: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("distinfo: %v\n", distinfo)

		if distinfo.Signature.Alg != distfile.B2sEd {
			fmt.Printf("incompatible sig file, expected ed25519 signature over blake2s digest\n")
			os.Exit(1)
		}

		if err := startVerifier(tk, appBin, distinfo); err != nil {
			fmt.Printf("couldn't load and start verifier: %v\n", err)
			exit(1)
		}

	case "install-pubkey":
		if *pubPath == "" {
			flag.Usage()
			os.Exit(1)
		}

		appPub := signify.PubKey{}
		err := appPub.FromFile(*pubPath)
		if err != nil {
			fmt.Printf("couldn't read file: %v\n", err)
			os.Exit(1)
		}

		appSig := signify.Signature{}
		err = appSig.FromFile(*sigPath)
		if err != nil {
			fmt.Printf("couldn't read file: %v\n", err)
			os.Exit(1)
		}

		if err := installPubkey(tk, appPub.Key, appSig.Sig); err != nil {
			fmt.Printf("couldn't set pubkey: %v\n", err)
			exit(1)
		}

	default:
		flag.Usage()
		exit(1)
	}
}
