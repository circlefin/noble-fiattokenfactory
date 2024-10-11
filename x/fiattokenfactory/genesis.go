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

package fiattokenfactory

import (
	"cosmossdk.io/errors"

	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/keeper"
	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k *keeper.Keeper, bankKeeper types.BankKeeper, genState types.GenesisState) {
	for _, elem := range genState.BlacklistedList {
		k.SetBlacklisted(ctx, elem)
	}

	if genState.Paused != nil {
		k.SetPaused(ctx, *genState.Paused)
	}

	if genState.MasterMinter != nil {
		k.SetMasterMinter(ctx, *genState.MasterMinter)
	}

	for _, elem := range genState.MintersList {
		k.SetMinters(ctx, elem)
	}

	if genState.Pauser != nil {
		k.SetPauser(ctx, *genState.Pauser)
	}

	if genState.Blacklister != nil {
		k.SetBlacklister(ctx, *genState.Blacklister)
	}

	if genState.Owner != nil {
		k.SetOwner(ctx, *genState.Owner)
	}

	for _, elem := range genState.MinterControllerList {
		k.SetMinterController(ctx, elem)
	}

	if genState.MintingDenom != nil {
		_, found := bankKeeper.GetDenomMetaData(ctx, genState.MintingDenom.Denom)
		if !found {
			panic(errors.Wrapf(types.ErrDenomNotRegistered, "fiattokenfactory minting denom %s is not registered in bank module denom_metadata", genState.MintingDenom.Denom))
		}
		k.SetMintingDenom(ctx, *genState.MintingDenom)
	}
}

// ExportGenesis returns the module's exported GenesisState
func ExportGenesis(ctx sdk.Context, k *keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()

	genesis.BlacklistedList = k.GetAllBlacklisted(ctx)

	paused := k.GetPaused(ctx)
	genesis.Paused = &paused

	masterMinter, found := k.GetMasterMinter(ctx)
	if found {
		genesis.MasterMinter = &masterMinter
	}
	genesis.MintersList = k.GetAllMinters(ctx)

	pauser, found := k.GetPauser(ctx)
	if found {
		genesis.Pauser = &pauser
	}

	blacklister, found := k.GetBlacklister(ctx)
	if found {
		genesis.Blacklister = &blacklister
	}

	owner, found := k.GetOwner(ctx)
	if found {
		genesis.Owner = &owner
	}
	genesis.MinterControllerList = k.GetAllMinterControllers(ctx)

	mintingDenom := k.GetMintingDenom(ctx)
	genesis.MintingDenom = &mintingDenom

	return genesis
}
