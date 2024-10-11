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

package keeper_test

import (
	"fmt"
	"testing"

	testkeeper "github.com/circlefin/noble-fiattokenfactory/testutil/keeper"
	"github.com/circlefin/noble-fiattokenfactory/testutil/sample"
	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/keeper"
	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestRemoveMinter_FromAddressIsNotController(t *testing.T) {
	mintingDenom := "uusdc"
	ftf, ctx := testkeeper.FiatTokenfactoryKeeper()
	ftf.SetMintingDenom(ctx, types.MintingDenom{Denom: mintingDenom})
	msgServer := keeper.NewMsgServerImpl(ftf)

	_, err := msgServer.RemoveMinter(sdk.WrapSDKContext(ctx), &types.MsgRemoveMinter{From: sample.AccAddress()})
	require.ErrorIs(t, err, types.ErrUnauthorized)
	require.ErrorContains(t, err, "minter controller not found")
}

func TestRemoveMinter_ControllerMinterMismatch(t *testing.T) {
	mintingDenom := "uusdc"
	controller := sample.TestAccount()
	minter := sample.TestAccount()
	notMinter := sample.AccAddress()
	ftf, ctx := testkeeper.FiatTokenfactoryKeeper()
	ftf.SetMintingDenom(ctx, types.MintingDenom{Denom: mintingDenom})
	msgServer := keeper.NewMsgServerImpl(ftf)
	ftf.SetMinterController(ctx, types.MinterController{Minter: minter.Address, Controller: controller.Address})

	_, err := msgServer.RemoveMinter(sdk.WrapSDKContext(ctx), &types.MsgRemoveMinter{From: controller.Address, Address: notMinter})
	require.ErrorIs(t, err, types.ErrUnauthorized)
	require.ErrorContains(t, err, fmt.Sprintf("minter address ≠ minter controller's minter address, (%s≠%s)", notMinter, minter.Address))
}

func TestRemoveMinter_MinterNotConfigured(t *testing.T) {
	mintingDenom := "uusdc"
	controller := sample.TestAccount()
	minter := sample.TestAccount()
	ftf, ctx := testkeeper.FiatTokenfactoryKeeper()
	ftf.SetMintingDenom(ctx, types.MintingDenom{Denom: mintingDenom})
	msgServer := keeper.NewMsgServerImpl(ftf)
	ftf.SetMinterController(ctx, types.MinterController{Minter: minter.Address, Controller: controller.Address})

	_, err := msgServer.RemoveMinter(sdk.WrapSDKContext(ctx), &types.MsgRemoveMinter{From: controller.Address, Address: minter.Address})
	require.ErrorIs(t, err, types.ErrUserNotFound)
	require.ErrorContains(t, err, "a minter with a given address doesn't exist")
}

func TestRemoveMinter_Success(t *testing.T) {
	mintingDenom := "uusdc"
	controller := sample.TestAccount()
	minter := sample.TestAccount()
	ftf, ctx := testkeeper.FiatTokenfactoryKeeper()
	ftf.SetMintingDenom(ctx, types.MintingDenom{Denom: mintingDenom})
	msgServer := keeper.NewMsgServerImpl(ftf)
	ftf.SetMinterController(ctx, types.MinterController{Minter: minter.Address, Controller: controller.Address})
	ftf.SetMinters(ctx, types.Minters{Address: minter.Address})

	res, err := msgServer.RemoveMinter(sdk.WrapSDKContext(ctx), &types.MsgRemoveMinter{From: controller.Address, Address: minter.Address})
	require.NoError(t, err)
	require.Equal(t, &types.MsgRemoveMinterResponse{}, res)
}
