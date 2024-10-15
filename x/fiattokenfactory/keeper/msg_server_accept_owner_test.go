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

func TestAcceptOwner_NoPendingOwner(t *testing.T) {
	ftf, ctx := testkeeper.FiatTokenfactoryKeeper()
	msgServer := keeper.NewMsgServerImpl(ftf)

	_, err := msgServer.AcceptOwner(sdk.WrapSDKContext(ctx), &types.MsgAcceptOwner{From: sample.AccAddress()})
	require.ErrorIs(t, types.ErrUserNotFound, err)
}

func TestAcceptOwner_FromAddressIsNotPendingOwner(t *testing.T) {
	ftf, ctx := testkeeper.FiatTokenfactoryKeeper()
	msgServer := keeper.NewMsgServerImpl(ftf)

	ftf.SetPendingOwner(ctx, types.Owner{Address: sample.AccAddress()})
	_, err := msgServer.AcceptOwner(sdk.WrapSDKContext(ctx), &types.MsgAcceptOwner{From: sample.AccAddress()})
	require.ErrorIs(t, err, types.ErrUnauthorized)
}

func TestAcceptOwner_Success(t *testing.T) {
	ftf, ctx := testkeeper.FiatTokenfactoryKeeper()
	msgServer := keeper.NewMsgServerImpl(ftf)
	pendingOwner := sample.AccAddress()

	ftf.SetPendingOwner(ctx, types.Owner{Address: pendingOwner})
	res, err := msgServer.AcceptOwner(sdk.WrapSDKContext(ctx), &types.MsgAcceptOwner{From: pendingOwner})
	require.NoError(t, err)
	require.Equal(t, &types.MsgAcceptOwnerResponse{}, res)
}
