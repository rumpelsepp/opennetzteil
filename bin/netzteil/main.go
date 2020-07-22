package main

import (
	"bytes"
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

const (
	operationSET = "set"
	operationGET = "get"
)

var logger = penlog.NewLogger("cli", os.Stderr)

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

type netzteilClient struct {
	client  http.Client
	baseURL *url.URL
}

func (c *netzteilClient) getDeviceList() ([]string, error) {
	var (
		uri        = *c.baseURL
		reqPath    = "/_netzteil/api/devices"
		parsedResp []string
	)
	uri.Path = path.Join(uri.Path, reqPath)
	resp, err := c.client.Get(uri.String())
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http error")
	}
	if err := recvJSON(resp, &parsedResp); err != nil {
		return nil, err
	}
	return parsedResp, nil
}

// TODO: find a way to unify getter/setter

func (c *netzteilClient) setOutParam(device, channel uint, state bool) error {
	var (
		uri     = *c.baseURL
		reqPath string
		body    string
	)
	// Special case for master channel
	if channel == 0 {
		reqPath = fmt.Sprintf("/_netzteil/api/devices/%d/out", device)
	} else {
		reqPath = fmt.Sprintf("/_netzteil/api/devices/%d/channels/%d/out", device, channel)
	}
	uri.Path = path.Join(uri.Path, reqPath)
	if state {
		body = "true"
	} else {
		body = "false"
	}
	req, err := http.NewRequest(http.MethodPut, uri.String(), strings.NewReader(body))
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	logger.LogDebug(resp)
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.LogError(err)
		logger.LogError(resp)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logger.LogErrorf(string(respBody))
	}
	return nil
}

func (c *netzteilClient) getOutParam(device, channel uint) (bool, error) {
	var (
		uri        = *c.baseURL
		reqPath    string
		parsedResp bool
	)
	// Special case for master channel
	if channel == 0 {
		reqPath = fmt.Sprintf("/_netzteil/api/devices/%d/out", device)
	} else {
		reqPath = fmt.Sprintf("/_netzteil/api/devices/%d/channels/%d/out", device, channel)
	}
	uri.Path = path.Join(uri.Path, reqPath)
	resp, err := c.client.Get(uri.String())
	if err != nil {
		return false, err
	}
	logger.LogDebug(resp)
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("http error")
	}
	if err := recvJSON(resp, &parsedResp); err != nil {
		return false, err
	}
	return parsedResp, nil
}

func (c *netzteilClient) getVoltage(device uint, channel uint) (float64, error) {
	var (
		uri        = *c.baseURL
		reqPath    string
		parsedResp float64
	)
	reqPath = fmt.Sprintf("/_netzteil/api/devices/%d/channels/%d/voltage", device, channel)
	uri.Path = path.Join(uri.Path, reqPath)
	resp, err := c.client.Get(uri.String())
	if err != nil {
		return 0, err
	}
	if err := recvJSON(resp, &parsedResp); err != nil {
		return 0, err
	}
	return parsedResp, nil
}

func (c *netzteilClient) setVoltage(device uint, channel uint, voltage float64) error {
	var (
		uri     = *c.baseURL
		reqPath string
		body    []byte
	)
	reqPath = fmt.Sprintf("/_netzteil/api/devices/%d/channels/%d/voltage", device, channel)
	uri.Path = path.Join(uri.Path, reqPath)
	body, _ = json.Marshal(voltage)
	req, err := http.NewRequest(http.MethodPut, uri.String(), bytes.NewReader(body))
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	logger.LogDebug(resp)
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.LogError(err)
		logger.LogError(resp)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logger.LogErrorf(string(respBody))
	}
	return nil
}

func (c *netzteilClient) setCurrent(device uint, channel uint, current float64) error {
	var (
		uri     = *c.baseURL
		reqPath string
		body    []byte
	)
	reqPath = fmt.Sprintf("/_netzteil/api/devices/%d/channels/%d/current", device, channel)
	uri.Path = path.Join(uri.Path, reqPath)
	body, _ = json.Marshal(current)
	req, err := http.NewRequest(http.MethodPut, uri.String(), bytes.NewReader(body))
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	logger.LogDebug(resp)
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.LogError(err)
		logger.LogError(resp)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logger.LogErrorf(string(respBody))
	}
	return nil
}

