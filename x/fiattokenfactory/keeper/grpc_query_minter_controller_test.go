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
	"strconv"
	"testing"

	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/circlefin/noble-fiattokenfactory/testutil/keeper"
	"github.com/circlefin/noble-fiattokenfactory/testutil/nullify"
	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func TestMinterControllerQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	msgs := createNMinterController(keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetMinterControllerRequest
		response *types.QueryGetMinterControllerResponse
		err      error
	}{
		{
			desc: "FirstControllerSuccess",
			request: &types.QueryGetMinterControllerRequest{
				ControllerAddress: msgs[0].Controller,
			},
			response: &types.QueryGetMinterControllerResponse{MinterController: msgs[0]},
		},
		{
			desc: "SecondControllerSuccess",
			request: &types.QueryGetMinterControllerRequest{
				ControllerAddress: msgs[1].Controller,
			},
			response: &types.QueryGetMinterControllerResponse{MinterController: msgs[1]},
		},
		{
			desc: "AddressIsNotController",
			request: &types.QueryGetMinterControllerRequest{
				ControllerAddress: strconv.Itoa(100000),
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.MinterController(ctx, tc.request)
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

func TestMinterControllerQueryPaginated_NoControllers(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	resp, err := keeper.MinterControllerAll(ctx, createQueryAllMinterControllerRequest(nil, 0, 0, true))
	require.NoError(t, err)
	require.Equal(t, 0, int(resp.Pagination.Total))
}

func TestMinterControllerQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	msgs := createNMinterController(keeper, ctx, 5)

	t.Run("LookupByOffset", func(t *testing.T) {
		step := 2
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.MinterControllerAll(ctx, createQueryAllMinterControllerRequest(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.MinterController), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.MinterController),
			)
		}
	})
	t.Run("LookupByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.MinterControllerAll(ctx, createQueryAllMinterControllerRequest(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.MinterController), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.MinterController),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("LookupAll", func(t *testing.T) {
		resp, err := keeper.MinterControllerAll(ctx, createQueryAllMinterControllerRequest(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.MinterController),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.MinterControllerAll(ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
	t.Run("PaginateError", func(t *testing.T) {
		_, err := keeper.MinterControllerAll(ctx, createQueryAllMinterControllerRequest([]byte("next bytes"), 1, 0, true))
		require.ErrorContains(t, err, "invalid request, either offset or key is expected, got both")
	})
}

func createQueryAllMinterControllerRequest(next []byte, offset uint64, limit uint64, total bool) *types.QueryAllMinterControllerRequest {
	return &types.QueryAllMinterControllerRequest{
		Pagination: &query.PageRequest{
			Key:        next,
			Offset:     offset,
			Limit:      limit,
			CountTotal: total,
		},
	}
}
