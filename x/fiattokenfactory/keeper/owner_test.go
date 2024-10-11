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

func TestOwnerGetAndSet_Success(t *testing.T) {
	owner := sample.TestAccount()
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	keeper.SetOwner(ctx, types.Owner{Address: owner.Address})

	assertOwner(t, *keeper, ctx, owner)
}

func TestOwnerSet_Overwrite(t *testing.T) {
	owner := sample.TestAccount()
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()
	keeper.SetOwner(ctx, types.Owner{Address: owner.Address})
	assertOwner(t, *keeper, ctx, owner)

	newOwner := sample.TestAccount()
	keeper.SetOwner(ctx, types.Owner{Address: newOwner.Address})
	assertOwner(t, *keeper, ctx, newOwner)
}

func TestOwnerGet_Unset(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	_, found := keeper.GetOwner(ctx)
	require.False(t, found)
}

func TestPendingOwnerGetAndSet_Success(t *testing.T) {
	pendingOwner := sample.TestAccount()
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	keeper.SetPendingOwner(ctx, types.Owner{Address: pendingOwner.Address})

	assertPendingOwner(t, *keeper, ctx, pendingOwner)
}

func TestPendingOwnerSet_Overwrite(t *testing.T) {
	pendingOwner := sample.TestAccount()
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()
	keeper.SetPendingOwner(ctx, types.Owner{Address: pendingOwner.Address})
	assertPendingOwner(t, *keeper, ctx, pendingOwner)

	newOwner := sample.TestAccount()
	keeper.SetPendingOwner(ctx, types.Owner{Address: newOwner.Address})
	assertPendingOwner(t, *keeper, ctx, newOwner)
}

func TestPendingOwnerGet_Unset(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	_, found := keeper.GetPendingOwner(ctx)
	require.False(t, found)
}

func TestPendingOwnerDelete_Unset(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	keeper.DeletePendingOwner(ctx)

	_, found := keeper.GetPendingOwner(ctx)
	require.False(t, found)
}

func TestPendingOwnerDelete_Success(t *testing.T) {
	pendingOwner := sample.TestAccount()
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()
	keeper.SetPendingOwner(ctx, types.Owner{Address: pendingOwner.Address})

	keeper.DeletePendingOwner(ctx)

	_, found := keeper.GetPendingOwner(ctx)
	require.False(t, found)
}

func assertOwner(t *testing.T, keeper keeper.Keeper, ctx sdk.Context, expectedOwner sample.Account) {
	rst, found := keeper.GetOwner(ctx)
	require.True(t, found)
	require.Equal(t,
		nullify.Fill(&types.Owner{Address: expectedOwner.Address}),
		nullify.Fill(&rst),
	)
}

func assertPendingOwner(t *testing.T, keeper keeper.Keeper, ctx sdk.Context, expectedPendingOwner sample.Account) {
	rst, found := keeper.GetPendingOwner(ctx)
	require.True(t, found)
	require.Equal(t,
		nullify.Fill(&types.Owner{Address: expectedPendingOwner.Address}),
		nullify.Fill(&rst),
	)
}
