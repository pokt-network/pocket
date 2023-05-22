package cli

import (
	"context"
	sha "crypto"
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

// TECHDEBT: (unittest) unit test the command: e.g. on number of arguments
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
			Use:   "Relay <servicer> <application> <relayChainID> <payload>",
			Short: "Relay <servicer> <application> <relayChainID> <payload>",
			Long: `Sends a trustless relay using <payload> as contents, to the specified active <servicer> in the the <application>'s session.
Will prompt the user for the *application* account passphrase`,
			Aliases: []string{},
			Args:    cobra.ExactArgs(4),
			RunE: func(cmd *cobra.Command, args []string) error {
				servicerAddr := args[0]
				applicationAddr := args[1]
				chain := args[2]
				relayPayload := args[3]

				// TODO: (SUGGESTION) refactor to decouple the client logic from the CLI/command
				pk, err := getPrivateKey(applicationAddr)
				if err != nil {
					return fmt.Errorf("error getting application's private key: %w", err)
				}

				session, servicer, err := fetchServicer(cmd.Context(), applicationAddr, chain, servicerAddr)
				if err != nil {
					return fmt.Errorf("error getting servicer for the relay: %w", err)
				}

				relay, err := buildRelay(relayPayload, pk, session, servicer)
				if err != nil {
					return fmt.Errorf("error building relay from payload: %w", err)
				}

				fmt.Printf("sending trustless relay for %s to %v with payload: %s\n", applicationAddr, servicer, relayPayload)

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

// TODO: (QUESTION): do we need/want a cli subcommand for fetching servicers?

// fetchServicer returns the servicer specified by the <servicer> argument.
// It validates the following conditions:
//
//	A. The <application> argument is the address of an active application
//	B. The <servicer> is the address of a servicer that is active in the application's current session.
//
// TODO: (SUGGESTION) use a package-internal interface for servicer and application?
// TODO: (SUGGESTION) use a struct as input to combine all fields (same for output)
func fetchServicer(ctx context.Context, appAddress, chain, servicerAddress string) (rpc.Session, rpc.ProtocolActor, error) {
	// TECHDEBT: cache session data
	session, err := getCurrentSession(ctx, appAddress, chain)
	if err != nil {
		return rpc.Session{}, rpc.ProtocolActor{}, fmt.Errorf("Error getting servicer: %w", err)
	}

	var (
		servicer rpc.ProtocolActor
		found    bool
	)
	// TODO: a map may be a better choice for storing servicers
	for _, s := range session.Servicers {
		if s.Address == servicerAddress {
			servicer = s
			found = true
			break
		}
	}

	// TODO: cover with unit tests
	if !found {
		return rpc.Session{}, rpc.ProtocolActor{}, fmt.Errorf("Error getting servicer: address %s does not match any servicers in the session", servicerAddress)
	}

	// TODO: cover with unit tests
	found = false
	for _, ch := range servicer.Chains {
		if ch == chain {
			found = true
			break
		}
	}

	if !found {
		return rpc.Session{}, rpc.ProtocolActor{}, fmt.Errorf("Error getting servicer: service %s does not support chain %s", servicerAddress, chain)
	}

	return *session, servicer, nil
}

func getCurrentSession(ctx context.Context, appAddress, chain string) (*rpc.Session, error) {
	// TODO: passing 0 as the height value to get the current session seems more optimal than this.
	currentHeight, err := getCurrentHeight(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error getting current session: %w", err)
	}

	req := rpc.SessionRequest{
		AppAddress: appAddress,
		Chain:      chain,
		// TODO: Geozone
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
	// TODO: refactor boiler-plate code
	if resp.HTTPResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Error getting current session: Unexpected status code %d for request %v", resp.HTTPResponse.StatusCode, req)
	}

	if resp.JSON200 == nil {
		return nil, fmt.Errorf("Error getting current session: Unexpected response %v", resp)
	}

	return resp.JSON200, nil
}

// TODO: reuse this function in the query commands
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

// TODO: (localnet) Publish Servicer(s) Host and Port as env. vars in K8s: similar to Validators
// TODO: (REFACTOR) should we move package-level variables (e.g. remoteCLIURL) to a cli object?
func sendTrustlessRelay(ctx context.Context, servicerUrl string, relay rpc.RelayRequest) (*rpc.PostV1ClientRelayResponse, error) {
	client, err := rpc.NewClientWithResponses(servicerUrl)
	if err != nil {
		return nil, err
	}

	return client.PostV1ClientRelayWithResponse(ctx, relay)
}

// TODO: (NICE) allow reading the relay request from the command line arguments AND from a file
func buildRelay(payload string, appPrivateKey crypto.PrivateKey, session rpc.Session, servicer rpc.ProtocolActor) (rpc.RelayRequest, error) {
	// TECHDEBT: This is mostly COPIED from pocket-go: we should refactor pocket-go code and import this functionality from there instead.
	relayPayload := rpc.Payload{
		Data:   payload,
		Method: "POST",
		// TODO: Path: load Path field from the corresponding Blockchain (e.g. database)
		// TODO: set Headers
	}

	relayMeta := rpc.RelayRequestMeta{
		BlockHeight: session.SessionHeight,
		// TODO: use Identifiable for Chain in Session (or string for Chain in Relay Meta)
		Chain: rpc.Identifiable{
			Id: session.Chain,
		},
		ServicerPubKey: servicer.PublicKey,
		// TODO: Geozone
		// TODO: Token
	}

	relay := rpc.RelayRequest{
		Payload: relayPayload,
		Meta:    relayMeta,
		// TODO: (QUESTION) why is there no Proof field in v1 struct?
	}
	reqBytes, err := json.Marshal(relay)
	if err != nil {
		return rpc.RelayRequest{}, fmt.Errorf("Error marshalling relay request %v: %w", relay, err)
	}
	hashedReq, err := hash(reqBytes)
	if err != nil {
		return rpc.RelayRequest{}, fmt.Errorf("Error hashing relay request bytes %s: %w", string(reqBytes), err)
	}
	signature, err := appPrivateKey.Sign(hashedReq)
	if err != nil {
		return relay, fmt.Errorf("Error signing relay: %w", err)
	}
	relay.Meta.Signature = hex.EncodeToString(signature)

	return relay, nil
}

func hash(data []byte) ([]byte, error) {
	hasher := sha.SHA3_256.New()
	if _, err := hasher.Write(data); err != nil {
		return nil, fmt.Errorf("Error hashing data: %w", err)
	}

	return hasher.Sum(nil), nil
}

// TODO: remove use of package-level variables
func getPrivateKey(address string) (crypto.PrivateKey, error) {
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
