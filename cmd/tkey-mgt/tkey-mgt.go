// SPDX-FileCopyrightText: 2025 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

package main

import (
	"bytes"
	"crypto/ed25519"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"tkey-mgt/sigfile"

	"github.com/tillitis/tkeyclient"
	"golang.org/x/crypto/blake2s"
)

// nolint:typecheck // Avoid lint error when the embedding file is missing.
//
//go:embed verifier.bin
var verifierBinary []byte

var expectClose = true

func verifyAppSignature(tk *tkeyclient.TillitisKey, pubKey [ed25519.PublicKeySize]byte, bin []byte, sig [ed25519.SignatureSize]byte) error {
	digest := blake2s.Sum256(bin)
	if !ed25519.Verify(pubKey[:], digest[:], sig[:]) {
		return fmt.Errorf("app signature invalid")
	}

	return nil
}

func eraseAll(tk *tkeyclient.TillitisKey) error {
	err := reset(tk, fwResetTypeStartFlash0, verifierResetDstCmdMode)
	if err != nil {
		return err
	}

	if expectClose {
		waitUntilPortClosed(tk)
		reconnect(tk)
	} else {
		time.Sleep(1000 * time.Millisecond)
	}

	fmt.Printf("Your TKey will begin to blink yellow.\n")
	fmt.Printf("Any data stored by any app will be erased and cannot be restored. Confirm the erase operation by touching the TKey touch sensor three times.\n")
	fmt.Printf("If you want to abort then wait for the process to timeout.\n")

	err = eraseAreas(tk)
	if err != nil {
		return err
	}

	fmt.Printf("\nAll data erased\n")

	return nil
}

