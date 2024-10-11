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

func TestMint_FromAddressIsNotMinter(t *testing.T) {
	var (
		minter       = sample.TestAccount()
		mintingDenom = "uusdc"
		allowance    = sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(10)}
	)
	_, ctx, msgServer := setupForMintTest(mintingDenom, minter, allowance)

	_, err := msgServer.Mint(sdk.WrapSDKContext(ctx), &types.MsgMint{From: sample.AccAddress()})
	require.ErrorIs(t, err, types.ErrUnauthorized)
	require.ErrorContains(t, err, "you are not a minter")
}

func TestMint_InvalidMinterAddress(t *testing.T) {
	var (
		minter       = sample.TestAccount()
		mintingDenom = "uusdc"
		allowance    = sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(10)}
	)
	ftf, ctx, msgServer := setupForMintTest(mintingDenom, minter, allowance)

	ftf.SetMinters(ctx, types.Minters{Address: "invalid address"})
	_, err := msgServer.Mint(sdk.WrapSDKContext(ctx), &types.MsgMint{From: "invalid address"})
	require.Error(t, err)
}

func TestMint_BlacklistedMinterAddress(t *testing.T) {
	var (
		minter       = sample.TestAccount()
		mintingDenom = "uusdc"
		allowance    = sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(10)}
	)
	ftf, ctx, msgServer := setupForMintTest(mintingDenom, minter, allowance)

	ftf.SetBlacklisted(ctx, types.Blacklisted{AddressBz: minter.AddressBz})
	_, err := msgServer.Mint(sdk.WrapSDKContext(ctx), &types.MsgMint{From: minter.Address})
	require.ErrorIs(t, err, types.ErrMint)
	require.ErrorContains(t, err, "minter address is blacklisted")
}

func TestMint_BlacklistedBech32mMinterAddress(t *testing.T) {
	var (
		minter       = sample.TestAccountBech32m()
		mintingDenom = "uusdc"
		allowance    = sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(10)}
	)
	ftf, ctx, msgServer := setupForMintTest(mintingDenom, minter, allowance)

	ftf.SetBlacklisted(ctx, types.Blacklisted{AddressBz: minter.AddressBz})
	_, err := msgServer.Mint(sdk.WrapSDKContext(ctx), &types.MsgMint{From: minter.Address})
	require.ErrorIs(t, err, types.ErrMint)
	require.ErrorContains(t, err, "minter address is blacklisted")
}

func TestMint_InvalidReceiverAddress(t *testing.T) {
	var (
		minter       = sample.TestAccount()
		mintingDenom = "uusdc"
		allowance    = sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(10)}
	)
	_, ctx, msgServer := setupForMintTest(mintingDenom, minter, allowance)

	_, err := msgServer.Mint(sdk.WrapSDKContext(ctx), &types.MsgMint{From: minter.Address, Address: "invalid address"})
	require.Error(t, err)
}

func TestMint_BlacklistedReceiverAddress(t *testing.T) {
	var (
		receiver     = sample.TestAccount()
		minter       = sample.TestAccount()
		mintingDenom = "uusdc"
		allowance    = sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(10)}
	)
	ftf, ctx, msgServer := setupForMintTest(mintingDenom, minter, allowance)

	ftf.SetBlacklisted(ctx, types.Blacklisted{AddressBz: receiver.AddressBz})
	_, err := msgServer.Mint(sdk.WrapSDKContext(ctx), &types.MsgMint{From: minter.Address, Address: receiver.Address})
	require.ErrorIs(t, err, types.ErrMint)
	require.ErrorContains(t, err, "receiver address is blacklisted")
}

func TestMint_BlacklistedBech32mReceiverAddress(t *testing.T) {
	var (
		receiver     = sample.TestAccountBech32m()
		minter       = sample.TestAccount()
		mintingDenom = "uusdc"
		allowance    = sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(10)}
	)
	ftf, ctx, msgServer := setupForMintTest(mintingDenom, minter, allowance)

	ftf.SetBlacklisted(ctx, types.Blacklisted{AddressBz: receiver.AddressBz})
	_, err := msgServer.Mint(sdk.WrapSDKContext(ctx), &types.MsgMint{From: minter.Address, Address: receiver.Address})
	require.ErrorIs(t, err, types.ErrMint)
	require.ErrorContains(t, err, "receiver address is blacklisted")
}

func TestMint_DenomIsMissing(t *testing.T) {
	var (
		receiver     = sample.TestAccount()
		minter       = sample.TestAccount()
		mintingDenom = "uusdc"
		allowance    = sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(10)}
	)
	_, ctx, msgServer := setupForMintTest(mintingDenom, minter, allowance)

	_, err := msgServer.Mint(sdk.WrapSDKContext(ctx), &types.MsgMint{
		From:    minter.Address,
		Address: receiver.Address,
		Amount:  sdk.Coin{},
	})
	require.ErrorIs(t, err, types.ErrMint)
	require.ErrorContains(t, err, "minting denom is incorrect")
}

func TestMint_DenomIsEmpty(t *testing.T) {
	var (
		receiver     = sample.TestAccount()
		minter       = sample.TestAccount()
		mintingDenom = "uusdc"
		allowance    = sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(10)}
	)
	_, ctx, msgServer := setupForMintTest(mintingDenom, minter, allowance)

	_, err := msgServer.Mint(sdk.WrapSDKContext(ctx), &types.MsgMint{
		From:    minter.Address,
		Address: receiver.Address,
		Amount:  sdk.Coin{Denom: ""},
	})
	require.ErrorIs(t, err, types.ErrMint)
	require.ErrorContains(t, err, "minting denom is incorrect")
}

