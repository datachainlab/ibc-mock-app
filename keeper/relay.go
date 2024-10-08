package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
)

func (k Keeper) sendPacket(
	ctx sdk.Context,
	sourcePort,
	sourceChannel string,
	message string,
	sender sdk.AccAddress,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
) error {
	channelCap, ok := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(sourcePort, sourceChannel))
	if !ok {
		return errorsmod.Wrap(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability")
	}

	_, err := k.ics4Wrapper.SendPacket(
		ctx,
		channelCap,
		sourcePort,
		sourceChannel,
		timeoutHeight,
		timeoutTimestamp,
		[]byte(message))
	return err
}