func updateApp1(tk *tkeyclient.TillitisKey, bin []byte, sig [ed25519.SignatureSize]byte) error {
	err := reset(tk, fwResetTypeStartFlash0, verifierResetDstCmdMode)
	if err != nil {
		return err
	}

	if expectClose {
		waitUntilPortClosed(tk)
		reconnect(tk)
	} else {
		time.Sleep(1000 * time.Millisecond)
	}

	pubkey, err := getPubkey(tk)
	if err != nil {
		return err
	}

	err = verifyAppSignature(tk, pubkey, bin, sig)
	if err != nil {
		return err
	}

	fmt.Printf("Your TKey will begin to blink yellow.\n")
	fmt.Printf("Any installed app will be replaced. To confirm the installation, touch the TKey three times.\n")
	fmt.Printf("If you want to abort then wait for the process to timeout.\n")

	digest := blake2s.Sum256(bin)

	if err := updateAppInit(tk, len(bin), digest, sig); err != nil {
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

func startVerifier(tk *tkeyclient.TillitisKey, pubKey [ed25519.PublicKeySize]byte, appBin []byte, sig [ed25519.SignatureSize]byte) error {
	var err error
	var secret []byte

	err = verifyAppSignature(tk, pubKey, appBin, sig)
	if err != nil {
		return err
	}

	err = reset(tk, fwResetTypeStartClient, verifierResetDstCmdMode)
	if err != nil {
		return err
	}

	if expectClose {
		waitUntilPortClosed(tk)
		reconnect(tk)
	} else {
		time.Sleep(1000 * time.Millisecond)
	}

	err = tk.LoadApp(verifierBinary, secret)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if err := setPubkey(tk, pubKey); err != nil {
		return err
	}

	digest := blake2s.Sum256(appBin)

	err = verify(tk, digest, sig)
	if err != nil {
		return err
	}

	if expectClose {
		waitUntilPortClosed(tk)
		reconnect(tk)
	} else {
		time.Sleep(1000 * time.Millisecond)
	}

	err = tk.LoadApp(appBin, []byte{})
	if err != nil {
		fmt.Printf("%v", err)
	}

	return nil
}

func installPubkey(tk *tkeyclient.TillitisKey, pubkey [32]byte) error {
	err := reset(tk, fwResetTypeStartFlash0, verifierResetDstCmdMode)
	if err != nil {
		return err
	}

	if expectClose {
		waitUntilPortClosed(tk)
		reconnect(tk)
	} else {
		time.Sleep(1000 * time.Millisecond)
	}

	currentPubkey, err := getPubkey(tk)
	if err != nil {
		return err
	}

	if bytes.Equal(currentPubkey[:], pubkey[:]) {
		return errors.New("pubkey already installed")
	}

	fmt.Printf("Your TKey will begin to blink yellow.\n")
	fmt.Printf("Confirm the pubkey update by touching the TKey touch sensor three times.\n")
	fmt.Printf("If you want to abort then wait for the process to timeout.\n")

	err = storePubkey(tk, pubkey)
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

	err = reset(tk, fwResetTypeStartDefault, verifierResetDstApp1)
	if err != nil {
		return err
	}

	return nil
}

func waitUntilPortClosed(tk *tkeyclient.TillitisKey) {
	_, _, _ = tk.ReadFrame(rspVerify, 0x01)
	_ = tk.Close()
}

func reconnect(tk *tkeyclient.TillitisKey) {
	time.Sleep(2000 * time.Millisecond)

	devPath, err := tkeyclient.DetectSerialPort(true)
	if err != nil {
		fmt.Printf("couldn't find any TKeys\n")
		os.Exit(1)
	}

	if err = tk.Connect(devPath, tkeyclient.WithSpeed(tkeyclient.SerialSpeed)); err != nil {
		fmt.Printf("Could not open %s: %v\n", devPath, err)
		os.Exit(1)
	}
}

func usage() {
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), "%s -cmd boot -app path -sig path -pub path-to-pubkey\n", os.Args[0])
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), "%s -cmd install -app path -sig path\n", os.Args[0])
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), "%s -cmd install-pubkey -pub path\n", os.Args[0])
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), "%s -cmd erase-areas\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	var err error

	cmd := flag.String("cmd", "", "Command")
	appPath := flag.String("app", "", "Path to app")
	sigPath := flag.String("sig", "", "Path to signature")
	pubPath := flag.String("pub", "", "Path to pubkey")
	port := flag.String("port", "", "TKey serial port")
	noExpectClose := flag.Bool("no-expect-close", false, "Do not expect serial port to disappear when TKey resets")
	flag.Usage = usage

	flag.Parse()

	expectClose = !*noExpectClose

	tkeyclient.SilenceLogging()

	devPath := *port
	if devPath == "" {
		devPath, err = tkeyclient.DetectSerialPort(true)
		if err != nil {
			fmt.Printf("couldn't find any TKeys\n")
			os.Exit(1)
		}
	}

	tk := tkeyclient.New()
	if err = tk.Connect(devPath, tkeyclient.WithSpeed(tkeyclient.SerialSpeed)); err != nil {
		fmt.Printf("Could not open %s: %v\n", devPath, err)
		os.Exit(1)
	}
	defer func() { _ = tk.Close() }()

	exit := func(code int) {
		_ = tk.Close()
		os.Exit(code)
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

		if *sigPath == "" {
			flag.Usage()
			os.Exit(1)
		}

		appBin, err := os.ReadFile(*appPath)
		if err != nil {
			fmt.Printf("couldn't read file: %v\n", err)
			os.Exit(1)
		}

		appSig, err := sigfile.ReadSig(*sigPath)
		if err != nil {
			fmt.Printf("couldn't read file: %v\n", err)
			os.Exit(1)
		}
		if appSig.Alg != [2]byte{'E', 'b'} {
			fmt.Printf("incompatible sig file, expected ed25519 signature over blake2s digest\n")
			os.Exit(1)
		}

		if err := updateApp1(tk, appBin, appSig.Sig); err != nil {
			fmt.Printf("couldn't update app slot 1: %v\n", err)
			exit(1)
		}

	case "boot":
		if *appPath == "" || *sigPath == "" || *pubPath == "" {
			flag.Usage()
			os.Exit(1)
		}

		appPub, err := sigfile.ReadKey(*pubPath)
		if err != nil {
			fmt.Printf("couldn't read file: %v\n", err)
			os.Exit(1)
		}

		appBin, err := os.ReadFile(*appPath)
		if err != nil {
			fmt.Printf("couldn't read file: %v\n", err)
			os.Exit(1)
		}

		appSig, err := sigfile.ReadSig(*sigPath)
		if err != nil {
			fmt.Printf("couldn't read file: %v\n", err)
			os.Exit(1)
		}
		if appSig.Alg != [2]byte{'E', 'b'} {
			fmt.Printf("incompatible sig file, expected ed25519 signature over blake2s digest\n")
			os.Exit(1)
		}

		if err := startVerifier(tk, appPub.Key, appBin, appSig.Sig); err != nil {
			fmt.Printf("couldn't load and start verifier: %v\n", err)
			exit(1)
		}

	case "install-pubkey":
		if *pubPath == "" {
			flag.Usage()
			os.Exit(1)
		}

		appPub, err := sigfile.ReadKey(*pubPath)
		if err != nil {
			fmt.Printf("couldn't read file: %v\n", err)
			os.Exit(1)
		}

		if err := installPubkey(tk, appPub.Key); err != nil {
			fmt.Printf("couldn't set pubkey: %v\n", err)
			exit(1)
		}

	default:
		flag.Usage()
		exit(1)
	}
}
