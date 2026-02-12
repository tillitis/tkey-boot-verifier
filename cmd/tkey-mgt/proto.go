// SPDX-FileCopyrightText: 2025 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

package main

import (
	"crypto/ed25519"
	"fmt"

	"github.com/tillitis/tkeyclient"
	"golang.org/x/crypto/blake2s"
)

type appCmd struct {
	code   byte
	name   string
	cmdLen tkeyclient.CmdLen
}

func (c appCmd) Code() byte {
	return c.code
}

func (c appCmd) CmdLen() tkeyclient.CmdLen {
	return c.cmdLen
}

func (c appCmd) Endpoint() tkeyclient.Endpoint {
	return tkeyclient.DestApp
}

func (c appCmd) String() string {
	return c.name
}

var (
	cmdVerify         = appCmd{0x01, "cmdVerify", tkeyclient.CmdLen128}
	cmdUpdateAppInit  = appCmd{0x03, "cmdUpdateAppInit", tkeyclient.CmdLen128}
	cmdUpdateAppChunk = appCmd{0x04, "cmdUpdateAppChunk", tkeyclient.CmdLen128}
	cmdGetPubkey      = appCmd{0x05, "cmdGetPubkey", tkeyclient.CmdLen1}
	cmdStorePubkey    = appCmd{0x06, "cmdStorePubkey", tkeyclient.CmdLen128}
	cmdSetPubkey      = appCmd{0x07, "cmdSetPubkey", tkeyclient.CmdLen128}
	cmdReset          = appCmd{0xfe, "cmdReset", tkeyclient.CmdLen4}

	rspVerify         = appCmd{0x01, "rspVerify", tkeyclient.CmdLen4}
	rspUpdateAppInit  = appCmd{0x03, "rspUpdateAppInit", tkeyclient.CmdLen4}
	rspUpdateAppChunk = appCmd{0x04, "rspUpdateAppChunk", tkeyclient.CmdLen4}
	rspGetPubkey      = appCmd{0x05, "rspGetPubkey", tkeyclient.CmdLen128}
	rspStorePubkey    = appCmd{0x06, "rspStorePubkey", tkeyclient.CmdLen4}
	rspSetPubkey      = appCmd{0x07, "rspSetPubkey", tkeyclient.CmdLen4}
)

const devicePresenceTimeoutS = 20
const devicePresenceRepeatDelayS = 1
const devicePresenceRepeats = 3
const storePubkeyTimeout = (devicePresenceTimeoutS + devicePresenceRepeatDelayS) * devicePresenceRepeats

type fwResetType uint8

const (
	fwResetTypeStartDefault   fwResetType = 0
	fwResetTypeStartFlash0    fwResetType = 1
	fwResetTypeStartFlash1    fwResetType = 2
	fwResetTypeStartFlash0Ver fwResetType = 3
	fwResetTypeStartFlash1Ver fwResetType = 4
	fwResetTypeStartClient    fwResetType = 5
	fwResetTypeStartClientVer fwResetType = 6
)

type resetDst uint8

const (
	verifierResetDstApp1    = 0
	verifierResetDstCmdMode = 1
)

func reset(tk *tkeyclient.TillitisKey, fwType fwResetType, verifierDst resetDst) error {
	id := 0x01

	tx, err := tkeyclient.NewFrameBuf(cmdReset, id)
	if err != nil {
		return err
	}

	tx[2] = uint8(fwType)
	tx[3] = uint8(verifierDst)

	tkeyclient.Dump("reset tx", tx)

	if err = tk.Write(tx); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return nil
}

func getPubkey(tk *tkeyclient.TillitisKey) ([ed25519.PublicKeySize]byte, error) {
	id := 0x01

	tx, err := tkeyclient.NewFrameBuf(cmdGetPubkey, id)
	if err != nil {
		return [32]byte{}, fmt.Errorf("NewFrameBuf: %w", err)
	}

	tkeyclient.Dump("get pubkey tx", tx)

	if err = tk.Write(tx); err != nil {
		return [32]byte{}, err
	}

	rx, _, err := tk.ReadFrame(rspGetPubkey, id)
	if err != nil {
		return [32]byte{}, fmt.Errorf("ReadFrame: %w", err)
	}

	tkeyclient.Dump("get pubkey rx", rx)

	if rx[2] != tkeyclient.StatusOK {
		return [32]byte{}, fmt.Errorf("cmdGetPubkey not OK")
	}

	pubkey := [ed25519.PublicKeySize]byte{}
	copy(pubkey[:], rx[3:3+len(pubkey)])

	return pubkey, nil
}

