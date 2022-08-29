package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/pokt-network/pocket/app/client/rpc"
	"github.com/pokt-network/pocket/shared/types/genesis/test_artifacts"
)

func QueryRPC(routeKey rpc.RouteKey, jsonArgs []byte) (string, error) {
	route, ok := rpc.RoutesMap[routeKey]
	if !ok {
		log.Fatalf("unable to find route with key %s", routeKey)
	}
	cliURL := remoteCLIURL + route.Path
	fmt.Printf("%s %s", route.Method, cliURL)
	req, err := http.NewRequest(route.Method, cliURL, bytes.NewBuffer(jsonArgs))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{
		Timeout: time.Duration(test_artifacts.DefaultRpcTimeout) * time.Millisecond,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bz, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	res, err := strconv.Unquote(string(bz))
	if err == nil {
		bz = []byte(res)
	}
	if resp.StatusCode == http.StatusOK {
		var prettyJSON bytes.Buffer
		err = json.Indent(&prettyJSON, bz, "", "    ")
		if err == nil {
			return prettyJSON.String(), nil
		}
		return string(bz), nil
	}
	return "", fmt.Errorf("the http status code was not okay: %d, and the status was: %s, with a response of %v", resp.StatusCode, resp.Status, string(bz))
}
