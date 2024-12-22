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
	"fmt"

	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"

	"cosmossdk.io/core/store"

	"cosmossdk.io/errors"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		logger       log.Logger
		storeService store.KVStoreService

		bankKeeper types.BankKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	logger log.Logger,
	storeService store.KVStoreService,

	bankKeeper types.BankKeeper,
) *Keeper {
	return &Keeper{
		cdc:          cdc,
		logger:       logger,
		storeService: storeService,
		bankKeeper:   bankKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return k.logger.With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SendRestrictionFn checks every $USDC transfer executed on the Noble chain against the blocklist and paused state
func (k Keeper) SendRestrictionFn(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) (newToAddr sdk.AccAddress, err error) {
	mintingDenom := k.GetMintingDenom(ctx)
	if amount := amt.AmountOf(mintingDenom.Denom); !amount.IsZero() {
		paused := k.GetPaused(ctx)
		if paused.Paused {
			return toAddr, errors.Wrapf(types.ErrPaused, "cannot perform token transfers")
		}

		_, found := k.GetBlacklisted(ctx, fromAddr.Bytes())
		if found {
			return toAddr, errors.Wrapf(types.ErrUnauthorized, "an address (%s) is blacklisted and can not send tokens", fromAddr.String())
		}

		_, found = k.GetBlacklisted(ctx, toAddr.Bytes())
		if found {
			return toAddr, errors.Wrapf(types.ErrUnauthorized, "an address (%s) is blacklisted and can not receive tokens", toAddr.String())
		}

		grantees := ctx.Value(types.GranteeKey)
		if grantees != nil {
			for _, grantee := range grantees.([]string) {
				_, addressBz, err := types.DecodeAddress(grantee)
				if err != nil {
					return toAddr, err
				}
				_, found := k.GetBlacklisted(ctx, addressBz)
				if found {
					return toAddr, errors.Wrapf(types.ErrUnauthorized, "an address (%s) is blacklisted and can not authorize tokens", toAddr.String())
				}
			}
		}
	}

	return toAddr, nil
}

// ValidatePrivileges checks if a specified address has already been assigned to a privileged role.
func (k Keeper) ValidatePrivileges(ctx context.Context, address string) error {
	acc, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return err
	}

	owner, found := k.GetOwner(ctx)
	if found && owner.Address == acc.String() {
		return errors.Wrapf(types.ErrAlreadyPrivileged, "cannot assign (%s) to owner role", acc.String())
	}

	blacklister, found := k.GetBlacklister(ctx)
	if found && blacklister.Address == acc.String() {
		return errors.Wrapf(types.ErrAlreadyPrivileged, "cannot assign (%s) to black lister role", acc.String())
	}

	masterminter, found := k.GetMasterMinter(ctx)
	if found && masterminter.Address == acc.String() {
		return errors.Wrapf(types.ErrAlreadyPrivileged, "cannot assign (%s) to master minter role", acc.String())
	}

	pauser, found := k.GetPauser(ctx)
	if found && pauser.Address == acc.String() {
		return errors.Wrapf(types.ErrAlreadyPrivileged, "cannot assign (%s) to pauser role", acc.String())
	}

	return nil
}
