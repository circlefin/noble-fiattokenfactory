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

func (k Keeper) MintersAll(ctx context.Context, req *types.QueryAllMintersRequest) (*types.QueryAllMintersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var minters []types.Minters

	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	mintersStore := prefix.NewStore(store, types.KeyPrefix(types.MintersKeyPrefix))

	pageRes, err := query.Paginate(mintersStore, req.Pagination, func(key []byte, value []byte) error {
		var minter types.Minters
		if err := k.cdc.Unmarshal(value, &minter); err != nil {
			return err
		}

		minters = append(minters, minter)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllMintersResponse{Minters: minters, Pagination: pageRes}, nil
}

func (k Keeper) Minters(ctx context.Context, req *types.QueryGetMintersRequest) (*types.QueryGetMintersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	val, found := k.GetMinters(
		ctx,
		req.Address,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetMintersResponse{Minters: val}, nil
}
