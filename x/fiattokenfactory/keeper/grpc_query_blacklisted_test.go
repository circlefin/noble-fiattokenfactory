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

	"github.com/btcsuite/btcd/btcutil/bech32"
	"github.com/circlefin/noble-fiattokenfactory/testutil/sample"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

func TestBlacklistedQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNBlacklisted(keeper, ctx, 2)
	msgs = append(msgs, createNBlacklistedBech32m(keeper, ctx, 1)...)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetBlacklistedRequest
		response *types.QueryGetBlacklistedResponse
		err      error
	}{
		{
			desc: "FirstBlacklistedSuccess",
			request: &types.QueryGetBlacklistedRequest{
				Address: msgs[0].address,
			},
			response: &types.QueryGetBlacklistedResponse{Blacklisted: msgs[0].bl},
		},
		{
			desc: "SecondBlacklistedSuccess",
			request: &types.QueryGetBlacklistedRequest{
				Address: msgs[1].address,
			},
			response: &types.QueryGetBlacklistedResponse{Blacklisted: msgs[1].bl},
		},
		{
			desc: "Bech32mBlacklistedSuccess",
			request: &types.QueryGetBlacklistedRequest{
				Address: msgs[2].address,
			},
			response: &types.QueryGetBlacklistedResponse{Blacklisted: msgs[2].bl},
		},
		{
			desc: "Bech32AddressIsNotBlacklisted",
			request: &types.QueryGetBlacklistedRequest{
				Address: sample.TestAccount().Address,
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "Bech32mAddressIsNotBlacklisted",
			request: &types.QueryGetBlacklistedRequest{
				Address: sample.TestAccountBech32m().Address,
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
		{
			desc: "MalformedAddress",
			request: &types.QueryGetBlacklistedRequest{
				Address: "malformed address",
			},
			err: bech32.ErrInvalidCharacter(' '),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.Blacklisted(wctx, tc.request)
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

func TestBlacklistedQueryPaginated_NoBlacklistedAddresses(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()
	wctx := sdk.WrapSDKContext(ctx)

	resp, err := keeper.BlacklistedAll(wctx, createQueryAllBlacklistedRequest(nil, 0, 0, true))
	require.NoError(t, err)
	require.Equal(t, 0, int(resp.Pagination.Total))
}

func TestBlacklistedQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper()
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNBlacklisted(keeper, ctx, 5)
	msgs = append(msgs, createNBlacklistedBech32m(keeper, ctx, 5)...)

	blacklisted := make([]types.Blacklisted, len(msgs))
	for i, msg := range msgs {
		blacklisted[i] = msg.bl
	}

	t.Run("LookupByOffset", func(t *testing.T) {
		step := 2
		for i := 0; i < len(blacklisted); i += step {
			resp, err := keeper.BlacklistedAll(wctx, createQueryAllBlacklistedRequest(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Blacklisted), step)
			require.Subset(t,
				nullify.Fill(blacklisted),
				nullify.Fill(resp.Blacklisted),
			)
		}
	})
	t.Run("LookupByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(blacklisted); i += step {
			resp, err := keeper.BlacklistedAll(wctx, createQueryAllBlacklistedRequest(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Blacklisted), step)
			require.Subset(t,
				nullify.Fill(blacklisted),
				nullify.Fill(resp.Blacklisted),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("LookupAll", func(t *testing.T) {
		resp, err := keeper.BlacklistedAll(wctx, createQueryAllBlacklistedRequest(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(blacklisted), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(blacklisted),
			nullify.Fill(resp.Blacklisted),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.BlacklistedAll(wctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
	t.Run("PaginateError", func(t *testing.T) {
		_, err := keeper.BlacklistedAll(wctx, createQueryAllBlacklistedRequest([]byte("next bytes"), 1, 0, true))
		require.ErrorContains(t, err, "invalid request, either offset or key is expected, got both")
	})
}

func createQueryAllBlacklistedRequest(next []byte, offset uint64, limit uint64, total bool) *types.QueryAllBlacklistedRequest {
	return &types.QueryAllBlacklistedRequest{
		Pagination: &query.PageRequest{
			Key:        next,
			Offset:     offset,
			Limit:      limit,
			CountTotal: total,
		},
	}
}
