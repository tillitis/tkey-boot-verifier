// SPDX-FileCopyrightText: 2026 Tillitis AB <tillitis.se>
// SPDX-License-Identifier: BSD-2-Clause

package main 

import (
	"encoding/json"
	"fmt"
	"os"

	"tkey-mgt/sigfile"
)

type DistInfo struct {
	Name        string
	AppDigest   [32]byte
	Installable bool
	Signature   sigfile.Signature
	PubKey      sigfile.PubKey
}


func ReadDistInfo(fileName string) (DistInfo, error) {
	var info DistInfo

	input, err := os.ReadFile(fileName)
	if err != nil {
		return DistInfo{}, fmt.Errorf("read dist info: %w", err)
	}

	if err := json.Unmarshal(input, &info); err != nil {
		return DistInfo{}, fmt.Errorf("decode dist info: %w", err)
	}

	return info, nil
}
func (d *DistInfo) ToBuffer() ([]byte, error) {
	j, err := json.Marshal(d)
	if err != nil {
		return []byte{}, fmt.Errorf("JSON marshalling failed: %w", err)
	}

	return j, nil
}

func (d *DistInfo) ToFile(fileName string, overwrite bool) error {
	buf, err := d.ToBuffer()
	if err != nil {
		return err
	}

	return writeFile(fileName, buf, overwrite)
}

// writeFile writes buf into file filename. If overwrite is true it
// overwrites any existing file, otherwise it returns an error.
func writeFile(filename string, buf []byte, overwrite bool) error {
	var f *os.File

	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o666)
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

	if _, err := f.Write(buf); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}
