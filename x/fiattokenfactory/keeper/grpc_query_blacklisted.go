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
	"context"

	"cosmossdk.io/store/prefix"
	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) BlacklistedAll(ctx context.Context, req *types.QueryAllBlacklistedRequest) (*types.QueryAllBlacklistedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var blacklisteds []types.Blacklisted

	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	blacklistedStore := prefix.NewStore(store, types.KeyPrefix(types.BlacklistedKeyPrefix))

	pageRes, err := query.Paginate(blacklistedStore, req.Pagination, func(key []byte, value []byte) error {
		var blacklisted types.Blacklisted
		if err := k.cdc.Unmarshal(value, &blacklisted); err != nil {
			return err
		}

		blacklisteds = append(blacklisteds, blacklisted)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllBlacklistedResponse{Blacklisted: blacklisteds, Pagination: pageRes}, nil
}

func (k Keeper) Blacklisted(ctx context.Context, req *types.QueryGetBlacklistedRequest) (*types.QueryGetBlacklistedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	_, addressBz, err := DecodeNoLimitToBase256(req.Address)
	if err != nil {
		return nil, err
	}

	val, found := k.GetBlacklisted(ctx, addressBz)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetBlacklistedResponse{Blacklisted: val}, nil
}
