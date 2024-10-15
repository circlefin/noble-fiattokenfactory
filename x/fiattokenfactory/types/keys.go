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

const (
	// ModuleName defines the module name
	ModuleName = "fiat-tokenfactory"

	// StoreKey defines the primary module store key
	StoreKey = "fiattokenfactory"

	// RouterKey defines the module's message routing key
	RouterKey = StoreKey

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_" + StoreKey

	PausedKey                 = "Paused/value/"
	MasterMinterKey           = "MasterMinter/value/"
	PauserKey                 = "Pauser/value/"
	BlacklisterKey            = "Blacklister/value/"
	OwnerKey                  = "Owner/value/"
	PendingOwnerKey           = "PendingOwner/value/"
	BlacklistedKeyPrefix      = "Blacklisted/value/"
	MintersKeyPrefix          = "Minters/value/"
	MinterControllerKeyPrefix = "MinterController/value/"

	GranteeKey = "SendRestrictionGrantees"
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

// BlacklistedKey returns the store key to retrieve a Blacklisted from the index fields
func BlacklistedKey(addressBz []byte) []byte {
	return append(addressBz, []byte("/")...)
}

// MintersKey returns the store key to retrieve a Minters from the index fields
func MintersKey(address string) []byte {
	return append([]byte(address), []byte("/")...)
}

// MinterControllerKey returns the store key to retrieve a MinterController from the index fields
func MinterControllerKey(controllerAddress string) []byte {
	return append([]byte(controllerAddress), []byte("/")...)
}

const (
	MintingDenomKey = "MintingDenom/value/"
)
