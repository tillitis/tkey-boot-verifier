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
	cmdGetCDI  = appCmd{0x01, "cmdGetCDI", tkeyclient.CmdLen1}
	rspGetCDI  = appCmd{0x01, "rspGetCDI", tkeyclient.CmdLen128}
	cmdGetNameVersion  = appCmd{0x02, "cmdGetNameVersion", tkeyclient.CmdLen1}
	rspGetNameVersion  = appCmd{0x02, "rspGetNameVersion", tkeyclient.CmdLen32}
)

func resetTypeFromInt(i int) (tkeyclient.ResetType, error) {
	if i < int(tkeyclient.RstTypeStartFlash0) || i > int(tkeyclient.RstTypeStartClientVer) {
		return 0, fmt.Errorf("invalid reset type: %d", i)
	}

	return tkeyclient.ResetType(i), nil
}

func nextAppDataFromInt(i int) (tkeyclient.NextAppData, error) {
	if i < 0 || i > 1 {
		return tkeyclient.NextAppData{}, fmt.Errorf("invalid reset dst: %d", i)
	}

	var d = tkeyclient.NewNextAppDataFromSlice([]byte{byte(i)})

	return d, nil
}

func getCDI(tk *tkeyclient.TillitisKey) (string, error) {
	id := 0x01
	tx, err := tkeyclient.NewFrameBuf(cmdGetCDI, id)
	if err != nil {
		return "", fmt.Errorf("NewFrameBuf: %w", err)
	}

	tkeyclient.Dump("GetCDI tx", tx)
	if err = tk.Write(tx); err != nil {
		return "", fmt.Errorf("write: %w", err)
	}

	tk.SetReadTimeoutNoErr(2)
	defer tk.SetReadTimeoutNoErr(0)

	rx, _, err := tk.ReadFrame(rspGetCDI, id)
	if err != nil {
		return "", fmt.Errorf("ReadFrame: %w", err)
	}

	cdi := fmt.Sprintf("%064x", rx[2:34])

	return cdi, nil
}

func getNameVersion(tk *tkeyclient.TillitisKey) (*tkeyclient.NameVersion, error) {
	id := 0x01
	tx, err := tkeyclient.NewFrameBuf(cmdGetNameVersion, id)
	if err != nil {
		return nil, fmt.Errorf("NewFrameBuf: %w", err)
	}

	tkeyclient.Dump("GetNameVersion tx", tx)
	if err = tk.Write(tx); err != nil {
		return nil, fmt.Errorf("write: %w", err)
	}

	tk.SetReadTimeoutNoErr(2)
	defer tk.SetReadTimeoutNoErr(0)

	rx, _, err := tk.ReadFrame(rspGetNameVersion, id)
	if err != nil {
		return nil, fmt.Errorf("ReadFrame: %w", err)
	}

	nameVer := &tkeyclient.NameVersion{}
	nameVer.Unpack(rx[2:])

	return nameVer, nil
}
