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
	"github.com/cosmos/cosmos-sdk/runtime"
)

// SetOwner set owner in the store
func (k Keeper) SetOwner(ctx context.Context, owner types.Owner) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	b := k.cdc.MustMarshal(&owner)
	store.Set(types.KeyPrefix(types.OwnerKey), b)
}

// GetOwner returns owner
func (k Keeper) GetOwner(ctx context.Context) (val types.Owner, found bool) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))

	b := store.Get(types.KeyPrefix(types.OwnerKey))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// SetPendingOwner set pending owner in the store
func (k Keeper) SetPendingOwner(ctx context.Context, owner types.Owner) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	b := k.cdc.MustMarshal(&owner)
	store.Set(types.KeyPrefix(types.PendingOwnerKey), b)
}

// DeletePendingOwner deletes the pending owner in the store
func (k Keeper) DeletePendingOwner(ctx context.Context) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store.Delete(types.KeyPrefix(types.PendingOwnerKey))
}

// GetPendingOwner returns pending owner
func (k Keeper) GetPendingOwner(ctx context.Context) (val types.Owner, found bool) {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))

	b := store.Get(types.KeyPrefix(types.PendingOwnerKey))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}
