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

	"github.com/btcsuite/btcd/btcutil/bech32"
	testkeeper "github.com/circlefin/noble-fiattokenfactory/testutil/keeper"
	"github.com/circlefin/noble-fiattokenfactory/testutil/sample"
	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/keeper"
	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestUnblacklist_BlacklisterNotSet(t *testing.T) {
	ftf, ctx := testkeeper.FiatTokenfactoryKeeper()
	msgServer := keeper.NewMsgServerImpl(ftf)

	_, err := msgServer.Unblacklist(sdk.WrapSDKContext(ctx), &types.MsgUnblacklist{From: sample.AccAddress()})
	require.ErrorIs(t, err, types.ErrUserNotFound)
	require.ErrorContains(t, err, "blacklister is not set")
}

func TestUnblacklist_FromAddressIsNotBlacklister(t *testing.T) {
	blacklister := sample.TestAccount()
	ftf, ctx := testkeeper.FiatTokenfactoryKeeper()
	msgServer := keeper.NewMsgServerImpl(ftf)
	ftf.SetBlacklister(ctx, types.Blacklister{Address: blacklister.Address})

	_, err := msgServer.Unblacklist(sdk.WrapSDKContext(ctx), &types.MsgUnblacklist{From: sample.AccAddress()})
	require.ErrorIs(t, err, types.ErrUnauthorized)
}

func TestUnblacklist_AddressIsInvalid(t *testing.T) {
	blacklister := sample.TestAccount()
	ftf, ctx := testkeeper.FiatTokenfactoryKeeper()
	msgServer := keeper.NewMsgServerImpl(ftf)
	ftf.SetBlacklister(ctx, types.Blacklister{Address: blacklister.Address})

	_, err := msgServer.Unblacklist(sdk.WrapSDKContext(ctx), &types.MsgUnblacklist{From: blacklister.Address, Address: "invalid address"})
	require.ErrorIs(t, err, bech32.ErrInvalidCharacter(32))
}

func TestUnblacklist_AddressIsNotBlacklisted(t *testing.T) {
	blacklister := sample.TestAccount()
	unblacklistedBech32User := sample.TestAccount()
	unblacklistedBech32mUser := sample.TestAccountBech32m()

	ftf, ctx := testkeeper.FiatTokenfactoryKeeper()
	msgServer := keeper.NewMsgServerImpl(ftf)
	ftf.SetBlacklister(ctx, types.Blacklister{Address: blacklister.Address})

	_, err := msgServer.Unblacklist(sdk.WrapSDKContext(ctx), &types.MsgUnblacklist{From: blacklister.Address, Address: unblacklistedBech32User.Address})
	require.ErrorIs(t, types.ErrUserNotFound, err)
	require.ErrorContains(t, err, "the specified address is not blacklisted")

	_, err = msgServer.Unblacklist(sdk.WrapSDKContext(ctx), &types.MsgUnblacklist{From: blacklister.Address, Address: unblacklistedBech32mUser.Address})
	require.ErrorIs(t, types.ErrUserNotFound, err)
	require.ErrorContains(t, err, "the specified address is not blacklisted")
}

func TestUnblacklist_Success(t *testing.T) {
	blacklister := sample.TestAccount()
	blacklistedBech32User := sample.TestAccount()
	blacklistedBech32mUser := sample.TestAccountBech32m()

	ftf, ctx := testkeeper.FiatTokenfactoryKeeper()
	msgServer := keeper.NewMsgServerImpl(ftf)
	ftf.SetBlacklister(ctx, types.Blacklister{Address: blacklister.Address})
	ftf.SetBlacklisted(ctx, types.Blacklisted{AddressBz: blacklistedBech32User.AddressBz})
	ftf.SetBlacklisted(ctx, types.Blacklisted{AddressBz: blacklistedBech32mUser.AddressBz})

	res, err := msgServer.Unblacklist(sdk.WrapSDKContext(ctx), &types.MsgUnblacklist{From: blacklister.Address, Address: blacklistedBech32User.Address})
	require.NoError(t, err)
	require.Equal(t, &types.MsgUnblacklistResponse{}, res)

	res, err = msgServer.Unblacklist(sdk.WrapSDKContext(ctx), &types.MsgUnblacklist{From: blacklister.Address, Address: blacklistedBech32mUser.Address})
	require.NoError(t, err)
	require.Equal(t, &types.MsgUnblacklistResponse{}, res)
}
