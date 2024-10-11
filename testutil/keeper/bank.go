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

package keeper

import (
	"context"

	"cosmossdk.io/errors"

	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

var ErrInsufficientFunds = errors.Register("mockbankkeeper", 1, "insufficient funds")

type MockBankKeeper struct {
	Balances map[string]sdk.Coins
}

var _ types.BankKeeper = MockBankKeeper{}

func (MockBankKeeper) SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins {
	return nil
}

func (k MockBankKeeper) MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error {
	address := authtypes.NewModuleAddress(moduleName).String()
	k.Balances[address] = k.Balances[address].Add(amt...)
	return nil
}

func (k MockBankKeeper) BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error {
	address := authtypes.NewModuleAddress(moduleName).String()

	prevBalance := k.Balances[address]
	newBalance, negative := prevBalance.SafeSub(amt...)
	if negative {
		return errors.Wrapf(ErrInsufficientFunds, "%s is smaller than %s", prevBalance, amt)
	}
	k.Balances[address] = newBalance
	return nil
}

func (k MockBankKeeper) SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	senderAddr := authtypes.NewModuleAddress(senderModule)

	return k.SendCoins(ctx, senderAddr, recipientAddr, amt)
}

func (k MockBankKeeper) SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	recipientAddr := authtypes.NewModuleAddress(recipientModule)

	return k.SendCoins(ctx, senderAddr, recipientAddr, amt)
}

func (k MockBankKeeper) SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	prevBalance := k.Balances[fromAddr.String()]
	newBalance, negative := prevBalance.SafeSub(amt...)
	if negative {
		return errors.Wrapf(ErrInsufficientFunds, "%s is smaller than %s", prevBalance, amt)
	}
	k.Balances[fromAddr.String()] = newBalance
	k.Balances[toAddr.String()] = k.Balances[toAddr.String()].Add(amt...)

	return nil
}

func (MockBankKeeper) GetDenomMetaData(ctx context.Context, denom string) (banktypes.Metadata, bool) {
	if denom == "uusdc" {
		return banktypes.Metadata{
			DenomUnits: []*banktypes.DenomUnit{
				{
					Denom: "uusdc",
				},
			},
		}, true
	}
	return banktypes.Metadata{}, false
}
