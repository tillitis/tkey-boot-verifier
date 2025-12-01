// SPDX-FileCopyrightText: 2025 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

package main

import (
	"fmt"

	"github.com/tillitis/tkeyclient"
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
	cmdReset = appCmd{0xfe, "cmdReset", tkeyclient.CmdLen4}
)

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

func fwResetTypeFromInt(i int) (fwResetType, error) {
	if i < int(fwResetTypeStartDefault) || i > int(fwResetTypeStartClientVer) {
		return 0, fmt.Errorf("invalid reset type: %d", i)
	}

	return fwResetType(i), nil
}

type resetDst uint8

const (
	verifierResetDstApp1    resetDst = 0
	verifierResetDstCmdMode resetDst = 1
)

func resetDstFromInt(i int) (resetDst, error) {
	if i < int(verifierResetDstApp1) || i > int(verifierResetDstCmdMode) {
		return 0, fmt.Errorf("invalid reset dst: %d", i)
	}

	return resetDst(i), nil
}

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
