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

func TestBurn_FromAddressIsMissing(t *testing.T) {
	mintingDenom := "uusdc"
	_, ctx, msgServer := setupForBurnTest(mintingDenom)

	_, err := msgServer.Burn(sdk.WrapSDKContext(ctx), &types.MsgBurn{})
	require.ErrorIs(t, err, types.ErrBurn)
	require.ErrorContains(t, err, "you are not a minter")
}

func TestBurn_FromAddressIsNotMinter(t *testing.T) {
	mintingDenom := "uusdc"
	_, ctx, msgServer := setupForBurnTest(mintingDenom)

	_, err := msgServer.Burn(sdk.WrapSDKContext(ctx), &types.MsgBurn{From: "notMinter"})
	require.ErrorIs(t, err, types.ErrBurn)
	require.ErrorContains(t, err, "you are not a minter")
}

func TestBurn_InvalidMinterAddress(t *testing.T) {
	mintingDenom := "uusdc"
	ftf, ctx, msgServer := setupForBurnTest(mintingDenom)
	amount := sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(1)}

	invalidMinter := make([]byte, 256)
	invalidMinterAddress := sdk.AccAddress(invalidMinter).String()
	ftf.SetMinters(ctx, types.Minters{Address: invalidMinterAddress})
	_, err := msgServer.Burn(sdk.WrapSDKContext(ctx), &types.MsgBurn{From: invalidMinterAddress, Amount: amount})
	require.ErrorIs(t, err, types.ErrBurn)
}

func TestBurn_BlacklistedMinterAddress(t *testing.T) {
	mintingDenom := "uusdc"
	minter := sample.TestAccount()
	amount := sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(1)}
	ftf, ctx, msgServer := setupForBurnTest(mintingDenom)

	ftf.SetBlacklisted(ctx, types.Blacklisted{AddressBz: minter.AddressBz})
	ftf.SetMinters(ctx, types.Minters{Address: minter.Address})
	_, err := msgServer.Burn(sdk.WrapSDKContext(ctx), &types.MsgBurn{From: minter.Address, Amount: amount})
	require.ErrorIs(t, err, types.ErrBurn)
	require.ErrorContains(t, err, "minter address is blacklisted")
}

func TestBurn_BlacklistedBech32mMinterAddress(t *testing.T) {
	mintingDenom := "uusdc"
	minter := sample.TestAccountBech32m()
	amount := sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(1)}
	ftf, ctx, msgServer := setupForBurnTest(mintingDenom)

	ftf.SetBlacklisted(ctx, types.Blacklisted{AddressBz: minter.AddressBz})
	ftf.SetMinters(ctx, types.Minters{Address: minter.Address})
	_, err := msgServer.Burn(sdk.WrapSDKContext(ctx), &types.MsgBurn{From: minter.Address, Amount: amount})
	require.ErrorIs(t, err, types.ErrBurn)
	require.ErrorContains(t, err, "minter address is blacklisted")
}

func TestBurn_DenomIsMissing(t *testing.T) {
	mintingDenom := "uusdc"
	minter := sample.TestAccount()
	ftf, ctx, msgServer := setupForBurnTest(mintingDenom)

	ftf.SetMinters(ctx, types.Minters{Address: minter.Address})
	_, err := msgServer.Burn(sdk.WrapSDKContext(ctx), &types.MsgBurn{From: minter.Address, Amount: sdk.Coin{}})
	require.ErrorIs(t, err, types.ErrBurn)
	require.ErrorContains(t, err, "burning denom is incorrect")
}

func TestBurn_DenomIsEmpty(t *testing.T) {
	mintingDenom := "uusdc"
	minter := sample.TestAccount()
	ftf, ctx, msgServer := setupForBurnTest(mintingDenom)

	ftf.SetMinters(ctx, types.Minters{Address: minter.Address})
	_, err := msgServer.Burn(sdk.WrapSDKContext(ctx), &types.MsgBurn{From: minter.Address, Amount: sdk.Coin{Denom: ""}})
	require.ErrorIs(t, err, types.ErrBurn)
	require.ErrorContains(t, err, "burning denom is incorrect")
}

func TestBurn_DenomIsIncorrect(t *testing.T) {
	mintingDenom := "uusdc"
	minter := sample.TestAccount()
	ftf, ctx, msgServer := setupForBurnTest(mintingDenom)

	ftf.SetMinters(ctx, types.Minters{Address: minter.Address})
	_, err := msgServer.Burn(sdk.WrapSDKContext(ctx), &types.MsgBurn{From: minter.Address, Amount: sdk.Coin{Denom: "notDenom"}})
	require.ErrorIs(t, err, types.ErrBurn)
	require.ErrorContains(t, err, "burning denom is incorrect")
}

