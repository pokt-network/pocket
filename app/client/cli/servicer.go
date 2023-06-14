package cli

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/pokt-network/pocket/rpc"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/crypto"
)

func init() {
	rootCmd.AddCommand(NewServicerCommand())
}

func NewServicerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "Servicer",
		Short:   "Servicer specific commands",
		Aliases: []string{"servicer"},
		Args:    cobra.ExactArgs(0),
	}

	cmds := servicerCommands()
	applySubcommandOptions(cmds, attachPwdFlagToSubcommands())
	applySubcommandOptions(cmds, attachKeybaseFlagsToSubcommands())

	cmd.AddCommand(cmds...)

	return cmd
}

func servicerCommands() []*cobra.Command {
	cmdDef := actorCmdDef{"Servicer", coreTypes.ActorType_ACTOR_TYPE_SERVICER}
	cmds := []*cobra.Command{
		newStakeCmd(cmdDef),
		newEditStakeCmd(cmdDef),
		newUnstakeCmd(cmdDef),
		newUnpauseCmd(cmdDef),
		{
			// IMPROVE: allow reading the relay payload from a file with the serialized protobuf via [--input_file]
			Use:   "Relay <applicationAddrHex> <servicerAddrHex> <relayChainID> <relayPayload>",
			Short: "Relay <applicationAddrHex> <servicerAddrHex> <relayChainID> <relayPayload>",
			Long: `Sends a trustless relay using <relayPayload> as contents, to the specified active <servicerAddrHex> in the the <applicationAddrHex>'s session.
Will prompt the user for the *application* account passphrase`,
			Aliases: []string{},
			Args:    cobra.ExactArgs(4),
			RunE: func(cmd *cobra.Command, args []string) error {
				applicationAddr := args[0]
				servicerAddr := args[1]
				chain := args[2]
				relayPayload := args[3]

				// REFACTOR: decouple the client logic from the CLI
				//	The client will: send the trustless relay and return the response (using a single function as entrypoint)
				//	The CLI will:
				//		1) extract the required input from the command arguments
				//		2) call the client function (with the inputs above) that performs the trustless relay
				pk, err := getPrivateKeyFromKeybase(applicationAddr)
				if err != nil {
					return fmt.Errorf("error getting application's private key: %w", err)
				}

				// TECHDEBT(#791): cache session data
				session, err := getCurrentSession(cmd.Context(), applicationAddr, chain)
				if err != nil {
					return fmt.Errorf("Error getting current session: %w", err)
				}

				servicer, err := validateServicer(cmd.Context(), session, servicerAddr)
				if err != nil {
					return fmt.Errorf("error getting servicer for the relay: %w", err)
				}

				relay, err := buildRelay(relayPayload, pk, session, servicer)
				if err != nil {
					return fmt.Errorf("error building relay from payload: %w", err)
				}

				fmt.Printf("sending trustless relay for %s to %s with payload: %s\n", applicationAddr, servicerAddr, relayPayload)

				resp, err := sendTrustlessRelay(cmd.Context(), servicer.ServiceUrl, relay)
				if err != nil {
					return err
				}

				fmt.Printf("HTTP status code: %d\n", resp.HTTPResponse.StatusCode)
				fmt.Println("Response: ", resp.JSON200)

				return nil
			},
		},
	}

	return cmds
}

// TODO: add a cli command for fetching sessions
// validateServicer returns the servicer specified by the <servicer> argument.
// It validates that the <servicer> is the address of a servicer that is active in the current session.
func validateServicer(ctx context.Context, session *rpc.Session, servicerAddress string) (*rpc.ProtocolActor, error) {
	for _, s := range session.Servicers {
		if s.Address == servicerAddress {
			return &s, nil
		}
	}

	// ADDTEST: cover with gherkin tests
	return nil, fmt.Errorf("Error getting servicer: address %s does not match any servicers in the session %d", servicerAddress, session.SessionNumber)
}

