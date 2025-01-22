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

func (k msgServer) Mint(goCtx context.Context, msg *types.MsgMint) (*types.MsgMintResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return k.Keeper.Mint(ctx, msg)
}

func (k Keeper) Mint(ctx sdk.Context, msg *types.MsgMint) (*types.MsgMintResponse, error) {
	minter, found := k.GetMinters(ctx, msg.From)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "you are not a minter")
	}

	_, addressBz, err := DecodeNoLimitToBase256(msg.From)
	if err != nil {
		return nil, err
	}

	_, found = k.GetBlacklisted(ctx, addressBz)
	if found {
		return nil, sdkerrors.Wrapf(types.ErrMint, "minter address is blacklisted")
	}

	_, addressBz, err = DecodeNoLimitToBase256(msg.Address)
	if err != nil {
		return nil, err
	}

	_, found = k.GetBlacklisted(ctx, addressBz)
	if found {
		return nil, sdkerrors.Wrapf(types.ErrMint, "receiver address is blacklisted")
	}

	mintingDenom := k.GetMintingDenom(ctx)

	if msg.Amount.Denom != mintingDenom.Denom {
		return nil, sdkerrors.Wrapf(types.ErrMint, "minting denom is incorrect")
	}

	if msg.Amount.IsNil() || !msg.Amount.IsPositive() {
		return nil, sdkerrors.Wrap(types.ErrMint, "minting amount is invalid")
	}

	if minter.Allowance.IsLT(msg.Amount) {
		return nil, sdkerrors.Wrapf(types.ErrMint, "minting amount is greater than the allowance")
	}

	paused := k.GetPaused(ctx)

	if paused.Paused {
		return nil, sdkerrors.Wrapf(types.ErrMint, "minting is paused")
	}

	minter.Allowance = minter.Allowance.Sub(msg.Amount)

	k.SetMinters(ctx, minter)

	amount := sdk.NewCoins(msg.Amount)

	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, amount); err != nil {
		return nil, sdkerrors.Wrap(types.ErrMint, err.Error())
	}

	receiver, _ := sdk.AccAddressFromBech32(msg.Address)

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiver, amount); err != nil {
		return nil, sdkerrors.Wrap(types.ErrSendCoinsToAccount, err.Error())
	}

	err = ctx.EventManager().EmitTypedEvent(msg)

	return &types.MsgMintResponse{}, err
}
