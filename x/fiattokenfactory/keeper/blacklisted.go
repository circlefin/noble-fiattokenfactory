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
)

// SetBlacklisted set a specific blacklisted in the store from its index
func (k Keeper) SetBlacklisted(ctx context.Context, blacklisted types.Blacklisted) {
	adapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(adapter, types.KeyPrefix(types.BlacklistedKeyPrefix))
	b := k.cdc.MustMarshal(&blacklisted)
	store.Set(types.BlacklistedKey(blacklisted.AddressBz), b)
}

// GetBlacklisted returns a blacklisted from its index
func (k Keeper) GetBlacklisted(ctx context.Context, addressBz []byte) (val types.Blacklisted, found bool) {
	adapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(adapter, types.KeyPrefix(types.BlacklistedKeyPrefix))

	b := store.Get(types.BlacklistedKey(addressBz))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveBlacklisted removes a blacklisted from the store
func (k Keeper) RemoveBlacklisted(ctx context.Context, addressBz []byte) {
	adapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(adapter, types.KeyPrefix(types.BlacklistedKeyPrefix))
	store.Delete(types.BlacklistedKey(addressBz))
}

// GetAllBlacklisted returns all blacklisted
func (k Keeper) GetAllBlacklisted(ctx context.Context) (list []types.Blacklisted) {
	adapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(adapter, types.KeyPrefix(types.BlacklistedKeyPrefix))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Blacklisted
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
