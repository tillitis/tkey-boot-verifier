// SPDX-FileCopyrightText: 2023 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

type signature struct {
	Alg    [2]byte
	KeyNum [8]byte
	Sig    [64]byte
}

// readBase64 reads the file in filename with base64, decodes it and
// returns a binary representation
func readBase64(filename string) ([]byte, error) {
	input, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	lines := strings.Split(string(input), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("Too few lines in file %s", filename)
	}

	data, err := base64.StdEncoding.DecodeString(lines[1])
	if err != nil {
		return nil, fmt.Errorf("could not decode: %w", err)
	}

	return data, nil
}

func readSig(filename string) (*signature, error) {
	var sig signature

	buf, err := readBase64(filename)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	r := bytes.NewReader(buf)
	err = binary.Read(r, binary.BigEndian, &sig)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return &sig, nil
}
