package cli

import (
	"strconv"

	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdUpdatePauser() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-pauser [address]",
		Short: "Broadcast message update-pauser",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argAddress := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdatePauser(
				clientCtx.GetFromAddress().String(),
				argAddress,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
