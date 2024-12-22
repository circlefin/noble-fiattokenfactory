// Copyright 2024 Circle Internet Group, Inc.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"github.com/btcsuite/btcd/btcutil/bech32"
)

// DecodeAddress decodes the given address, returning the base32 encoded address bytes
// Support both bech32 and bech32m
func DecodeAddress(address string) (string, []byte, error) {
	return bech32.DecodeNoLimit(address)
}

// ConvertToBase256 converts base32 encoded address to base256
func ConvertToBase256(address []byte) ([]byte, error) {
	return bech32.ConvertBits(address, 5, 8, false)
}

// MustConvertToBase256 converts base32 encoded address to base256
// Panic if there is any error
func MustConvertToBase256(address []byte) []byte {
	bz, err := bech32.ConvertBits(address, 5, 8, false)
	if err != nil {
		panic(err)
	}

	return bz
}
