package keeper

import (
	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

func (k *Keeper) HandleDeliverTxEvent(ctx sdk.Context, event abci.Event) error {
	if event.Type == banktypes.EventTypeTransfer {
		var recipient string
		var sender string
		coins := sdk.NewCoins()

		for _, attribute := range event.Attributes {
			switch string(attribute.Key) {
			case banktypes.AttributeKeyRecipient:
				recipient = string(attribute.Value)
			case banktypes.AttributeKeySender:
				sender = string(attribute.Value)
			case sdk.AttributeKeyAmount:
				coins, _ = sdk.ParseCoinsNormalized(string(attribute.Value))
			}
		}

		if !coins.AmountOf(k.GetMintingDenom(ctx).Denom).IsZero() {
			// Check paused state.
			if k.GetPaused(ctx).Paused {
				return errors.Wrap(types.ErrPaused, "can not perform token transfers")
			}

			// Check if sender is blacklisted.
			_, found := k.GetBlacklisted(ctx, sdk.MustAccAddressFromBech32(sender))
			if found {
				return errors.Wrapf(types.ErrUnauthorized, "%s can not send tokens", sender)
			}

			// Check if recipient is blacklisted.
			_, found = k.GetBlacklisted(ctx, sdk.MustAccAddressFromBech32(recipient))
			if found {
				return errors.Wrapf(types.ErrUnauthorized, "%s can not receive tokens", recipient)
			}
		}
	}

	return nil
}
