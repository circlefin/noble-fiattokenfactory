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

package types_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/circlefin/noble-fiattokenfactory/testutil/sample"
	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"
)

func createValidGenesis() *types.GenesisState {
	return &types.GenesisState{
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
				Address:   sample.AccAddress(),
				Allowance: sdk.NewCoin("uusdc", math.NewInt(1)),
			},
			{
				Address:   sample.AccAddress(),
				Allowance: sdk.NewCoin("uusdc", math.NewInt(1)),
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

func TestGenesisState_Validate(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		genState func() *types.GenesisState
		valid    bool
		error    string
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis,
			valid:    true,
		},
		{
			desc:     "happy path",
			genState: createValidGenesis,
			valid:    true,
		},
		{
			desc: "duplicated blacklist entries",
			genState: func() *types.GenesisState {
				genesis := createValidGenesis()
				addressBz := sample.AddressBz()
				genesis.BlacklistedList = []types.Blacklisted{
					{
						AddressBz: addressBz,
					}, {
						AddressBz: addressBz,
					},
				}
				return genesis
			},
			valid: false,
			error: "duplicated index for blacklisted",
		},
		{
			desc: "duplicated minter entries",
			genState: func() *types.GenesisState {
				genesis := createValidGenesis()
				address := sample.AccAddress()
				genesis.MintersList = []types.Minters{
					{
						Address:   address,
						Allowance: sdk.Coin{Denom: "uusdc", Amount: math.NewInt(10)},
					}, {
						Address:   address,
						Allowance: sdk.Coin{Denom: "uusdc", Amount: math.NewInt(5)},
					},
				}
				return genesis
			},
			valid: false,
			error: "duplicated index for minters",
		},
		{
			desc: "invalid minter address",
			genState: func() *types.GenesisState {
				genesis := createValidGenesis()
				genesis.MintersList = []types.Minters{
					{
						Address:   "not an address",
						Allowance: sdk.Coin{Denom: "uusdc", Amount: math.NewInt(10)},
					},
				}
				return genesis
			},
			valid: false,
			error: "invalid minter address",
		},
		{
			desc: "minter allowance is nil",
			genState: func() *types.GenesisState {
				genesis := createValidGenesis()
				genesis.MintersList = []types.Minters{
					{
						Address: sample.AccAddress(),
					},
				}
				return genesis
			},
			valid: false,
			error: "minter allowance cannot be nil or negative",
		},
		{
			desc: "minter allowance is negative",
			genState: func() *types.GenesisState {
				genesis := createValidGenesis()
				genesis.MintersList = []types.Minters{
					{
						Address:   sample.AccAddress(),
						Allowance: sdk.Coin{Denom: "uusdc", Amount: math.NewInt(-1)},
					},
				}
				return genesis
			},
			valid: false,
			error: "minter allowance cannot be nil or negative",
		},
		{
			desc: "minter allowance is zero succeeds",
			genState: func() *types.GenesisState {
				genesis := createValidGenesis()
				genesis.MintersList = []types.Minters{
					{
						Address:   sample.AccAddress(),
						Allowance: sdk.Coin{Denom: "uusdc", Amount: math.NewInt(0)},
					},
				}
				return genesis
			},
			valid: true,
		},
		{
			desc: "minter allowance is missing a denom",
			genState: func() *types.GenesisState {
				genesis := createValidGenesis()
				genesis.MintersList = []types.Minters{
					{
						Address:   sample.AccAddress(),
						Allowance: sdk.Coin{Amount: math.NewInt(-1)},
					},
				}
				return genesis
			},
			valid: false,
			error: "minter allowance cannot be nil or negative",
		},
		{
			desc: "minter allowance has empty denom",
			genState: func() *types.GenesisState {
				genesis := createValidGenesis()
				genesis.MintersList = []types.Minters{
					{
						Address:   sample.AccAddress(),
						Allowance: sdk.Coin{Denom: "", Amount: math.NewInt(-1)},
					},
				}
				return genesis
			},
			valid: false,
			error: "minter allowance cannot be nil or negative",
		},
		{
			desc: "duplicated minter controller entries",
			genState: func() *types.GenesisState {
				genesis := createValidGenesis()
				controller := sample.AccAddress()
				genesis.MinterControllerList = []types.MinterController{
					{
						Controller: controller,
						Minter:     sample.AccAddress(),
					}, {
						Controller: controller,
						Minter:     sample.AccAddress(),
					},
				}
				return genesis
			},
			valid: false,
			error: "duplicated index for minterController",
		},
		{
			desc: "minter controller has invalid minter address",
			genState: func() *types.GenesisState {
				genesis := createValidGenesis()
				genesis.MinterControllerList = []types.MinterController{
					{
						Controller: sample.AccAddress(),
						Minter:     "not an address",
					},
				}
				return genesis
			},
			valid: false,
			error: "minter controller has invalid minter address",
		},
		{
			desc: "minter controller has invalid controller address",
			genState: func() *types.GenesisState {
				genesis := createValidGenesis()
				genesis.MinterControllerList = []types.MinterController{
					{
						Controller: "not an address",
						Minter:     sample.AccAddress(),
					},
				}
				return genesis
			},
			valid: false,
			error: "minter controller has invalid controller address",
		},
		// {
		// 	desc: "owner address is not provided",
		// 	genState: func() *types.GenesisState {
		// 		genesis := createValidGenesis()
		// 		genesis.Owner = nil
		// 		return genesis
		// 	},
		// 	valid: false,
		// 	error: "owner address must be provided",
		// },
		{
			desc: "owner address is invalid",
			genState: func() *types.GenesisState {
				genesis := createValidGenesis()
				genesis.Owner = &types.Owner{
					Address: "not an address",
				}
				return genesis
			},
			valid: false,
			error: "invalid owner address",
		},
		{
			desc: "master minter address is invalid",
			genState: func() *types.GenesisState {
				genesis := createValidGenesis()
				genesis.MasterMinter = &types.MasterMinter{
					Address: "not an address",
				}
				return genesis
			},
			valid: false,
			error: "invalid master minter address",
		},
		{
			desc: "pauser address is invalid",
			genState: func() *types.GenesisState {
				genesis := createValidGenesis()
				genesis.Pauser = &types.Pauser{
					Address: "not an address",
				}
				return genesis
			},
			valid: false,
			error: "invalid pauser address",
		},
		// {
		// 	desc: "pause state is not provided",
		// 	genState: func() *types.GenesisState {
		// 		genesis := createValidGenesis()
		// 		genesis.Paused = nil
		// 		return genesis
		// 	},
		// 	valid: false,
		// 	error: "paused state must be provided",
		// },
		{
			desc: "blacklister address is invalid",
			genState: func() *types.GenesisState {
				genesis := createValidGenesis()
				genesis.Blacklister = &types.Blacklister{
					Address: "not an address",
				}
				return genesis
			},
			valid: false,
			error: "invalid black lister address",
		},
		{
			desc: "insufficient separation of privileges",
			genState: func() *types.GenesisState {
				genesis := createValidGenesis()
				address := sample.AccAddress()
				genesis.Owner = &types.Owner{
					Address: address,
				}
				genesis.MasterMinter = &types.MasterMinter{
					Address: address,
				}
				return genesis
			},
			valid: false,
			error: "address is already assigned to privileged role",
		},
		// {
		// 	desc: "minting denom is not provided",
		// 	genState: func() *types.GenesisState {
		// 		genesis := createValidGenesis()
		// 		genesis.MintingDenom = nil
		// 		return genesis
		// 	},
		// 	valid: false,
		// 	error: "minting denom must be provided",
		// },
		{
			desc: "minting denom is empty",
			genState: func() *types.GenesisState {
				genesis := createValidGenesis()
				genesis.MintingDenom = &types.MintingDenom{
					Denom: "",
				}
				return genesis
			},
			valid: false,
			error: "minting denom cannot be an empty string",
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState().Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.error)
			}
		})
	}
}
