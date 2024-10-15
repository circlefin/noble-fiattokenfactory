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
	"strconv"
	"testing"

	keepertest "github.com/circlefin/noble-fiattokenfactory/testutil/keeper"
	"github.com/circlefin/noble-fiattokenfactory/testutil/nullify"
	"github.com/circlefin/noble-fiattokenfactory/testutil/sample"
	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/keeper"
	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// Prevent strconv unused error
var _ = strconv.IntSize

type blacklistedWrapper struct {
	address string
	bl      types.Blacklisted
}

func createNBlacklisted(keeper *keeper.Keeper, ctx sdk.Context, n int) []blacklistedWrapper {
	items := make([]blacklistedWrapper, n)
	for i := range items {
		acc := sample.TestAccount()
		items[i].address = acc.Address
		items[i].bl.AddressBz = acc.AddressBz

		keeper.SetBlacklisted(ctx, items[i].bl)
	}
	return items
}

func createNBlacklistedBech32m(keeper *keeper.Keeper, ctx sdk.Context, n int) []blacklistedWrapper {
	items := make([]blacklistedWrapper, n)
	for i := range items {
		acc := sample.TestAccountBech32m()
		items[i].address = acc.Address
		items[i].bl.AddressBz = acc.AddressBz

		keeper.SetBlacklisted(ctx, items[i].bl)
	}
	return items
}

func TestBlacklistedGetAndSet_Success(t *testing.T) {
	accBech32, accBech32M := sample.TestAccount(), sample.TestAccountBech32m()
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	keeper.SetBlacklisted(ctx, types.Blacklisted{AddressBz: accBech32.AddressBz})
	keeper.SetBlacklisted(ctx, types.Blacklisted{AddressBz: accBech32M.AddressBz})

	assertAccountIsBlacklisted(t, *keeper, ctx, accBech32)
	assertAccountIsBlacklisted(t, *keeper, ctx, accBech32M)
}

func TestBlacklistedGet_AccountNotBlacklisted(t *testing.T) {
	accBech32, accBech32M := sample.TestAccount(), sample.TestAccountBech32m()
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	assertAccountIsNotBlacklisted(t, *keeper, ctx, accBech32)
	assertAccountIsNotBlacklisted(t, *keeper, ctx, accBech32M)
}

func TestBlacklistedSet_AlreadyBlacklisted(t *testing.T) {
	accBech32, accBech32M := sample.TestAccount(), sample.TestAccountBech32m()
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	keeper.SetBlacklisted(ctx, types.Blacklisted{AddressBz: accBech32.AddressBz})
	keeper.SetBlacklisted(ctx, types.Blacklisted{AddressBz: accBech32M.AddressBz})
	assertAccountIsBlacklisted(t, *keeper, ctx, accBech32)
	assertAccountIsBlacklisted(t, *keeper, ctx, accBech32M)

	keeper.SetBlacklisted(ctx, types.Blacklisted{AddressBz: accBech32.AddressBz})
	keeper.SetBlacklisted(ctx, types.Blacklisted{AddressBz: accBech32M.AddressBz})

	assertAccountIsBlacklisted(t, *keeper, ctx, accBech32)
	assertAccountIsBlacklisted(t, *keeper, ctx, accBech32M)
}

func TestBlacklistSet_Multiple(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()
	items := createNBlacklisted(keeper, ctx, 10)
	items = append(items, createNBlacklistedBech32m(keeper, ctx, 10)...)
	for _, item := range items {
		assertAddressIsBlacklisted(t, *keeper, ctx, item.bl)
	}
}

func TestBlacklistedRemove_AccountNotBlacklisted(t *testing.T) {
	accBech32, accBech32M := sample.TestAccount(), sample.TestAccountBech32m()
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	keeper.RemoveBlacklisted(ctx, accBech32.AddressBz)
	keeper.RemoveBlacklisted(ctx, accBech32M.AddressBz)

	assertAccountIsNotBlacklisted(t, *keeper, ctx, accBech32)
	assertAccountIsNotBlacklisted(t, *keeper, ctx, accBech32M)
}

func TestBlacklistedRemove_Success(t *testing.T) {
	accBech32, accBech32M := sample.TestAccount(), sample.TestAccountBech32m()
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	keeper.SetBlacklisted(ctx, types.Blacklisted{AddressBz: accBech32.AddressBz})
	keeper.SetBlacklisted(ctx, types.Blacklisted{AddressBz: accBech32M.AddressBz})

	assertAccountIsBlacklisted(t, *keeper, ctx, accBech32)
	assertAccountIsBlacklisted(t, *keeper, ctx, accBech32M)

	keeper.RemoveBlacklisted(ctx, accBech32.AddressBz)
	keeper.RemoveBlacklisted(ctx, accBech32M.AddressBz)

	assertAccountIsNotBlacklisted(t, *keeper, ctx, accBech32)
	assertAccountIsNotBlacklisted(t, *keeper, ctx, accBech32M)
}

func TestBlacklistedRemove_Multiple(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()
	items := createNBlacklisted(keeper, ctx, 10)
	items = append(items, createNBlacklistedBech32m(keeper, ctx, 10)...)
	for _, item := range items {
		assertAddressIsBlacklisted(t, *keeper, ctx, item.bl)

		keeper.RemoveBlacklisted(ctx, item.bl.AddressBz)

		assertAddressIsNotBlacklisted(t, *keeper, ctx, item.bl)
	}
}

func TestBlacklistedGetAll_EmptyBlacklist(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	blacklisted := keeper.GetAllBlacklisted(ctx)
	require.Empty(t, blacklisted)
}

func TestBlacklistedGetAll_NonEmptyBlacklist(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()
	items := createNBlacklisted(keeper, ctx, 10)
	items = append(items, createNBlacklistedBech32m(keeper, ctx, 10)...)
	blacklisted := make([]types.Blacklisted, len(items))
	for i, item := range items {
		blacklisted[i] = item.bl
	}
	require.ElementsMatch(t,
		nullify.Fill(blacklisted),
		nullify.Fill(keeper.GetAllBlacklisted(ctx)),
	)
}

func assertAccountIsBlacklisted(t *testing.T, keeper keeper.Keeper, ctx sdk.Context, acc sample.Account) {
	assertAddressIsBlacklisted(t, keeper, ctx, types.Blacklisted{AddressBz: acc.AddressBz})
}

func assertAddressIsBlacklisted(t *testing.T, keeper keeper.Keeper, ctx sdk.Context, bl types.Blacklisted) {
	rst, found := keeper.GetBlacklisted(ctx, bl.AddressBz)
	require.True(t, found)
	require.Equal(t,
		nullify.Fill(&bl),
		nullify.Fill(&rst),
	)
}

func assertAccountIsNotBlacklisted(t *testing.T, keeper keeper.Keeper, ctx sdk.Context, acc sample.Account) {
	assertAddressIsNotBlacklisted(t, keeper, ctx, types.Blacklisted{AddressBz: acc.AddressBz})
}

func assertAddressIsNotBlacklisted(t *testing.T, keeper keeper.Keeper, ctx sdk.Context, bl types.Blacklisted) {
	_, found := keeper.GetBlacklisted(ctx, bl.AddressBz)
	require.False(t, found)
}
