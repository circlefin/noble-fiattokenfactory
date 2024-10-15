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

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/circlefin/noble-fiattokenfactory/testutil/keeper"
	"github.com/circlefin/noble-fiattokenfactory/testutil/nullify"
	"github.com/circlefin/noble-fiattokenfactory/testutil/sample"
	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/keeper"
	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
)

func createTestBlacklister(keeper *keeper.Keeper, ctx sdk.Context) types.Blacklister {
	item := types.Blacklister{}
	keeper.SetBlacklister(ctx, item)
	return item
}

func TestBlacklisterGetAndSet_Success(t *testing.T) {
	blacklister := sample.TestAccount()
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	keeper.SetBlacklister(ctx, types.Blacklister{Address: blacklister.Address})

	assertBlacklister(t, *keeper, ctx, blacklister)
}

func TestBlacklisterSet_Overwrite(t *testing.T) {
	blacklister := sample.TestAccount()
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()
	keeper.SetBlacklister(ctx, types.Blacklister{Address: blacklister.Address})
	assertBlacklister(t, *keeper, ctx, blacklister)

	newBlacklister := sample.TestAccount()
	keeper.SetBlacklister(ctx, types.Blacklister{Address: newBlacklister.Address})
	assertBlacklister(t, *keeper, ctx, newBlacklister)
}

func TestBlacklisterGet_Unset(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	_, found := keeper.GetBlacklister(ctx)
	require.False(t, found)
}

func assertBlacklister(t *testing.T, keeper keeper.Keeper, ctx sdk.Context, expectedBlacklister sample.Account) {
	rst, found := keeper.GetBlacklister(ctx)
	require.True(t, found)
	require.Equal(t,
		nullify.Fill(&types.Blacklister{Address: expectedBlacklister.Address}),
		nullify.Fill(&rst),
	)
}
