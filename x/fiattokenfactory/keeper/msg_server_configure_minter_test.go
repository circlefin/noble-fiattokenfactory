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
	"testing"

	"cosmossdk.io/math"
	testkeeper "github.com/circlefin/noble-fiattokenfactory/testutil/keeper"
	"github.com/circlefin/noble-fiattokenfactory/testutil/sample"
	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/keeper"
	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestConfigureMinter_DenomIsMissing(t *testing.T) {
	mintingDenom := "uusdc"
	allowance := sdk.Coin{Amount: math.NewInt(1)}
	_, ctx, msgServer := setupForConfigureMinterTest(mintingDenom)

	_, err := msgServer.ConfigureMinter(sdk.WrapSDKContext(ctx), &types.MsgConfigureMinter{Allowance: allowance})
	require.ErrorIs(t, err, types.ErrMint)
	require.ErrorContains(t, err, "minting denom is incorrect")
}

func TestConfigureMinter_DenomIsEmpty(t *testing.T) {
	mintingDenom := "uusdc"
	allowance := sdk.Coin{Denom: "", Amount: math.NewInt(1)}
	_, ctx, msgServer := setupForConfigureMinterTest(mintingDenom)

	_, err := msgServer.ConfigureMinter(sdk.WrapSDKContext(ctx), &types.MsgConfigureMinter{Allowance: allowance})
	require.ErrorIs(t, err, types.ErrMint)
	require.ErrorContains(t, err, "minting denom is incorrect")
}

func TestConfigureMinter_IncorrectDenom(t *testing.T) {
	mintingDenom := "uusdc"
	allowance := sdk.Coin{Denom: "fakeDenom", Amount: math.NewInt(1)}
	_, ctx, msgServer := setupForConfigureMinterTest(mintingDenom)

	_, err := msgServer.ConfigureMinter(sdk.WrapSDKContext(ctx), &types.MsgConfigureMinter{Allowance: allowance})
	require.ErrorIs(t, err, types.ErrMint)
	require.ErrorContains(t, err, "minting denom is incorrect")
}

func TestConfigureMinter_NilAllowance(t *testing.T) {
	mintingDenom := "uusdc"
	allowance := sdk.Coin{Denom: mintingDenom}
	_, ctx, msgServer := setupForConfigureMinterTest(mintingDenom)

	_, err := msgServer.ConfigureMinter(sdk.WrapSDKContext(ctx), &types.MsgConfigureMinter{Allowance: allowance})
	require.ErrorIs(t, err, types.ErrMint)
	require.ErrorContains(t, err, "allowance amount is invalid")
}

func TestConfigureMinter_NegativeAllowance(t *testing.T) {
	mintingDenom := "uusdc"
	allowance := sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(-1)}
	_, ctx, msgServer := setupForConfigureMinterTest(mintingDenom)

	_, err := msgServer.ConfigureMinter(sdk.WrapSDKContext(ctx), &types.MsgConfigureMinter{Allowance: allowance})
	require.ErrorIs(t, err, types.ErrMint)
	require.ErrorContains(t, err, "allowance amount is invalid")
}

func TestConfigureMinter_FromAddressIsNotController(t *testing.T) {
	mintingDenom := "uusdc"
	allowance := sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(1)}
	_, ctx, msgServer := setupForConfigureMinterTest(mintingDenom)

	_, err := msgServer.ConfigureMinter(sdk.WrapSDKContext(ctx), &types.MsgConfigureMinter{From: sample.AccAddress(), Allowance: allowance})
	require.ErrorIs(t, err, types.ErrUnauthorized)
	require.ErrorContains(t, err, "minter controller not found")
}

func TestConfigureMinter_ControllerMinterMismatch(t *testing.T) {
	mintingDenom := "uusdc"
	allowance := sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(1)}
	controller := sample.TestAccount()
	minter := sample.TestAccount()
	ftf, ctx, msgServer := setupForConfigureMinterTest(mintingDenom)
	ftf.SetMinterController(ctx, types.MinterController{Controller: controller.Address, Minter: minter.Address})

	_, err := msgServer.ConfigureMinter(sdk.WrapSDKContext(ctx), &types.MsgConfigureMinter{From: controller.Address, Address: sample.AccAddress(), Allowance: allowance})
	require.ErrorIs(t, err, types.ErrUnauthorized)
	require.ErrorContains(t, err, "minter address ≠ minter controller's minter address")
}

func TestConfigureMinter_Paused(t *testing.T) {
	mintingDenom := "uusdc"
	allowance := sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(1)}
	controller := sample.TestAccount()
	minter := sample.TestAccount()
	ftf, ctx, msgServer := setupForConfigureMinterTest(mintingDenom)
	ftf.SetMinterController(ctx, types.MinterController{Controller: controller.Address, Minter: minter.Address})
	ftf.SetPaused(ctx, types.Paused{Paused: true})

	_, err := msgServer.ConfigureMinter(sdk.WrapSDKContext(ctx), &types.MsgConfigureMinter{From: controller.Address, Address: minter.Address, Allowance: allowance})
	require.ErrorIs(t, err, types.ErrMint)
	require.ErrorContains(t, err, "minting is paused")
}

func TestConfigureMinter_ZeroAllowanceSuccess(t *testing.T) {
	mintingDenom := "uusdc"
	allowance := sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(0)}
	controller := sample.TestAccount()
	minter := sample.TestAccount()
	ftf, ctx, msgServer := setupForConfigureMinterTest(mintingDenom)
	ftf.SetMinterController(ctx, types.MinterController{Controller: controller.Address, Minter: minter.Address})

	res, err := msgServer.ConfigureMinter(sdk.WrapSDKContext(ctx), &types.MsgConfigureMinter{From: controller.Address, Address: minter.Address, Allowance: allowance})
	require.NoError(t, err)
	require.Equal(t, &types.MsgConfigureMinterResponse{}, res)
}

func TestConfigureMinter_Success(t *testing.T) {
	mintingDenom := "uusdc"
	allowance := sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(1)}
	controller := sample.TestAccount()
	minter := sample.TestAccount()
	ftf, ctx, msgServer := setupForConfigureMinterTest(mintingDenom)
	ftf.SetMinterController(ctx, types.MinterController{Controller: controller.Address, Minter: minter.Address})

	res, err := msgServer.ConfigureMinter(sdk.WrapSDKContext(ctx), &types.MsgConfigureMinter{From: controller.Address, Address: minter.Address, Allowance: allowance})
	require.NoError(t, err)
	require.Equal(t, &types.MsgConfigureMinterResponse{}, res)
}

func setupForConfigureMinterTest(mintingDenom string) (*keeper.Keeper, sdk.Context, types.MsgServer) {
	ftf, ctx := testkeeper.FiatTokenfactoryKeeper()
	ftf.SetMintingDenom(ctx, types.MintingDenom{Denom: mintingDenom})
	ftf.SetPaused(ctx, types.Paused{Paused: false})
	msgServer := keeper.NewMsgServerImpl(ftf)
	return ftf, ctx, msgServer
}
