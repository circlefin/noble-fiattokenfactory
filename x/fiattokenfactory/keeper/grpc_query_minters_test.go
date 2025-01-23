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

func TestMintersQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	msgs := createNMinters(keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetMintersRequest
		response *types.QueryGetMintersResponse
		err      error
	}{
		{
			desc: "FirstMinterSuccess",
			request: &types.QueryGetMintersRequest{
				Address: msgs[0].Address,
			},
			response: &types.QueryGetMintersResponse{Minters: msgs[0]},
		},
		{
			desc: "SecondMinterSuccess",
			request: &types.QueryGetMintersRequest{
				Address: msgs[1].Address,
			},
			response: &types.QueryGetMintersResponse{Minters: msgs[1]},
		},
		{
			desc: "AddressIsNotMinter",
			request: &types.QueryGetMintersRequest{
				Address: strconv.Itoa(100000),
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.Minters(ctx, tc.request)
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

func TestMintersQueryPaginated_NoMinters(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	resp, err := keeper.MintersAll(ctx, createQueryAllMintersRequest(nil, 0, 0, true))
	require.NoError(t, err)
	require.Equal(t, 0, int(resp.Pagination.Total))
}

func TestMintersQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()

	msgs := createNMinters(keeper, ctx, 5)

	t.Run("LookupByOffset", func(t *testing.T) {
		step := 2
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.MintersAll(ctx, createQueryAllMintersRequest(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Minters), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.Minters),
			)
		}
	})
	t.Run("LookupByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.MintersAll(ctx, createQueryAllMintersRequest(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Minters), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.Minters),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("LookupAll", func(t *testing.T) {
		resp, err := keeper.MintersAll(ctx, createQueryAllMintersRequest(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.Minters),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.MintersAll(ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
	t.Run("PaginateError", func(t *testing.T) {
		_, err := keeper.MintersAll(ctx, createQueryAllMintersRequest([]byte("next bytes"), 1, 0, true))
		require.ErrorContains(t, err, "invalid request, either offset or key is expected, got both")
	})
}

func createQueryAllMintersRequest(next []byte, offset uint64, limit uint64, total bool) *types.QueryAllMintersRequest {
	return &types.QueryAllMintersRequest{
		Pagination: &query.PageRequest{
			Key:        next,
			Offset:     offset,
			Limit:      limit,
			CountTotal: total,
		},
	}
}
