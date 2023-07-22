package types

import (
	"fmt"
	"testing"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestMessageUpgrade_ValidateBasic(t *testing.T) {
	signer, err := crypto.GenerateAddress()
	require.NoError(t, err)

	validVersion := "1.0.0"
	validHeight := int64(1)

	msg := MessageUpgrade{
		Signer:  signer,
		Version: validVersion,
		Height:  validHeight,
	}

	err = msg.ValidateBasic()
	require.NoError(t, err)

	// Test missing signer
	msgMissingSigner := proto.Clone(&msg).(*MessageUpgrade)
	msgMissingSigner.Signer = nil
	coreErr := msgMissingSigner.ValidateBasic()
	require.NotNil(t, coreErr)
	require.Equal(t, coreTypes.ErrEmptyAddress().Code(), coreErr.Code())

	// Test invalid signer
	msgInvalidSigner := proto.Clone(&msg).(*MessageUpgrade)
	invalidSigner := "invalid_signer"
	msgInvalidSigner.Signer = []byte(invalidSigner)
	coreErr = msgInvalidSigner.ValidateBasic()
	require.NotNil(t, coreErr)
	require.Equal(t, coreTypes.ErrInvalidAddressLen(crypto.ErrInvalidAddressLen(len(invalidSigner))).Code(), coreErr.Code())

	// Test invalid version
	msgInvalidVersion := proto.Clone(&msg).(*MessageUpgrade)
	msgInvalidVersion.Version = "invalid_version"
	coreErr = msgInvalidVersion.ValidateBasic()
	require.NotNil(t, coreErr)
	require.Equal(t, coreTypes.ErrInvalidProtocolVersion(msgInvalidVersion.Version).Code(), coreErr.Code())

	// Test invalid height
	msgInvalidHeight := proto.Clone(&msg).(*MessageUpgrade)
	msgInvalidHeight.Height = -1
	require.Equal(t, coreTypes.ErrInvalidBlockHeight().Code(), msgInvalidHeight.ValidateBasic().Code())
}

func ExampleMessageUpgrade() {
	// Create a new MessageUpgrade
	msg := &MessageUpgrade{
		Signer:  []byte("da034209758b78eaea06dd99c07909ab54c99b45"),
		Version: "1.2.3",
		Height:  10,
	}

	fmt.Printf("Signer: %s\n", msg.Signer)
	fmt.Printf("Version: %s\n", msg.Version)
	fmt.Printf("Height: %d\n", msg.Height)
	// Output:
	// Signer: da034209758b78eaea06dd99c07909ab54c99b45
	// Version: 1.2.3
	// Height: 10
}