func storePubkey(tk *tkeyclient.TillitisKey, pubkey [32]byte) error {
	id := 0x01

	tx, err := tkeyclient.NewFrameBuf(cmdStorePubkey, id)
	if err != nil {
		return err
	}

	copy(tx[2:], pubkey[:])

	tkeyclient.Dump("store pubkey tx", tx)

	if err = tk.Write(tx); err != nil {
		return err
	}

	// Read response
	const margin = 2
	tk.SetReadTimeoutNoErr(storePubkeyTimeout + margin)
	defer tk.SetReadTimeoutNoErr(0)

	rx, _, err := tk.ReadFrame(rspStorePubkey, id)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	tkeyclient.Dump("set pubkey rx", rx)

	if rx[2] != tkeyclient.StatusOK {
		return fmt.Errorf("cmdStorePubkey not OK")
	}

	return nil
}

func setPubkey(tk *tkeyclient.TillitisKey, pubkey [32]byte) error {
	id := 0x01

	tx, err := tkeyclient.NewFrameBuf(cmdSetPubkey, id)
	if err != nil {
		return err
	}

	copy(tx[2:], pubkey[:])

	tkeyclient.Dump("set pubkey tx", tx)

	if err = tk.Write(tx); err != nil {
		return err
	}

	// Read response
	rx, _, err := tk.ReadFrame(rspSetPubkey, id)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if rx[2] != tkeyclient.StatusOK {
		return fmt.Errorf("cmdSetPubkey not OK")
	}
	return nil
}

func updateAppInit(tk *tkeyclient.TillitisKey, size int, digest [blake2s.Size]byte, sig [ed25519.SignatureSize]byte) error {
	id := 0x01

	tx, err := tkeyclient.NewFrameBuf(cmdUpdateAppInit, id)
	if err != nil {
		return err
	}

	tx[2] = byte(size)
	tx[3] = byte(size >> 8)
	tx[4] = byte(size >> 16)
	tx[5] = byte(size >> 24)
	copy(tx[6:], digest[:])
	copy(tx[38:], sig[:])

	tkeyclient.Dump("update app1 tx", tx)

	if err = tk.Write(tx); err != nil {
		return err
	}

	// Read response
	rx, _, err := tk.ReadFrame(rspUpdateAppInit, id)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	tkeyclient.Dump("update app1 rx", rx)

	if rx[2] != tkeyclient.StatusOK {
		return fmt.Errorf("cmdUpdateAppInit not OK")
	}

	return nil
}

func writeChunk(tk *tkeyclient.TillitisKey, chunk []byte) error {
	id := 0x01

	tx, err := tkeyclient.NewFrameBuf(cmdUpdateAppChunk, id)
	if err != nil {
		return err
	}

	copy(tx[2:], chunk)

	tkeyclient.Dump("update app1 chunk tx", tx)

	if err = tk.Write(tx); err != nil {
		return fmt.Errorf("%w", err)
	}

	// Read response
	rx, _, err := tk.ReadFrame(rspUpdateAppChunk, id)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if rx[2] != tkeyclient.StatusOK {
		return fmt.Errorf("cmdUpdateAppChunk not OK")
	}

	return nil
}

// verify sends
// - framing header 1 byte
// - 0x01 (verify) 1 byte
// - digest 32 bytes
// - signature 64 bytes
func verify(tk *tkeyclient.TillitisKey, digest [blake2s.Size]byte, sig [ed25519.SignatureSize]byte) error {
	id := 0x01

	tx, err := tkeyclient.NewFrameBuf(cmdVerify, id)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	copy(tx[2:], digest[:])
	copy(tx[34:], sig[:])

	tkeyclient.Dump("verify tx", tx)

	if err = tk.Write(tx); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}
