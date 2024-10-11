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
	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/keeper"
	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
)

func createTestMintingDenom(keeper *keeper.Keeper, ctx sdk.Context) types.MintingDenom {
	item := types.MintingDenom{
		Denom: "uusdc",
	}
	keeper.SetMintingDenom(ctx, item)
	return item
}

func TestMintingDenomGetAndSet_Success(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()
	expectedDenom := createTestMintingDenom(keeper, ctx)

	denom := keeper.GetMintingDenom(ctx)
	require.Equal(t,
		nullify.Fill(&expectedDenom),
		nullify.Fill(&denom),
	)
}

func TestSetMintingDenom_AlreadySetPanics(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()
	createTestMintingDenom(keeper, ctx)

	require.Panics(t, func() { keeper.SetMintingDenom(ctx, types.MintingDenom{Denom: "uusdc"}) })
}

func TestSetMintingDenom_NoMetadataPanics(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	// in the mock bank keeper, it provides metadata only for uusdc
	require.Panics(t, func() { keeper.SetMintingDenom(ctx, types.MintingDenom{Denom: "notadenom"}) })
}

func TestMintingDenomGet_UnsetPanics(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	require.Panics(t, func() { keeper.GetMintingDenom(ctx) })
}

func TestIsMintingDenomSet_Unset(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	isSet := keeper.MintingDenomSet(ctx)
	require.False(t, isSet)
}

func TestIsMintingDenomSet_AlreadySet(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()
	createTestMintingDenom(keeper, ctx)

	isSet := keeper.MintingDenomSet(ctx)
	require.True(t, isSet)
}
