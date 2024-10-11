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

	"cosmossdk.io/math"
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

func createNMinters(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.Minters {
	items := make([]types.Minters, n)
	for i := range items {
		items[i].Address = strconv.Itoa(i)

		keeper.SetMinters(ctx, items[i])
	}
	return items
}

func TestMintersGetAndSet_Success(t *testing.T) {
	minter := sample.TestAccount()
	allowance := sdk.Coin{Denom: "uusdc", Amount: math.NewInt(1)}
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	keeper.SetMinters(ctx, types.Minters{Address: minter.Address, Allowance: allowance})

	assertAccountIsAMinter(t, *keeper, ctx, minter, allowance)
}

func TestMintersGet_AddressIsNotMinter(t *testing.T) {
	acc := sample.TestAccount()
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	assertAccountIsNotAMinter(t, *keeper, ctx, acc)
}

func TestMintersSet_UpdateMinterAllowance(t *testing.T) {
	minter := sample.TestAccount()
	allowance := sdk.Coin{Denom: "uusdc", Amount: math.NewInt(1)}
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()
	keeper.SetMinters(ctx, types.Minters{Address: minter.Address, Allowance: allowance})
	assertAccountIsAMinter(t, *keeper, ctx, minter, allowance)

	newAllowance := sdk.Coin{Denom: "uusdc", Amount: math.NewInt(5)}
	keeper.SetMinters(ctx, types.Minters{Address: minter.Address, Allowance: newAllowance})

	assertAccountIsAMinter(t, *keeper, ctx, minter, newAllowance)
}

func TestMintersRemove_AddressIsNotMinter(t *testing.T) {
	acc := sample.TestAccount()
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	keeper.RemoveMinters(ctx, acc.Address)

	assertAccountIsNotAMinter(t, *keeper, ctx, acc)
}

func TestMintersRemove_AddressIsMinter(t *testing.T) {
	minter := sample.TestAccount()
	allowance := sdk.Coin{Denom: "uusdc", Amount: math.NewInt(1)}
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()
	keeper.SetMinters(ctx, types.Minters{Address: minter.Address, Allowance: allowance})
	assertAccountIsAMinter(t, *keeper, ctx, minter, allowance)

	keeper.RemoveMinters(ctx, minter.Address)

	assertAccountIsNotAMinter(t, *keeper, ctx, minter)
}

func TestMintersGetAll_EmptyMinterList(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	controllers := keeper.GetAllMinters(ctx)
	require.Empty(t, controllers)
}

func TestMintersGetAll_NonEmptyMinterList(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()
	items := createNMinters(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllMinters(ctx)),
	)
}

func assertAccountIsAMinter(t *testing.T, keeper keeper.Keeper, ctx sdk.Context, minter sample.Account, allowance sdk.Coin) {
	rst, found := keeper.GetMinters(ctx, minter.Address)
	require.True(t, found)
	require.Equal(t,
		nullify.Fill(&types.Minters{Address: minter.Address, Allowance: allowance}),
		nullify.Fill(&rst),
	)
}

func assertAccountIsNotAMinter(t *testing.T, keeper keeper.Keeper, ctx sdk.Context, minter sample.Account) {
	_, found := keeper.GetMinters(ctx, minter.Address)
	require.False(t, found)
}
