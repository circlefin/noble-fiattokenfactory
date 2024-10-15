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
	"fmt"

	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	"github.com/cosmos/cosmos-sdk/runtime"

	"cosmossdk.io/store/prefix"
)

// SetMintingDenom set mintingDenom in the store
func (k *Keeper) SetMintingDenom(ctx context.Context, mintingDenom types.MintingDenom) {
	if k.MintingDenomSet(ctx) {
		panic(types.ErrMintingDenomSet)
	}

	_, found := k.bankKeeper.GetDenomMetaData(ctx, mintingDenom.Denom)
	if !found {
		panic(fmt.Sprintf("Denom metadata for '%s' should be set", mintingDenom.Denom))
	}

	adapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(adapter, types.KeyPrefix(types.MintingDenomKey))
	b := k.cdc.MustMarshal(&mintingDenom)
	store.Set(types.KeyPrefix(types.MintingDenomKey), b)
}

// GetMintingDenom returns mintingDenom
func (k *Keeper) GetMintingDenom(ctx context.Context) (val types.MintingDenom) {
	adapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(adapter, types.KeyPrefix(types.MintingDenomKey))

	b := store.Get(types.KeyPrefix(types.MintingDenomKey))
	if b == nil {
		panic("Minting denom is not set")
	}

	k.cdc.MustUnmarshal(b, &val)
	return val
}

// MintingDenomSet returns true if the MintingDenom is already set in the store, it returns false otherwise.
func (k Keeper) MintingDenomSet(ctx context.Context) bool {
	adapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(adapter, types.KeyPrefix(types.MintingDenomKey))

	b := store.Get(types.KeyPrefix(types.MintingDenomKey))

	return b != nil
}
