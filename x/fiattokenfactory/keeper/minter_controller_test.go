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

func createNMinterController(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.MinterController {
	items := make([]types.MinterController, n)
	for i := range items {
		items[i].Controller = strconv.Itoa(i)

		keeper.SetMinterController(ctx, items[i])
	}
	return items
}

func TestMinterControllerGetAndSet_Success(t *testing.T) {
	controller := sample.TestAccount()
	minter := sample.TestAccount()
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	keeper.SetMinterController(ctx, types.MinterController{Controller: controller.Address, Minter: minter.Address})

	assertAccountIsAController(t, *keeper, ctx, controller, minter)
}

func TestMinterControllerGet_AddressIsNotController(t *testing.T) {
	acc := sample.TestAccount()
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	assertAccountIsNotAController(t, *keeper, ctx, acc)
}

func TestMinterControllerSet_UpdateControlledMinter(t *testing.T) {
	controller := sample.TestAccount()
	minter := sample.TestAccount()
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()
	keeper.SetMinterController(ctx, types.MinterController{Controller: controller.Address, Minter: minter.Address})
	assertAccountIsAController(t, *keeper, ctx, controller, minter)

	newMinter := sample.TestAccount()
	keeper.SetMinterController(ctx, types.MinterController{Controller: controller.Address, Minter: newMinter.Address})

	assertAccountIsAController(t, *keeper, ctx, controller, newMinter)
}

func TestMinterControllerDelete_AddressIsNotController(t *testing.T) {
	acc := sample.TestAccount()
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	keeper.DeleteMinterController(ctx, acc.Address)

	assertAccountIsNotAController(t, *keeper, ctx, acc)
}

func TestMinterControllerDelete_AddressIsController(t *testing.T) {
	controller := sample.TestAccount()
	minter := sample.TestAccount()
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()
	keeper.SetMinterController(ctx, types.MinterController{Controller: controller.Address, Minter: minter.Address})
	assertAccountIsAController(t, *keeper, ctx, controller, minter)

	keeper.DeleteMinterController(ctx, controller.Address)

	assertAccountIsNotAController(t, *keeper, ctx, controller)
}

func TestMinterControllerGetAll_EmptyControllerList(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	controllers := keeper.GetAllMinterControllers(ctx)
	require.Empty(t, controllers)
}

func TestMinterControllerGetAll_NonEmptyControllerList(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()
	items := createNMinterController(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllMinterControllers(ctx)),
	)
}

func assertAccountIsAController(t *testing.T, keeper keeper.Keeper, ctx sdk.Context, controller sample.Account, minter sample.Account) {
	rst, found := keeper.GetMinterController(ctx, controller.Address)
	require.True(t, found)
	require.Equal(t,
		nullify.Fill(&types.MinterController{Controller: controller.Address, Minter: minter.Address}),
		nullify.Fill(&rst),
	)
}

func assertAccountIsNotAController(t *testing.T, keeper keeper.Keeper, ctx sdk.Context, controller sample.Account) {
	_, found := keeper.GetMinterController(ctx, controller.Address)
	require.False(t, found)
}
