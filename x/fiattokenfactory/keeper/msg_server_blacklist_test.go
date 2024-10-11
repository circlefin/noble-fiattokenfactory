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

func TestBlacklist_BlacklisterNotSet(t *testing.T) {
	ftf, ctx := testkeeper.FiatTokenfactoryKeeper()
	msgServer := keeper.NewMsgServerImpl(ftf)

	_, err := msgServer.Blacklist(sdk.WrapSDKContext(ctx), &types.MsgBlacklist{From: sample.TestAccount().Address})
	require.ErrorIs(t, err, types.ErrUserNotFound)
	require.ErrorContains(t, err, "blacklister is not set")
}

func TestBlacklist_FromAddressIsNotBlacklister(t *testing.T) {
	blacklister := sample.TestAccount()
	ftf, ctx := testkeeper.FiatTokenfactoryKeeper()
	msgServer := keeper.NewMsgServerImpl(ftf)
	ftf.SetBlacklister(ctx, types.Blacklister{Address: blacklister.Address})

	_, err := msgServer.Blacklist(sdk.WrapSDKContext(ctx), &types.MsgBlacklist{From: sample.AccAddress()})
	require.ErrorIs(t, err, types.ErrUnauthorized)
	require.ErrorContains(t, err, "you are not the blacklister")
}

func TestBlacklist_AddressIsInvalid(t *testing.T) {
	blacklister := sample.TestAccount()
	ftf, ctx := testkeeper.FiatTokenfactoryKeeper()
	msgServer := keeper.NewMsgServerImpl(ftf)
	ftf.SetBlacklister(ctx, types.Blacklister{Address: blacklister.Address})

	_, err := msgServer.Blacklist(sdk.WrapSDKContext(ctx), &types.MsgBlacklist{From: blacklister.Address, Address: "invalid address"})
	require.ErrorIs(t, err, bech32.ErrInvalidCharacter(32))
}

func TestBlacklist_AddressAlreadyBlacklisted(t *testing.T) {
	blacklister := sample.TestAccount()
	blacklistedUserBech32 := sample.TestAccount()
	blacklistedUserBech32m := sample.TestAccountBech32m()

	ftf, ctx := testkeeper.FiatTokenfactoryKeeper()
	msgServer := keeper.NewMsgServerImpl(ftf)
	ftf.SetBlacklister(ctx, types.Blacklister{Address: blacklister.Address})
	ftf.SetBlacklisted(ctx, types.Blacklisted{AddressBz: blacklistedUserBech32.AddressBz})
	ftf.SetBlacklisted(ctx, types.Blacklisted{AddressBz: blacklistedUserBech32m.AddressBz})

	_, err := msgServer.Blacklist(sdk.WrapSDKContext(ctx), &types.MsgBlacklist{From: blacklister.Address, Address: blacklistedUserBech32.Address})
	require.ErrorIs(t, err, types.ErrUserBlacklisted)
	_, err = msgServer.Blacklist(sdk.WrapSDKContext(ctx), &types.MsgBlacklist{From: blacklister.Address, Address: blacklistedUserBech32m.Address})
	require.ErrorIs(t, err, types.ErrUserBlacklisted)
}

func TestBlacklist_Success(t *testing.T) {
	blacklister := sample.TestAccount()
	newUserBech32 := sample.TestAccount()
	newUserBech32m := sample.TestAccountBech32m()

	ftf, ctx := testkeeper.FiatTokenfactoryKeeper()
	msgServer := keeper.NewMsgServerImpl(ftf)
	ftf.SetBlacklister(ctx, types.Blacklister{Address: blacklister.Address})

	res, err := msgServer.Blacklist(sdk.WrapSDKContext(ctx), &types.MsgBlacklist{From: blacklister.Address, Address: newUserBech32.Address})
	require.NoError(t, err)
	require.Equal(t, &types.MsgBlacklistResponse{}, res)

	res, err = msgServer.Blacklist(sdk.WrapSDKContext(ctx), &types.MsgBlacklist{From: blacklister.Address, Address: newUserBech32m.Address})
	require.NoError(t, err)
	require.Equal(t, &types.MsgBlacklistResponse{}, res)
}
