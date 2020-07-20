package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/Fraunhofer-AISEC/penlog"
	"github.com/spf13/pflag"
)

var logger = penlog.NewLogger("cli", os.Stderr)

type netzteilClient struct {
	client  http.Client
	baseURL *url.URL
}

func recvJSON(r *http.Response, data interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if err := json.Unmarshal(body, &data); err != nil {
		return fmt.Errorf("json decoding error: %s", string(body))
	}
	return nil
}

func (c *netzteilClient) setMaster(device uint, state bool) error {
	var (
		reqPath = fmt.Sprintf("/_netzteil/api/devices/%d/out", device)
		uri     = *c.baseURL
		body    string
	)
	if state {
		body = "true"
	} else {
		body = "false"
	}
	uri.Path = path.Join(uri.Path, reqPath)
	req, err := http.NewRequest(http.MethodPut, uri.String(), strings.NewReader(body))
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	logger.LogDebug(resp)
	return nil
}

func (c *netzteilClient) getMaster(device uint) (bool, error) {
	var (
		reqPath = fmt.Sprintf("/_netzteil/api/devices/%d/out", device)
		uri     = *c.baseURL
	)
	uri.Path = path.Join(uri.Path, reqPath)
	resp, err := c.client.Get(uri.String())
	if err != nil {
		return false, err
	}
	logger.LogDebug(resp)
	var parsedResp bool
	if err := recvJSON(resp, &parsedResp); err != nil {
		return false, err
	}
	return parsedResp, nil
}

func main() {
	var (
		device  = pflag.UintP("device", "d", 0, "device index")
		op      = pflag.StringP("operation", "o", "get", "operation, either 'get' or 'set'")
		opArg   = pflag.StringP("arg", "a", "", "argument for the operation")
		ep      = pflag.StringP("endpoint", "e", "", "endpoint to manipulate")
		verbose = pflag.BoolP("verbose", "v", false, "enable debug log")
	)
	pflag.Parse()

	if !*verbose {
		logger.SetLogLevel(penlog.PrioInfo)
	}

	rawURL := pflag.Arg(0)
	urlParsed, err := url.Parse(rawURL)
	if err != nil {
		logger.LogCritical(err)
		os.Exit(1)
	}
	*op = strings.ToLower(*op)
	if *op != "set" && *op != "get" {
		logger.LogCritical("invalid operation: either 'get' or 'set'")
		os.Exit(1)
	}

	client := netzteilClient{
		client: http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: urlParsed,
	}

	switch *ep {
	case "master":
		switch *op {
		case "get":
			state, err := client.getMaster(*device)
			if err != nil {
				logger.LogCritical(err)
				os.Exit(1)
			}
			fmt.Println(state)
		case "set":
			arg, err := strconv.ParseBool(*opArg)
			if err != nil {
				logger.LogCritical(err)
				os.Exit(1)
			}
			if err := client.setMaster(*device, arg); err != nil {
				logger.LogCritical(err)
				os.Exit(1)
			}
		}
	default:
		logger.LogCritical("endpoint not available")
		os.Exit(1)
	}
}
