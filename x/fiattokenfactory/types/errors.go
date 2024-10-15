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

// DONTCOVER

import (
	"cosmossdk.io/errors"
)

// x/fiattokenfactory module sentinel errors
var (
	ErrUnauthorized       = errors.Register(ModuleName, 2, "unauthorized")
	ErrUserNotFound       = errors.Register(ModuleName, 3, "user not found")
	ErrMint               = errors.Register(ModuleName, 4, "tokens can not be minted")
	ErrSendCoinsToAccount = errors.Register(ModuleName, 5, "can't send tokens to account")
	ErrBurn               = errors.Register(ModuleName, 6, "tokens can not be burned")
	ErrPaused             = errors.Register(ModuleName, 7, "the chain is paused")
	ErrMintingDenomSet    = errors.Register(ModuleName, 9, "the minting denom has already been set")
	ErrUserBlacklisted    = errors.Register(ModuleName, 10, "user is already blacklisted")
	ErrAlreadyPrivileged  = errors.Register(ModuleName, 11, "address is already assigned to privileged role")
	ErrDenomNotRegistered = errors.Register(ModuleName, 12, "denom not registered in bank module")

	ErrInvalidAddress = errors.Register(ModuleName, 100, "invalid address")
	ErrInvalidCoins   = errors.Register(ModuleName, 101, "invalid coins")
	ErrInvalidType    = errors.Register(ModuleName, 102, "invalid type")
)
