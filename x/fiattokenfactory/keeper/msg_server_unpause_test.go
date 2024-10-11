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

	testkeeper "github.com/circlefin/noble-fiattokenfactory/testutil/keeper"
	"github.com/circlefin/noble-fiattokenfactory/testutil/sample"
	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/keeper"
	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestUnpause_PauserNotSet(t *testing.T) {
	ftf, ctx := testkeeper.FiatTokenfactoryKeeper()
	msgServer := keeper.NewMsgServerImpl(ftf)

	_, err := msgServer.Unpause(sdk.WrapSDKContext(ctx), &types.MsgUnpause{})
	require.ErrorIs(t, err, types.ErrUserNotFound)
	require.ErrorContains(t, err, "pauser is not set")
}

func TestUnpause_FromAddressIsNotPauser(t *testing.T) {
	pauser := sample.TestAccount()
	ftf, ctx := testkeeper.FiatTokenfactoryKeeper()
	msgServer := keeper.NewMsgServerImpl(ftf)
	ftf.SetPauser(ctx, types.Pauser{Address: pauser.Address})

	_, err := msgServer.Unpause(sdk.WrapSDKContext(ctx), &types.MsgUnpause{From: sample.AccAddress()})
	require.ErrorIs(t, err, types.ErrUnauthorized)
	require.ErrorContains(t, err, "you are not the pauser")
}

func TestUnpause_AlreadyUnpausedSuccess(t *testing.T) {
	pauser := sample.TestAccount()
	ftf, ctx := testkeeper.FiatTokenfactoryKeeper()
	msgServer := keeper.NewMsgServerImpl(ftf)
	ftf.SetPauser(ctx, types.Pauser{Address: pauser.Address})
	ftf.SetPaused(ctx, types.Paused{Paused: false})

	res, err := msgServer.Unpause(sdk.WrapSDKContext(ctx), &types.MsgUnpause{From: pauser.Address})
	require.NoError(t, err)
	require.Equal(t, &types.MsgUnpauseResponse{}, res)
}

func TestUnpause_Success(t *testing.T) {
	pauser := sample.TestAccount()
	ftf, ctx := testkeeper.FiatTokenfactoryKeeper()
	msgServer := keeper.NewMsgServerImpl(ftf)
	ftf.SetPauser(ctx, types.Pauser{Address: pauser.Address})
	ftf.SetPaused(ctx, types.Paused{Paused: true})

	res, err := msgServer.Unpause(sdk.WrapSDKContext(ctx), &types.MsgUnpause{From: pauser.Address})
	require.NoError(t, err)
	require.Equal(t, &types.MsgUnpauseResponse{}, res)
}
