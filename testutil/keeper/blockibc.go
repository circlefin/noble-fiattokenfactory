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

package keeper

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/circlefin/noble-fiattokenfactory/x/blockibc"
	fiattokenfactorykeeper "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/keeper"
	fiattokenfactorytypes "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v10/modules/core/exported"
	"github.com/cosmos/ibc-go/v10/testing/mock"
)

func BlockIBC() (blockibc.IBCMiddleware, *fiattokenfactorykeeper.Keeper, sdk.Context) {
	keys := storetypes.NewKVStoreKeys(fiattokenfactorytypes.StoreKey)
	ctx := testutil.DefaultContextWithKeys(keys, nil, nil)

	cdc := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())

	transferAppModule := mock.NewAppModule()
	transferIBCModule := mock.NewIBCModule(
		&transferAppModule,
		mock.NewIBCApp(transfertypes.ModuleName),
	)

	// override the mock ibc_module OnRecvPacket method since it expects specific packet data to return a successful acknowledgment.
	transferIBCModule.IBCApp.OnRecvPacket = func(ctx sdk.Context, channelVersion string, packet channeltypes.Packet, relayer sdk.AccAddress) exported.Acknowledgement {
		return mock.MockAcknowledgement
	}

	ftfKeeper := fiattokenfactorykeeper.NewKeeper(
		cdc, nil, runtime.NewKVStoreService(keys[fiattokenfactorytypes.StoreKey]), MockBankKeeper{},
	)
	ftfKeeper.SetMintingDenom(ctx, fiattokenfactorytypes.MintingDenom{Denom: "uusdc"})
	ftfKeeper.SetPaused(ctx, fiattokenfactorytypes.Paused{Paused: false})

	return blockibc.NewIBCMiddleware(
		transferIBCModule,
		ftfKeeper,
	), ftfKeeper, ctx
}