func TestBurn_Paused(t *testing.T) {
	mintingDenom := "uusdc"
	minter := sample.TestAccount()
	amount := sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(1)}
	ftf, ctx, msgServer := setupForBurnTest(mintingDenom)

	ftf.SetMinters(ctx, types.Minters{Address: minter.Address})
	ftf.SetPaused(ctx, types.Paused{Paused: true})
	_, err := msgServer.Burn(sdk.WrapSDKContext(ctx), &types.MsgBurn{From: minter.Address, Amount: amount})
	require.ErrorIs(t, err, types.ErrBurn)
	require.ErrorContains(t, err, "burning is paused")
}

func TestBurn_NilAmount(t *testing.T) {
	mintingDenom := "uusdc"
	minter := sample.TestAccount()
	ftf, ctx, msgServer := setupForBurnTest(mintingDenom)

	ftf.SetMinters(ctx, types.Minters{Address: minter.Address})
	_, err := msgServer.Burn(sdk.WrapSDKContext(ctx), &types.MsgBurn{From: minter.Address, Amount: sdk.Coin{Denom: mintingDenom}})
	require.ErrorIs(t, err, types.ErrBurn)
	require.ErrorContains(t, err, "burning amount is invalid")
}

func TestBurn_NegativeAmount(t *testing.T) {
	mintingDenom := "uusdc"
	minter := sample.TestAccount()
	amount := sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(-1)}
	ftf, ctx, msgServer := setupForBurnTest(mintingDenom)

	ftf.SetMinters(ctx, types.Minters{Address: minter.Address})
	_, err := msgServer.Burn(sdk.WrapSDKContext(ctx), &types.MsgBurn{From: minter.Address, Amount: amount})
	require.ErrorIs(t, err, types.ErrBurn)
	require.ErrorContains(t, err, "burning amount is invalid")
}

func TestBurn_ZeroAmount(t *testing.T) {
	mintingDenom := "uusdc"
	minter := sample.TestAccount()
	amount := sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(0)}
	ftf, ctx, msgServer := setupForBurnTest(mintingDenom)

	ftf.SetMinters(ctx, types.Minters{Address: minter.Address})
	_, err := msgServer.Burn(sdk.WrapSDKContext(ctx), &types.MsgBurn{From: minter.Address, Amount: amount})
	require.ErrorIs(t, err, types.ErrBurn)
	require.ErrorContains(t, err, "burning amount is invalid")
}

func TestBurn_AmountExceedsBurnerBalance(t *testing.T) {
	mintingDenom := "uusdc"
	minter := sample.TestAccount()
	amount := sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(10)}
	ftf, ctx, msgServer := setupForBurnTest(mintingDenom)

	ftf.SetMinters(ctx, types.Minters{Address: minter.Address})
	_, err := msgServer.Burn(sdk.WrapSDKContext(ctx), &types.MsgBurn{From: minter.Address, Amount: amount})
	require.ErrorIs(t, err, types.ErrBurn)
	require.ErrorContains(t, err, "insufficient funds")
}

func TestBurn_Success(t *testing.T) {
	mintingDenom := "uusdc"
	minter := sample.TestAccount()
	amount := sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(10)}
	allowance := sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(10)}
	ftf, ctx, msgServer := setupForBurnTest(mintingDenom)

	ftf.SetMinters(ctx, types.Minters{Address: minter.Address, Allowance: allowance})

	// Mint once to ensure there are funds to burn
	mintRes, err := msgServer.Mint(sdk.WrapSDKContext(ctx), &types.MsgMint{From: minter.Address, Address: minter.Address, Amount: amount})
	require.NoError(t, err)
	require.Equal(t, &types.MsgMintResponse{}, mintRes)

	res, err := msgServer.Burn(sdk.WrapSDKContext(ctx), &types.MsgBurn{From: minter.Address, Amount: amount})
	require.NoError(t, err)
	require.Equal(t, &types.MsgBurnResponse{}, res)
}

func setupForBurnTest(mintingDenom string) (*keeper.Keeper, sdk.Context, types.MsgServer) {
	ftf, ctx := testkeeper.FiatTokenfactoryKeeper()
	ftf.SetMintingDenom(ctx, types.MintingDenom{Denom: mintingDenom})
	ftf.SetPaused(ctx, types.Paused{Paused: false})
	msgServer := keeper.NewMsgServerImpl(ftf)
	return ftf, ctx, msgServer
}
