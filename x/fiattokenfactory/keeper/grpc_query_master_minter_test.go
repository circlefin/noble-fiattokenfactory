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

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/circlefin/noble-fiattokenfactory/testutil/keeper"
	"github.com/circlefin/noble-fiattokenfactory/testutil/nullify"
	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
)

func TestMasterMinterQuery_NoMasterMinterConfigured(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	_, err := keeper.MasterMinter(ctx, &types.QueryGetMasterMinterRequest{})
	require.ErrorIs(t, err, status.Error(codes.NotFound, "not found"))
}

func TestMasterMinterQuery(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	expectedMasterMinter := createTestMasterMinter(keeper, ctx)

	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetMasterMinterRequest
		response *types.QueryGetMasterMinterResponse
		err      error
	}{
		{
			desc:     "Success",
			request:  &types.QueryGetMasterMinterRequest{},
			response: &types.QueryGetMasterMinterResponse{MasterMinter: expectedMasterMinter},
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.MasterMinter(ctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.Equal(t,
					nullify.Fill(tc.response),
					nullify.Fill(response),
				)
			}
		})
	}
}