func (c *netzteilClient) getCurrent(device uint, channel uint) (float64, error) {
	var (
		uri        = *c.baseURL
		reqPath    string
		parsedResp float64
	)
	reqPath = fmt.Sprintf("/_netzteil/api/devices/%d/channels/%d/current", device, channel)
	uri.Path = path.Join(uri.Path, reqPath)
	resp, err := c.client.Get(uri.String())
	if err != nil {
		return 0, err
	}
	if err := recvJSON(resp, &parsedResp); err != nil {
		return 0, err
	}
	return parsedResp, nil
}

func (c *netzteilClient) getChannel(device uint, channel uint) (bool, error) {
	return c.getOutParam(device, channel)
}

func (c *netzteilClient) setChannel(device uint, channel uint, state bool) error {
	return c.setOutParam(device, channel, state)
}

func (c *netzteilClient) getMaster(device uint) (bool, error) {
	return c.getOutParam(device, 0)
}

func (c *netzteilClient) setMaster(device uint, state bool) error {
	return c.setOutParam(device, 0, state)
}

func main() {
	var (
		device  = pflag.UintP("device", "d", 1, "device index")
		channel = pflag.UintP("channel", "c", 1, "channel index")
		op      = pflag.StringP("operation", "o", "get", "operation, either 'get' or 'set'")
		opArg   = pflag.StringP("arg", "a", "", "argument for the operation")
		ep      = pflag.StringP("endpoint", "e", "", "endpoint to manipulate: master, out")
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
	if *op != operationGET && *op != operationSET {
		logger.LogCritical("invalid operation: either 'get' or 'set'")
		os.Exit(1)
	}
	if *ep == "" {
		logger.LogCritical("no endpoint specified")
		os.Exit(1)
	}

	client := netzteilClient{
		client: http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: urlParsed,
	}

	switch *ep {
	case "devices":
		devices, err := client.getDeviceList()
		if err != nil {
			logger.LogCritical(err)
			os.Exit(1)
		}
		for i, device := range devices {
			if device == "" {
				device = "NO DESCRIPTION"
			}
			fmt.Printf("%02d: %s\n", i+1, device)
		}
	case "voltage":
		switch *op {
		case operationGET:
			voltage, err := client.getVoltage(*device, *channel)
			if err != nil {
				logger.LogCritical(err)
				os.Exit(1)
			}
			fmt.Println(voltage)
		case operationSET:
			arg, err := strconv.ParseFloat(*opArg, 32)
			if err != nil {
				logger.LogCritical(err)
				os.Exit(1)
			}
			if err := client.setVoltage(*device, *channel, arg); err != nil {
				logger.LogCritical(err)
				os.Exit(1)
			}
		}
	case "current":
		switch *op {
		case operationGET:
			current, err := client.getCurrent(*device, *channel)
			if err != nil {
				logger.LogCritical(err)
				os.Exit(1)
			}
			fmt.Println(current)
		case operationSET:
			arg, err := strconv.ParseFloat(*opArg, 32)
			if err != nil {
				logger.LogCritical(err)
				os.Exit(1)
			}
			if err := client.setCurrent(*device, *channel, arg); err != nil {
				logger.LogCritical(err)
				os.Exit(1)
			}
		}
	case "out":
		switch *op {
		case operationGET:
			state, err := client.getChannel(*device, *channel)
			if err != nil {
				logger.LogCritical(err)
				os.Exit(1)
			}
			fmt.Println(state)
		case operationSET:
			arg, err := strconv.ParseBool(*opArg)
			if err != nil {
				logger.LogCritical(err)
				os.Exit(1)
			}
			if err := client.setChannel(*device, *channel, arg); err != nil {
				logger.LogCritical(err)
				os.Exit(1)
			}
		}
	case "master":
		switch *op {
		case operationGET:
			state, err := client.getMaster(*device)
			if err != nil {
				logger.LogCritical(err)
				os.Exit(1)
			}
			fmt.Println(state)
		case operationSET:
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
