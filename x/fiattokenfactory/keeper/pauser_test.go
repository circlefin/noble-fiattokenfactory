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

func createTestPauser(keeper *keeper.Keeper, ctx sdk.Context) types.Pauser {
	item := types.Pauser{}
	keeper.SetPauser(ctx, item)
	return item
}

func TestPauserGetAndSet_Success(t *testing.T) {
	pauser := sample.TestAccount()
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	keeper.SetPauser(ctx, types.Pauser{Address: pauser.Address})

	assertPauser(t, *keeper, ctx, pauser)
}

func TestPauserSet_Overwrite(t *testing.T) {
	pauser := sample.TestAccount()
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()
	keeper.SetPauser(ctx, types.Pauser{Address: pauser.Address})
	assertPauser(t, *keeper, ctx, pauser)

	newPauser := sample.TestAccount()
	keeper.SetPauser(ctx, types.Pauser{Address: newPauser.Address})
	assertPauser(t, *keeper, ctx, newPauser)
}

func TestPauserGet_Unset(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	_, found := keeper.GetPauser(ctx)
	require.False(t, found)
}

func assertPauser(t *testing.T, keeper keeper.Keeper, ctx sdk.Context, expectedPauser sample.Account) {
	rst, found := keeper.GetPauser(ctx)
	require.True(t, found)
	require.Equal(t,
		nullify.Fill(&types.Pauser{Address: expectedPauser.Address}),
		nullify.Fill(&rst),
	)
}
