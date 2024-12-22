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

	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) Burn(goCtx context.Context, msg *types.MsgBurn) (*types.MsgBurnResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return k.Keeper.Burn(ctx, msg)
}

func (k Keeper) Burn(ctx sdk.Context, msg *types.MsgBurn) (*types.MsgBurnResponse, error) {
	_, found := k.GetMinters(ctx, msg.From)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrBurn, "%v: you are not a minter", types.ErrUnauthorized)
	}

	_, addressBz, err := types.DecodeAddress(msg.From)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrBurn, err.Error())
	}

	_, found = k.GetBlacklisted(ctx, addressBz)
	if found {
		return nil, sdkerrors.Wrap(types.ErrBurn, "minter address is blacklisted")
	}

	mintingDenom := k.GetMintingDenom(ctx)

	if msg.Amount.Denom != mintingDenom.Denom {
		return nil, sdkerrors.Wrap(types.ErrBurn, "burning denom is incorrect")
	}

	if msg.Amount.IsNil() || !msg.Amount.IsPositive() {
		return nil, sdkerrors.Wrap(types.ErrBurn, "burning amount is invalid")
	}

	paused := k.GetPaused(ctx)

	if paused.Paused {
		return nil, sdkerrors.Wrap(types.ErrBurn, "burning is paused")
	}

	minterAddress, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrBurn, err.Error())
	}

	amount := sdk.NewCoins(msg.Amount)

	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, minterAddress, types.ModuleName, amount)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrBurn, err.Error())
	}

	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, amount); err != nil {
		return nil, sdkerrors.Wrap(types.ErrBurn, err.Error())
	}

	err = ctx.EventManager().EmitTypedEvent(msg)

	return &types.MsgBurnResponse{}, err
}
