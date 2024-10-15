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

func createTestPaused(keeper *keeper.Keeper, ctx sdk.Context, isPaused bool) types.Paused {
	item := types.Paused{Paused: isPaused}
	keeper.SetPaused(ctx, item)
	return item
}

func TestPausedGetAndSet_Success(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	pauseState := types.Paused{Paused: true}
	keeper.SetPaused(ctx, pauseState)

	rst := keeper.GetPaused(ctx)
	require.Equal(t,
		nullify.Fill(&pauseState),
		nullify.Fill(&rst),
	)
}

func TestPausedSet_Overwrite(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	pausedState := types.Paused{Paused: true}
	keeper.SetPaused(ctx, pausedState)

	rst := keeper.GetPaused(ctx)
	require.Equal(t,
		nullify.Fill(&pausedState),
		nullify.Fill(&rst),
	)

	unpausedState := types.Paused{Paused: true}
	keeper.SetPaused(ctx, unpausedState)

	rst = keeper.GetPaused(ctx)
	require.Equal(t,
		nullify.Fill(&unpausedState),
		nullify.Fill(&rst),
	)
}

func TestPausedGet_UnsetPanics(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	require.Panics(t, func() { keeper.GetPaused(ctx) })
}
