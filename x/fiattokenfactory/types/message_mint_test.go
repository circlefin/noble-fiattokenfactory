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
	"testing"

	"cosmossdk.io/math"
	"github.com/circlefin/noble-fiattokenfactory/testutil/sample"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestMsgMint_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgMint
		err  error
	}{
		{
			name: "invalid from",
			msg: MsgMint{
				From:    "invalid_address",
				Address: sample.AccAddress(),
			},
			err: ErrInvalidAddress,
		},
		{
			name: "invalid recipient address",
			msg: MsgMint{
				From:    sample.AccAddress(),
				Address: "invalid_address",
			},
			err: ErrInvalidAddress,
		},
		{
			name: "empty recipient address",
			msg: MsgMint{
				From:    sample.AccAddress(),
				Address: "",
			},
			err: ErrInvalidAddress,
		},
		{
			name: "amount is empty",
			msg: MsgMint{
				From:    sample.AccAddress(),
				Address: sample.AccAddress(),
				Amount:  sdk.Coin{},
			},
			err: ErrInvalidCoins,
		},
		{
			name: "amount is missing",
			msg: MsgMint{
				From:    sample.AccAddress(),
				Address: sample.AccAddress(),
			},
			err: ErrInvalidCoins,
		},
		{
			name: "amount has missing amount",
			msg: MsgMint{
				From:    sample.AccAddress(),
				Address: sample.AccAddress(),
				Amount:  sdk.Coin{Denom: "uusdc"},
			},
			err: ErrInvalidCoins,
		},
		{
			name: "amount is zero",
			msg: MsgMint{
				From:    sample.AccAddress(),
				Address: sample.AccAddress(),
				Amount:  sdk.Coin{Denom: "uusdc", Amount: math.NewInt(0)},
			},
			err: ErrInvalidCoins,
		},
		{
			name: "amount is negative",
			msg: MsgMint{
				From:    sample.AccAddress(),
				Address: sample.AccAddress(),
				Amount:  sdk.Coin{Denom: "uusdc", Amount: math.NewInt(-1)},
			},
			err: ErrInvalidCoins,
		},
		{
			name: "happy path",
			msg: MsgMint{
				From:    sample.AccAddress(),
				Address: sample.AccAddress(),
				Amount:  sdk.Coin{Denom: "uusdc", Amount: math.NewInt(1)},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}
