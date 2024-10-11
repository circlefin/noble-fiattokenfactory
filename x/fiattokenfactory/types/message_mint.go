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
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (msg *MsgMint) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return errors.Wrapf(ErrInvalidAddress, "invalid from address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return errors.Wrapf(ErrInvalidAddress, "invalid address (%s)", err)
	}

	if msg.Amount.IsNil() {
		return errors.Wrap(ErrInvalidCoins, "mint amount cannot be nil")
	}

	if msg.Amount.IsNegative() {
		return errors.Wrap(ErrInvalidCoins, "mint amount cannot be negative")
	}

	if msg.Amount.IsZero() {
		return errors.Wrap(ErrInvalidCoins, "mint amount cannot be zero")
	}

	return nil
}
