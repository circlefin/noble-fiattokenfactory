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

	"github.com/btcsuite/btcd/btcutil/bech32"
	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) Blacklist(goCtx context.Context, msg *types.MsgBlacklist) (*types.MsgBlacklistResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	blacklister, found := k.GetBlacklister(ctx)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUserNotFound, "blacklister is not set")
	}

	if blacklister.Address != msg.From {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "you are not the blacklister")
	}

	_, addressBz, err := bech32.DecodeToBase256(msg.Address)
	if err != nil {
		return nil, err
	}

	_, found = k.GetBlacklisted(ctx, addressBz)
	if found {
		return nil, types.ErrUserBlacklisted
	}

	blacklisted := types.Blacklisted{
		AddressBz: addressBz,
	}

	k.SetBlacklisted(ctx, blacklisted)

	err = ctx.EventManager().EmitTypedEvent(msg)

	return &types.MsgBlacklistResponse{}, err
}
