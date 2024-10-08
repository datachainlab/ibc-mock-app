package mockapp

import (
	"bytes"
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"

	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	"github.com/datachainlab/ibc-mock-app/keeper"
	"github.com/datachainlab/ibc-mock-app/types"
)

const (
	MOCK_PACKET_DATA       = "mock packet data"
	MOCK_ASYNC_PACKET_DATA = "mock async packet data"

	SUCCESSFUL_ACKNOWLEDGEMENT       = "mock acknowledgement"
	SUCCESSFUL_ASYNC_ACKNOWLEDGEMENT = "mock async acknowledgement"
	FAILED_ACKNOWLEDGEMENT           = "mock failed acknowledgement"
)

var (
	_ porttypes.IBCModule        = (*IBCModule)(nil)
	_ porttypes.UpgradableModule = (*IBCModule)(nil)
)

// IBCModule implements the ICS26 interface for mockapp given the mockapp keeper.
type IBCModule struct {
	keeper keeper.Keeper
}

// NewIBCModule creates a new IBCModule given the keeper
func NewIBCModule(k keeper.Keeper) IBCModule {
	return IBCModule{
		keeper: k,
	}
}

// ValidateChannelParams does validation of a newly created mockapp channel.
func ValidateChannelParams(
	ctx sdk.Context,
	keeper keeper.Keeper,
	order channeltypes.Order,
	portID string,
	channelID string,
) error {
	// Require portID is the portID mockapp module is bound to
	boundPort := keeper.GetPort(ctx)
	if boundPort != portID {
		return errorsmod.Wrapf(porttypes.ErrInvalidPort, "invalid port: %s, expected %s", portID, boundPort)
	}

	return nil
}

// OnChanOpenInit implements the IBCModule interface
func (im IBCModule) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) (string, error) {
	if err := ValidateChannelParams(ctx, im.keeper, order, portID, channelID); err != nil {
		return "", err
	}

	if strings.TrimSpace(version) == "" {
		version = types.Version
	}

	if version != types.Version {
		return "", errorsmod.Wrapf(types.ErrInvalidVersion, "got %s, expected %s", version, types.Version)
	}

	// Claim channel capability passed back by IBC module
	if err := im.keeper.ClaimCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID)); err != nil {
		return "", err
	}

	return version, nil
}

// OnChanOpenTry implements the IBCModule interface.
func (im IBCModule) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (string, error) {
	if err := ValidateChannelParams(ctx, im.keeper, order, portID, channelID); err != nil {
		return "", err
	}

	if counterpartyVersion != types.Version {
		return "", errorsmod.Wrapf(types.ErrInvalidVersion, "invalid counterparty version: got: %s, expected %s", counterpartyVersion, types.Version)
	}

	// OpenTry must claim the channelCapability that IBC passes into the callback
	if err := im.keeper.ClaimCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID)); err != nil {
		return "", err
	}

	return types.Version, nil
}

// OnChanOpenAck implements the IBCModule interface
func (im IBCModule) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	_ string,
	counterpartyVersion string,
) error {
	if counterpartyVersion != types.Version {
		return errorsmod.Wrapf(types.ErrInvalidVersion, "invalid counterparty version: %s, expected %s", counterpartyVersion, types.Version)
	}
	return nil
}

// OnChanOpenConfirm implements the IBCModule interface
func (im IBCModule) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return nil
}

// OnChanCloseInit implements the IBCModule interface
func (im IBCModule) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// Disallow user-initiated channel closing for mockapp channels
	return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "user cannot close channel")
}

// OnChanCloseConfirm implements the IBCModule interface
func (im IBCModule) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return nil
}

// OnRecvPacket implements the IBCModule interface. A successful acknowledgement
// is returned if the packet data is successfully decoded and the receive application
// logic returns without error.
func (im IBCModule) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	data := packet.GetData()
	if bytes.Equal(data, []byte(MOCK_PACKET_DATA)) {
		return types.NewSuccessfulAcknowledgement(SUCCESSFUL_ACKNOWLEDGEMENT)
	} else if bytes.Equal(data, []byte(MOCK_ASYNC_PACKET_DATA)) {
		return nil
	} else {
		return types.NewFailedAcknowledgement(FAILED_ACKNOWLEDGEMENT)
	}
}

// OnAcknowledgementPacket implements the IBCModule interface
func (im IBCModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	ack []byte,
	relayer sdk.AccAddress,
) error {
	data := packet.GetData()

	if bytes.Equal(data, []byte(MOCK_PACKET_DATA)) {
		expected := types.NewSuccessfulAcknowledgement(SUCCESSFUL_ACKNOWLEDGEMENT).Acknowledgement()
		if !bytes.Equal(ack, expected) {
			return errorsmod.Wrapf(types.ErrUnexpectedAcknowledgement, "got: %s, expected: %s", ack, expected)
		}
	} else if bytes.Equal(data, []byte(MOCK_ASYNC_PACKET_DATA)) {
		expected := types.NewSuccessfulAcknowledgement(SUCCESSFUL_ASYNC_ACKNOWLEDGEMENT).Acknowledgement()
		if !bytes.Equal(ack, expected) {
			return errorsmod.Wrapf(types.ErrUnexpectedAcknowledgement, "got: %s, expected: %s", ack, expected)
		}
	} else {
		expected := types.NewFailedAcknowledgement(FAILED_ACKNOWLEDGEMENT).Acknowledgement()
		if !bytes.Equal(ack, expected) {
			return errorsmod.Wrapf(types.ErrUnexpectedAcknowledgement, "got: %s, expected: %s", ack, expected)
		}
	}

	return nil
}

// OnTimeoutPacket implements the IBCModule interface
func (im IBCModule) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	return nil
}

// OnChanUpgradeInit enables additional custom logic to be executed when the channel upgrade is initialized.
// It must validate the proposed version, order, and connection hops.
// NOTE: in the case of crossing hellos, this callback may be executed on both chains.
// NOTE: Any IBC application state changes made in this callback handler are not committed.
func (IBCModule) OnChanUpgradeInit(
	ctx sdk.Context,
	portID, channelID string,
	proposedOrder channeltypes.Order,
	proposedConnectionHops []string,
	proposedVersion string,
) (string, error) {
	return proposedVersion, nil
}

// OnChanUpgradeTry enables additional custom logic to be executed in the ChannelUpgradeTry step of the
// channel upgrade handshake. It must validate the proposed version (provided by the counterparty), order,
// and connection hops.
// NOTE: Any IBC application state changes made in this callback handler are not committed.
func (IBCModule) OnChanUpgradeTry(
	ctx sdk.Context,
	portID, channelID string,
	proposedOrder channeltypes.Order,
	proposedConnectionHops []string,
	counterpartyVersion string,
) (string, error) {
	return counterpartyVersion, nil
}

// OnChanUpgradeAck enables additional custom logic to be executed in the ChannelUpgradeAck step of the
// channel upgrade handshake. It must validate the version proposed by the counterparty.
// NOTE: Any IBC application state changes made in this callback handler are not committed.
func (IBCModule) OnChanUpgradeAck(
	ctx sdk.Context,
	portID,
	channelID,
	counterpartyVersion string,
) error {
	return nil
}

// OnChanUpgradeOpen enables additional custom logic to be executed when the channel upgrade has successfully completed, and the channel
// has returned to the OPEN state. Any logic associated with changing of the channel fields should be performed
// in this callback.
func (IBCModule) OnChanUpgradeOpen(
	ctx sdk.Context,
	portID,
	channelID string,
	proposedOrder channeltypes.Order,
	proposedConnectionHops []string,
	proposedVersion string,
) {
}
