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
	"fmt"
	"testing"

	testkeeper "github.com/circlefin/noble-fiattokenfactory/testutil/keeper"
	"github.com/circlefin/noble-fiattokenfactory/testutil/sample"
	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/keeper"
	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestRemoveMinterController_MasterMinterNotSet(t *testing.T) {
	ftf, ctx := testkeeper.FiatTokenfactoryKeeper()
	msgServer := keeper.NewMsgServerImpl(ftf)

	_, err := msgServer.RemoveMinterController(sdk.WrapSDKContext(ctx), &types.MsgRemoveMinterController{})
	require.ErrorIs(t, types.ErrUserNotFound, err)
	require.ErrorContains(t, err, "master minter is not set")
}

func TestRemoveMinterController_FromAddressIsNotMasterMinter(t *testing.T) {
	masterMinter := sample.TestAccount()
	ftf, ctx := testkeeper.FiatTokenfactoryKeeper()
	msgServer := keeper.NewMsgServerImpl(ftf)
	ftf.SetMasterMinter(ctx, types.MasterMinter{Address: masterMinter.Address})

	_, err := msgServer.RemoveMinterController(sdk.WrapSDKContext(ctx), &types.MsgRemoveMinterController{From: sample.AccAddress()})
	require.ErrorIs(t, types.ErrUnauthorized, err)
	require.ErrorContains(t, err, "you are not the master minter")
}

func TestRemoveMinterController_ControllerDoesNotExist(t *testing.T) {
	masterMinter := sample.TestAccount()
	notController := sample.AccAddress()
	ftf, ctx := testkeeper.FiatTokenfactoryKeeper()
	msgServer := keeper.NewMsgServerImpl(ftf)
	ftf.SetMasterMinter(ctx, types.MasterMinter{Address: masterMinter.Address})

	_, err := msgServer.RemoveMinterController(sdk.WrapSDKContext(ctx), &types.MsgRemoveMinterController{From: masterMinter.Address, Controller: notController})
	require.ErrorIs(t, types.ErrUserNotFound, err)
	require.ErrorContains(t, err, fmt.Sprintf("minter controller with a given address (%s) doesn't exist", notController))
}

func TestRemoveMinterController_Success(t *testing.T) {
	masterMinter := sample.TestAccount()
	controller := sample.TestAccount()
	minter := sample.TestAccount()
	ftf, ctx := testkeeper.FiatTokenfactoryKeeper()
	msgServer := keeper.NewMsgServerImpl(ftf)
	ftf.SetMasterMinter(ctx, types.MasterMinter{Address: masterMinter.Address})
	ftf.SetMinterController(ctx, types.MinterController{Minter: minter.Address, Controller: controller.Address})

	res, err := msgServer.RemoveMinterController(sdk.WrapSDKContext(ctx), &types.MsgRemoveMinterController{From: masterMinter.Address, Controller: controller.Address})
	require.NoError(t, err)
	require.Equal(t, &types.MsgRemoveMinterControllerResponse{}, res)
}
