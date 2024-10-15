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

package fiattokenfactory_test

import (
	"testing"

	keepertest "github.com/circlefin/noble-fiattokenfactory/testutil/keeper"
	"github.com/circlefin/noble-fiattokenfactory/testutil/nullify"
	"github.com/circlefin/noble-fiattokenfactory/testutil/sample"
	fiattokenfactory "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory"
	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"

	"github.com/stretchr/testify/require"
)

func TestInitGenesis_denomWithoutMetadata(t *testing.T) {
	genesisState := types.GenesisState{
		BlacklistedList:      []types.Blacklisted{},
		Paused:               nil,
		MasterMinter:         nil,
		MintersList:          []types.Minters{},
		Pauser:               nil,
		Blacklister:          nil,
		Owner:                nil,
		MinterControllerList: []types.MinterController{},
		MintingDenom: &types.MintingDenom{
			Denom: "notadenom",
		},
	}

	k, ctx := keepertest.FiatTokenfactoryKeeper()
	require.Panics(t, func() { fiattokenfactory.InitGenesis(ctx, k, keepertest.MockBankKeeper{}, genesisState) })
}

func TestInitGenesis_MissingFields(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		genState types.GenesisState
	}{
		{
			desc: "all nil or empty list",
			genState: types.GenesisState{
				BlacklistedList:      []types.Blacklisted{},
				Paused:               nil,
				MasterMinter:         nil,
				MintersList:          []types.Minters{},
				Pauser:               nil,
				Blacklister:          nil,
				Owner:                nil,
				MinterControllerList: []types.MinterController{},
				MintingDenom:         nil,
			},
		},
		{
			desc:     "all unspecified",
			genState: types.GenesisState{},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			k, ctx := keepertest.FiatTokenfactoryKeeper()
			fiattokenfactory.InitGenesis(ctx, k, keepertest.MockBankKeeper{}, tc.genState)

			_, found := k.GetOwner(ctx)
			require.False(t, found)
			_, found = k.GetMasterMinter(ctx)
			require.False(t, found)
			_, found = k.GetPauser(ctx)
			require.False(t, found)
			_, found = k.GetBlacklister(ctx)
			require.False(t, found)

			require.Panics(t, func() { k.GetMintingDenom(ctx) })
			require.Panics(t, func() { k.GetPaused(ctx) })

			require.Empty(t, k.GetAllBlacklisted(ctx))
			require.Empty(t, k.GetAllMinters(ctx))
			require.Empty(t, k.GetAllMinterControllers(ctx))
		})
	}
}

func TestInitGenesis_allSpecifiedPasses(t *testing.T) {
	genesisState := createCompleteValidGenesis()

	k, ctx := keepertest.FiatTokenfactoryKeeper()
	fiattokenfactory.InitGenesis(ctx, k, keepertest.MockBankKeeper{}, genesisState)

	owner, found := k.GetOwner(ctx)
	require.True(t, found)
	require.Equal(t, genesisState.Owner, &owner)
	masterMinter, found := k.GetMasterMinter(ctx)
	require.True(t, found)
	require.Equal(t, genesisState.MasterMinter, &masterMinter)
	pauser, found := k.GetPauser(ctx)
	require.True(t, found)
	require.Equal(t, genesisState.Pauser, &pauser)
	blacklister, found := k.GetBlacklister(ctx)
	require.True(t, found)
	require.Equal(t, genesisState.Blacklister, &blacklister)

	paused := k.GetPaused(ctx)
	require.Equal(t, genesisState.Paused, &paused)
	denom := k.GetMintingDenom(ctx)
	require.Equal(t, genesisState.MintingDenom, &denom)

	require.ElementsMatch(t,
		nullify.Fill(genesisState.BlacklistedList),
		nullify.Fill(k.GetAllBlacklisted(ctx)))
	require.ElementsMatch(t,
		nullify.Fill(genesisState.MintersList),
		nullify.Fill(k.GetAllMinters(ctx)))
	require.ElementsMatch(t,
		nullify.Fill(genesisState.MinterControllerList),
		nullify.Fill(k.GetAllMinterControllers(ctx)))
}

func TestExportGenesis(t *testing.T) {
	genesisState := createCompleteValidGenesis()

	k, ctx := keepertest.FiatTokenfactoryKeeper()
	fiattokenfactory.InitGenesis(ctx, k, keepertest.MockBankKeeper{}, genesisState)

	got := fiattokenfactory.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.ElementsMatch(t, genesisState.BlacklistedList, got.BlacklistedList)
	require.Equal(t, genesisState.Paused, got.Paused)
	require.Equal(t, genesisState.MasterMinter, got.MasterMinter)
	require.ElementsMatch(t, genesisState.MintersList, got.MintersList)
	require.Equal(t, genesisState.Pauser, got.Pauser)
	require.Equal(t, genesisState.Blacklister, got.Blacklister)
	require.Equal(t, genesisState.Owner, got.Owner)
	require.ElementsMatch(t, genesisState.MinterControllerList, got.MinterControllerList)
	require.Equal(t, genesisState.MintingDenom, got.MintingDenom)
}

func createCompleteValidGenesis() types.GenesisState {
	return types.GenesisState{
		BlacklistedList: []types.Blacklisted{
			{
				AddressBz: sample.AddressBz(),
			},
			{
				AddressBz: sample.AddressBz(),
			},
		},
		Paused: &types.Paused{
			Paused: true,
		},
		MasterMinter: &types.MasterMinter{
			Address: sample.AccAddress(),
		},
		MintersList: []types.Minters{
			{
				Address: sample.AccAddress(),
			},
			{
				Address: sample.AccAddress(),
			},
		},
		Pauser: &types.Pauser{
			Address: sample.AccAddress(),
		},
		Blacklister: &types.Blacklister{
			Address: sample.AccAddress(),
		},
		Owner: &types.Owner{
			Address: sample.AccAddress(),
		},
		MinterControllerList: []types.MinterController{
			{
				Controller: sample.AccAddress(),
				Minter:     sample.AccAddress(),
			},
			{
				Controller: sample.AccAddress(),
				Minter:     sample.AccAddress(),
			},
		},
		MintingDenom: &types.MintingDenom{
			Denom: "uusdc",
		},
	}
}
