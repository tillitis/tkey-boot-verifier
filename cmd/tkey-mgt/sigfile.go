// SPDX-FileCopyrightText: 2023 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strings"
)

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

func readKey(filename string) (*pubKey, error) {
	var pub pubKey

	buf, err := readBase64(filename)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	r := bytes.NewReader(buf)
	err = binary.Read(r, binary.BigEndian, &pub)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return &pub, nil
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

// writeBase64 encodes data in base64 and writes it the file given in
// filename. If overwrite is true it overwrites any existing file,
// otherwise it returns an error.
func writeBase64(filename string, data any, comment string, overwrite bool) error {
	var buf bytes.Buffer

	err := binary.Write(&buf, binary.BigEndian, data)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	b64 := base64.StdEncoding.EncodeToString(buf.Bytes())
	b64 += "\n"

	var f *os.File

	f, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o666)
	if err != nil {
		if os.IsExist(err) && overwrite {
			f, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0o666)
			if err != nil {
				return fmt.Errorf("%w", err)
			}
		} else {
			return fmt.Errorf("%w", err)
		}
	}

	defer f.Close()

	_, err = f.Write([]byte(fmt.Sprintf("untrusted comment: %s\n", comment)))
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	_, err = f.Write([]byte(b64))
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

// writeRetry writes the data in the file given in filename as base64.
// If a file already exists it prompts interactively for permission to
// overwrite the file.
func writeRetry(filename string, data any, comment string) error {
	err := writeBase64(filename, data, comment, false)
	if os.IsExist(errors.Unwrap(err)) {
		le.Printf("File %v exists. Overwrite [y/n]?", filename)
		reader := bufio.NewReader(os.Stdin)
		overWriteP, _ := reader.ReadString('\n')

		// Trim space to normalize Windows line endings
		overWriteP = strings.TrimSpace(overWriteP)

		if overWriteP == "y" {
			err = writeBase64(filename, data, comment, true)
		} else {
			le.Printf("Aborted\n")
			os.Exit(1)
		}
	}

	if !os.IsExist(errors.Unwrap(err)) && err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}
