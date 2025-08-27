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

package blockibc

import (
	"cosmossdk.io/errors"
	"github.com/btcsuite/btcd/btcutil/bech32"
	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/keeper"
	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v10/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"
)

var _ porttypes.IBCModule = &IBCMiddleware{}

// IBCMiddleware implements the tokenfactory keeper in order to check against blacklisted addresses.
type IBCMiddleware struct {
	app    porttypes.IBCModule
	keeper *keeper.Keeper
}

// NewIBCMiddleware creates a new IBCMiddleware given the keeper and underlying application.
func NewIBCMiddleware(app porttypes.IBCModule, k *keeper.Keeper) IBCMiddleware {
	return IBCMiddleware{
		app:    app,
		keeper: k,
	}
}

// OnChanOpenInit implements the IBCModule interface.
func (im IBCMiddleware) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	counterparty channeltypes.Counterparty,
	version string,
) (string, error) {
	return im.app.OnChanOpenInit(ctx, order, connectionHops, portID, channelID, counterparty, version)
}

// OnChanOpenTry implements the IBCModule interface.
func (im IBCMiddleware) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID, channelID string,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (version string, err error) {
	return im.app.OnChanOpenTry(ctx, order, connectionHops, portID, channelID, counterparty, counterpartyVersion)
}

// OnChanOpenAck implements the IBCModule interface.
func (im IBCMiddleware) OnChanOpenAck(
	ctx sdk.Context,
	portID, channelID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) error {
	return im.app.OnChanOpenAck(ctx, portID, channelID, counterpartyChannelID, counterpartyVersion)
}

// OnChanOpenConfirm implements the IBCModule interface.
func (im IBCMiddleware) OnChanOpenConfirm(ctx sdk.Context, portID, channelID string) error {
	return im.app.OnChanOpenConfirm(ctx, portID, channelID)
}

// OnChanCloseInit implements the IBCModule interface.
func (im IBCMiddleware) OnChanCloseInit(ctx sdk.Context, portID, channelID string) error {
	return im.app.OnChanCloseInit(ctx, portID, channelID)
}

// OnChanCloseConfirm implements the IBCModule interface.
func (im IBCMiddleware) OnChanCloseConfirm(ctx sdk.Context, portID, channelID string) error {
	return im.app.OnChanCloseConfirm(ctx, portID, channelID)
}

// OnRecvPacket intercepts the packet data and checks the sender and receiver address against
// the blacklisted addresses held in the tokenfactory keeper. If the address is found in the blacklist, an
// acknowledgment error is returned.
func (im IBCMiddleware) OnRecvPacket(
	ctx sdk.Context,
	portID string,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	var data transfertypes.FungibleTokenPacketData
	var ackErr error
	if err := types.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		ackErr = errors.Wrapf(types.ErrInvalidType, "cannot unmarshal ICS-20 transfer packet data")
		return channeltypes.NewErrorAcknowledgement(ackErr)
	}

	denomTrace := transfertypes.ParseDenomTrace(data.Denom)

	mintingDenom := im.keeper.GetMintingDenom(ctx)
	if denomTrace.Base != mintingDenom.Denom {
		return im.app.OnRecvPacket(ctx, portID, packet, relayer)
	}

	if im.keeper.GetPaused(ctx).Paused {
		return channeltypes.NewErrorAcknowledgement(types.ErrPaused)
	}

	_, addressBz, err := bech32.DecodeToBase256(data.Receiver)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	_, found := im.keeper.GetBlacklisted(ctx, addressBz)
	if found {
		ackErr = errors.Wrapf(types.ErrUnauthorized, "receiver address is blacklisted")
		return channeltypes.NewErrorAcknowledgement(ackErr)
	}

	_, addressBz, err = bech32.DecodeToBase256(data.Sender)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	_, found = im.keeper.GetBlacklisted(ctx, addressBz)
	if found {
		ackErr = errors.Wrapf(types.ErrUnauthorized, "sender address is blacklisted")
		return channeltypes.NewErrorAcknowledgement(ackErr)
	}

	return im.app.OnRecvPacket(ctx, portID, packet, relayer)
}

// OnAcknowledgementPacket implements the IBCModule interface.
func (im IBCMiddleware) OnAcknowledgementPacket(
	ctx sdk.Context,
	portID string,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	return im.app.OnAcknowledgementPacket(ctx, portID, packet, acknowledgement, relayer)
}

// OnTimeoutPacket implements the IBCModule interface.
func (im IBCMiddleware) OnTimeoutPacket(ctx sdk.Context, portID string, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	return im.app.OnTimeoutPacket(ctx, portID, packet, relayer)
}