func getCurrentSession(ctx context.Context, appAddress, chain string) (*rpc.Session, error) {
	// CONSIDERATION: passing 0 as the height value to get the current session seems more optimal than this.
	currentHeight, err := getCurrentHeight(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error getting current session: %w", err)
	}

	req := rpc.SessionRequest{
		AppAddress: appAddress,
		Chain:      chain,
		// TODO(#697): Geozone
		SessionHeight: currentHeight,
	}

	client, err := rpc.NewClientWithResponses(remoteCLIURL)
	if err != nil {
		return nil, fmt.Errorf("Error getting current session for app/chain/height: %s/%s/%d: %w", appAddress, chain, currentHeight, err)
	}

	resp, err := client.PostV1ClientGetSessionWithResponse(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("Error getting current session with request %v: %w", req, err)
	}

	// IMPROVE: move the HTTP response processing code to a separate function to enable reuse.
	if resp.HTTPResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Error getting current session: Unexpected status code %d for request %v", resp.HTTPResponse.StatusCode, req)
	}

	if resp.JSON200 == nil {
		return nil, fmt.Errorf("Error getting current session: Unexpected response %v", resp)
	}

	return resp.JSON200, nil
}

// REFACTOR: reuse this function in all the query commands
func getCurrentHeight(ctx context.Context) (int64, error) {
	client, err := rpc.NewClientWithResponses(remoteCLIURL)
	if err != nil {
		return 0, fmt.Errorf("Error getting current height: %w", err)
	}
	resp, err := client.GetV1QueryHeightWithResponse(ctx)
	if err != nil {
		return 0, fmt.Errorf("Error getting current height: %w", err)
	}
	if resp.HTTPResponse.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("Error getting current height: Unexpected status code %d", resp.HTTPResponse.StatusCode)
	}

	if resp.JSON200 == nil {
		return 0, fmt.Errorf("Error getting current height: Unexpected response %v", resp)
	}

	return resp.JSON200.Height, nil
}

// IMPROVE: [K8s][LocalNet] Publish Servicer(s) Host and Port as env. vars in K8s: similar to Validators
// REFACTOR: move package-level variables (e.g. remoteCLIURL) to a cli object and consider storing it in the context
func sendTrustlessRelay(ctx context.Context, servicerUrl string, relay *rpc.RelayRequest) (*rpc.PostV1ClientRelayResponse, error) {
	client, err := rpc.NewClientWithResponses(servicerUrl)
	if err != nil {
		return nil, err
	}

	return client.PostV1ClientRelayWithResponse(ctx, *relay)
}

func buildRelay(payload string, appPrivateKey crypto.PrivateKey, session *rpc.Session, servicer *rpc.ProtocolActor) (*rpc.RelayRequest, error) {
	// TECHDEBT: This is mostly COPIED from pocket-go: we should refactor pocket-go code and import this functionality from there instead.
	relayPayload := rpc.Payload{
		// INCOMPLETE(#803): need to unmarshal into JSONRPC and other supported relay formats once proto-generated custom types are added.
		Jsonrpc: "2.0",
		Method:  payload,
		// TODO: Path: load Path field from the corresponding Blockchain (e.g. database)
		// TODO: set Headers
	}

	relayMeta := rpc.RelayRequestMeta{
		BlockHeight: session.SessionHeight,
		// TODO: Make Chain Identifier type consistent in Session and Meta use Identifiable for Chain in Session (or string for Chain in Relay Meta)
		Chain: rpc.Identifiable{
			Id: session.Chain,
		},
		ServicerPubKey: servicer.PublicKey,
		// TODO(#697): Geozone
	}

	relay := &rpc.RelayRequest{
		Payload: relayPayload,
		Meta:    relayMeta,
	}
	// TECHDEBT: Evaluate which fields we should and shouldn't marshall when signing the payload
	reqBytes, err := json.Marshal(relay)
	if err != nil {
		return nil, fmt.Errorf("Error marshalling relay request %v: %w", relay, err)
	}
	hashedReq := crypto.SHA3Hash(reqBytes)
	signature, err := appPrivateKey.Sign(hashedReq)
	if err != nil {
		return relay, fmt.Errorf("Error signing relay: %w", err)
	}
	relay.Meta.Signature = hex.EncodeToString(signature)

	return relay, nil
}

// TECHDEBT: remove use of package-level variables, such as nonInteractive, remoteCLIURL, pwd, etc.
func getPrivateKeyFromKeybase(address string) (crypto.PrivateKey, error) {
	kb, err := keybaseForCLI()
	if err != nil {
		return nil, err
	}

	if !nonInteractive {
		pwd = readPassphrase(pwd)
	}

	pk, err := kb.GetPrivKey(address, pwd)
	if err != nil {
		return nil, err
	}
	if err := kb.Stop(); err != nil {
		return nil, err
	}

	return pk, nil
}
