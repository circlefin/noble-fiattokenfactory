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

func createTestMasterMinter(keeper *keeper.Keeper, ctx sdk.Context) types.MasterMinter {
	item := types.MasterMinter{}
	keeper.SetMasterMinter(ctx, item)
	return item
}

func TestMasterMinterGetAndSet_Success(t *testing.T) {
	masterMinter := sample.TestAccount()
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	keeper.SetMasterMinter(ctx, types.MasterMinter{Address: masterMinter.Address})

	assertMasterMinter(t, *keeper, ctx, masterMinter)
}

func TestMasterMinterSet_Overwrite(t *testing.T) {
	masterMinter := sample.TestAccount()
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()
	keeper.SetMasterMinter(ctx, types.MasterMinter{Address: masterMinter.Address})
	assertMasterMinter(t, *keeper, ctx, masterMinter)

	newMasterMinter := sample.TestAccount()
	keeper.SetMasterMinter(ctx, types.MasterMinter{Address: newMasterMinter.Address})
	assertMasterMinter(t, *keeper, ctx, newMasterMinter)
}

func TestMasterMinterGet_Unset(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	_, found := keeper.GetMasterMinter(ctx)
	require.False(t, found)
}

func assertMasterMinter(t *testing.T, keeper keeper.Keeper, ctx sdk.Context, expectedMasterMinter sample.Account) {
	rst, found := keeper.GetMasterMinter(ctx)
	require.True(t, found)
	require.Equal(t,
		nullify.Fill(&types.MasterMinter{Address: expectedMasterMinter.Address}),
		nullify.Fill(&rst),
	)
}
