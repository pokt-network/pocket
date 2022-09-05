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

	"github.com/pokt-network/pocket/shared/types/genesis/test_artifacts"
)

func QueryRPC(path string, jsonArgs []byte) (string, error) {
	// preparing the request
	cliURL := remoteCLIURL + path
	log.Println(cliURL)
	req, err := http.NewRequest(http.MethodPost, cliURL, bytes.NewBuffer(jsonArgs))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{
		Timeout: time.Duration(test_artifacts.DefaultRpcTimeout) * time.Millisecond,
	}

	// executing the request
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// reading the response
	bz, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	res, err := strconv.Unquote(string(bz))
	if err != nil {
		return "", err
	}
	bz = []byte(res)

	// returning OK / KO
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
