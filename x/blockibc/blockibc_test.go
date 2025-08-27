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

package blockibc_test

import (
	"testing"

	"github.com/circlefin/noble-fiattokenfactory/testutil/keeper"
	"github.com/circlefin/noble-fiattokenfactory/testutil/sample"
	fiattokenfactorytypes "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	"github.com/cosmos/cosmos-sdk/x/auth/codec"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	"github.com/stretchr/testify/require"
)

func TestBlockIBC(t *testing.T) {
	// ARRANGE: Mock sender and receiver.
	sender, receiver := sample.TestAccount(), sample.TestAccount()
	senderBech32m, receiverBech32m := sample.TestAccountBech32m(), sample.TestAccountBech32m()
	receiverAddress, _ := codec.NewBech32Codec("osmo").BytesToString(receiver.AddressBz)
	receiverBech32mAddress, _ := codec.NewBech32Codec("osmo").BytesToString(receiverBech32m.AddressBz)

	// ARRANGE: Organize table driven test cases.
	testCases := map[string]struct {
		toBlacklist         *sample.Account
		setPaused           bool
		packet              channeltypes.Packet
		expectSuccessfulAck bool
	}{
		"happy path": {
			toBlacklist:         nil,
			setPaused:           false,
			packet:              mockPacket(sender.Address, receiverAddress),
			expectSuccessfulAck: true,
		},
		"malformed ICS-20 packet data": {
			toBlacklist: nil,
			setPaused:   false,
			packet: func() channeltypes.Packet {
				packet := mockPacket(sender.Address, receiverAddress)
				packet.Data = []byte("malformed packet data")
				return packet
			}(),
			expectSuccessfulAck: false,
		},
		"uncontrolled denom": {
			toBlacklist: nil,
			setPaused:   false,
			packet: func() channeltypes.Packet {
				packet := mockPacket(sender.Address, receiverAddress)
				// transfer `ustake` instead of `uusdc`
				packet.Data = transfertypes.NewFungibleTokenPacketData(
					"ustake", "1000000", sender.Address, receiverAddress, "",
				).GetBytes()
				return packet
			}(),
			expectSuccessfulAck: true,
		},
		"tokenfactory paused": {
			toBlacklist:         nil,
			setPaused:           true,
			packet:              mockPacket(sender.Address, receiverAddress),
			expectSuccessfulAck: false,
		},
		"blacklisted bech32 sender": {
			toBlacklist:         &sender,
			setPaused:           false,
			packet:              mockPacket(sender.Address, receiverAddress),
			expectSuccessfulAck: false,
		},
		"blacklisted bech32m sender": {
			toBlacklist:         &senderBech32m,
			setPaused:           false,
			packet:              mockPacket(senderBech32m.Address, receiverAddress),
			expectSuccessfulAck: false,
		},
		"blacklisted bech32 receiver": {
			toBlacklist:         &receiver,
			setPaused:           false,
			packet:              mockPacket(sender.Address, receiverAddress),
			expectSuccessfulAck: false,
		},
		"blacklisted bech32m receiver": {
			toBlacklist:         &receiverBech32m,
			setPaused:           false,
			packet:              mockPacket(sender.Address, receiverBech32mAddress),
			expectSuccessfulAck: false,
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// ARRANGE: Mock middleware stack.
			middleware, ftf, ctx := keeper.BlockIBC()

			// ACT: Set paused and blacklisted state based on test case.
			ftf.SetPaused(ctx, fiattokenfactorytypes.Paused{Paused: tc.setPaused})
			if tc.toBlacklist != nil {
				ftf.SetBlacklisted(ctx, fiattokenfactorytypes.Blacklisted{
					AddressBz: tc.toBlacklist.AddressBz,
				})
			}

			// ACT: Receive transfer packet in middleware.
			ack := middleware.OnRecvPacket(ctx, transfertypes.PortID, tc.packet, nil)

			// ASSERT: Assert the acknowledgment's success based on the test case.
			var assertBool require.BoolAssertionFunc
			if tc.expectSuccessfulAck {
				assertBool = require.True
			} else {
				assertBool = require.False
			}
			assertBool(t, ack.Success())
		})
	}
}

func mockPacket(sender, receiver string) channeltypes.Packet {
	return channeltypes.NewPacket(
		transfertypes.NewFungibleTokenPacketData(
			"uusdc", "1000000", sender, receiver, "",
		).GetBytes(),
		1,
		transfertypes.PortID,
		"channel-0",
		transfertypes.PortID,
		"channel-0",
		clienttypes.Height{
			RevisionNumber: 0,
			RevisionHeight: 0,
		},
		1234,
	)
}