func TestMint_IncorrectDenom(t *testing.T) {
	var (
		receiver     = sample.TestAccount()
		minter       = sample.TestAccount()
		mintingDenom = "uusdc"
		allowance    = sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(10)}
	)
	_, ctx, msgServer := setupForMintTest(mintingDenom, minter, allowance)

	_, err := msgServer.Mint(sdk.WrapSDKContext(ctx), &types.MsgMint{
		From:    minter.Address,
		Address: receiver.Address,
		Amount:  sdk.Coin{Denom: "notMintingDenom"},
	})
	require.ErrorIs(t, err, types.ErrMint)
	require.ErrorContains(t, err, "minting denom is incorrect")
}

func TestMint_AmountExceedsAllowance(t *testing.T) {
	var (
		receiver     = sample.TestAccount()
		minter       = sample.TestAccount()
		mintingDenom = "uusdc"
		allowance    = sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(10)}
		amount       = sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(20)}
	)
	_, ctx, msgServer := setupForMintTest(mintingDenom, minter, allowance)

	_, err := msgServer.Mint(sdk.WrapSDKContext(ctx), &types.MsgMint{From: minter.Address, Address: receiver.Address, Amount: amount})
	require.ErrorIs(t, err, types.ErrMint)
	require.ErrorContains(t, err, "minting amount is greater than the allowance")
}

func TestMint_NilAmount(t *testing.T) {
	var (
		receiver     = sample.TestAccount()
		minter       = sample.TestAccount()
		mintingDenom = "uusdc"
		allowance    = sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(10)}
		amount       = sdk.Coin{Denom: mintingDenom}
	)
	_, ctx, msgServer := setupForMintTest(mintingDenom, minter, allowance)

	_, err := msgServer.Mint(sdk.WrapSDKContext(ctx), &types.MsgMint{From: minter.Address, Address: receiver.Address, Amount: amount})
	require.ErrorIs(t, err, types.ErrMint)
	require.ErrorContains(t, err, "minting amount is invalid")
}

func TestMint_NegativeAmount(t *testing.T) {
	var (
		receiver     = sample.TestAccount()
		minter       = sample.TestAccount()
		mintingDenom = "uusdc"
		allowance    = sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(10)}
		amount       = sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(-1)}
	)
	_, ctx, msgServer := setupForMintTest(mintingDenom, minter, allowance)

	_, err := msgServer.Mint(sdk.WrapSDKContext(ctx), &types.MsgMint{From: minter.Address, Address: receiver.Address, Amount: amount})
	require.ErrorIs(t, err, types.ErrMint)
	require.ErrorContains(t, err, "minting amount is invalid")
}

func TestMint_ZeroAmount(t *testing.T) {
	var (
		receiver     = sample.TestAccount()
		minter       = sample.TestAccount()
		mintingDenom = "uusdc"
		allowance    = sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(10)}
		amount       = sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(0)}
	)
	_, ctx, msgServer := setupForMintTest(mintingDenom, minter, allowance)

	_, err := msgServer.Mint(sdk.WrapSDKContext(ctx), &types.MsgMint{From: minter.Address, Address: receiver.Address, Amount: amount})
	require.ErrorIs(t, err, types.ErrMint)
	require.ErrorContains(t, err, "minting amount is invalid")
}

func TestMint_Paused(t *testing.T) {
	var (
		receiver     = sample.TestAccount()
		minter       = sample.TestAccount()
		mintingDenom = "uusdc"
		allowance    = sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(10)}
		amount       = sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(1)}
	)
	ftf, ctx, msgServer := setupForMintTest(mintingDenom, minter, allowance)

	ftf.SetPaused(ctx, types.Paused{Paused: true})
	_, err := msgServer.Mint(sdk.WrapSDKContext(ctx), &types.MsgMint{From: minter.Address, Address: receiver.Address, Amount: amount})
	require.ErrorIs(t, err, types.ErrMint)
	require.ErrorContains(t, err, "minting is paused")
}

func TestMint_AmountEqualsAllowance(t *testing.T) {
	var (
		receiver     = sample.TestAccount()
		minter       = sample.TestAccount()
		mintingDenom = "uusdc"
		allowance    = sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(10)}
		amount       = sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(10)}
	)
	_, ctx, msgServer := setupForMintTest(mintingDenom, minter, allowance)

	res, err := msgServer.Mint(sdk.WrapSDKContext(ctx), &types.MsgMint{From: minter.Address, Address: receiver.Address, Amount: amount})
	require.NoError(t, err)
	require.Equal(t, &types.MsgMintResponse{}, res)
}

func TestMint_Success(t *testing.T) {
	var (
		receiver     = sample.TestAccount()
		minter       = sample.TestAccount()
		mintingDenom = "uusdc"
		allowance    = sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(10)}
		amount       = sdk.Coin{Denom: mintingDenom, Amount: math.NewInt(1)}
	)
	_, ctx, msgServer := setupForMintTest(mintingDenom, minter, allowance)

	res, err := msgServer.Mint(sdk.WrapSDKContext(ctx), &types.MsgMint{From: minter.Address, Address: receiver.Address, Amount: amount})
	require.NoError(t, err)
	require.Equal(t, &types.MsgMintResponse{}, res)
}

func setupForMintTest(mintingDenom string, minter sample.Account, allowance sdk.Coin) (*keeper.Keeper, sdk.Context, types.MsgServer) {
	ftf, ctx := testkeeper.FiatTokenfactoryKeeper()
	ftf.SetMintingDenom(ctx, types.MintingDenom{Denom: mintingDenom})
	ftf.SetPaused(ctx, types.Paused{Paused: false})
	ftf.SetMinters(ctx, types.Minters{Address: minter.Address, Allowance: allowance})
	msgServer := keeper.NewMsgServerImpl(ftf)
	return ftf, ctx, msgServer
}
