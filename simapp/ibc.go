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

package simapp

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/circlefin/noble-fiattokenfactory/x/blockibc"
	"github.com/cosmos/cosmos-sdk/runtime"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/ibc-go/v10/modules/apps/transfer"
	transferkeeper "github.com/cosmos/ibc-go/v10/modules/apps/transfer/keeper"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v10/modules/core"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v10/modules/core/03-connection/types"
	porttypes "github.com/cosmos/ibc-go/v10/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v10/modules/core/keeper"
	soloclient "github.com/cosmos/ibc-go/v10/modules/light-clients/06-solomachine"
	tmclient "github.com/cosmos/ibc-go/v10/modules/light-clients/07-tendermint"
)

func (app *SimApp) RegisterIBCModules() error {
	if err := app.RegisterStores(
		storetypes.NewKVStoreKey(ibcexported.StoreKey),
		storetypes.NewKVStoreKey(transfertypes.StoreKey),
	); err != nil {
		return err
	}

	app.ParamsKeeper.Subspace(ibcexported.ModuleName).WithKeyTable(clienttypes.ParamKeyTable().RegisterParamSet(&connectiontypes.Params{}))
	app.ParamsKeeper.Subspace(transfertypes.ModuleName).WithKeyTable(transfertypes.ParamKeyTable())

	app.IBCKeeper = ibckeeper.NewKeeper(
		app.appCodec,
		runtime.NewKVStoreService(app.GetKey(ibcexported.StoreKey)),
		app.GetSubspace(ibcexported.ModuleName),
		app.UpgradeKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.TransferKeeper = transferkeeper.NewKeeper(
		app.appCodec,
		runtime.NewKVStoreService(app.GetKey(transfertypes.StoreKey)),
		app.GetSubspace(transfertypes.ModuleName),
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.MsgServiceRouter(),
		app.AccountKeeper,
		app.BankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	ibcRouter := porttypes.NewRouter().AddRoute(transfertypes.ModuleName, blockibc.NewIBCMiddleware(transfer.NewIBCModule(app.TransferKeeper), app.FiatTokenFactoryKeeper))
	app.IBCKeeper.SetRouter(ibcRouter)

	storeProvider := app.IBCKeeper.ClientKeeper.GetStoreProvider()
	tmLightClientModule := tmclient.NewLightClientModule(app.appCodec, storeProvider)
	app.IBCKeeper.ClientKeeper.AddRoute(tmclient.ModuleName, &tmLightClientModule)

	smLightClientModule := soloclient.NewLightClientModule(app.appCodec, storeProvider)
	app.IBCKeeper.ClientKeeper.AddRoute(soloclient.ModuleName, &smLightClientModule)

	if err := app.RegisterModules(
		ibc.NewAppModule(app.IBCKeeper),
		transfer.NewAppModule(app.TransferKeeper),
		tmclient.NewAppModule(tmLightClientModule),
		soloclient.NewAppModule(smLightClientModule),
	); err != nil {
		return err
	}

	return nil
}
