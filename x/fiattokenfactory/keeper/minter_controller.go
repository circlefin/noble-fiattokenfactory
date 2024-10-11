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

	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
)

// SetMinterController set a specific minterController in the store from its index
func (k Keeper) SetMinterController(ctx context.Context, minterController types.MinterController) {
	adapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(adapter, types.KeyPrefix(types.MinterControllerKeyPrefix))
	b := k.cdc.MustMarshal(&minterController)
	store.Set(types.MinterControllerKey(
		minterController.Controller,
	), b)
}

// GetMinterController returns a minterController from its index
func (k Keeper) GetMinterController(
	ctx context.Context,
	controller string,
) (val types.MinterController, found bool) {
	adapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(adapter, types.KeyPrefix(types.MinterControllerKeyPrefix))

	b := store.Get(types.MinterControllerKey(
		controller,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveMinterController removes a minterController from the store
func (k Keeper) DeleteMinterController(
	ctx context.Context,
	controller string,
) {
	adapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(adapter, types.KeyPrefix(types.MinterControllerKeyPrefix))
	store.Delete(types.MinterControllerKey(
		controller,
	))
}

// GetAllMinterController returns all minterController
func (k Keeper) GetAllMinterControllers(ctx context.Context) (list []types.MinterController) {
	adapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(adapter, types.KeyPrefix(types.MinterControllerKeyPrefix))
	iterator := store.Iterator(nil, nil)

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.MinterController
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
