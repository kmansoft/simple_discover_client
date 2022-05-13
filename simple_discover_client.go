package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

func fatal(msg string, err error) {
	fmt.Printf("Fatal error %s: %v\n", msg, err)
	os.Exit(1)
}

const (
	contentTypeJson = "application/json; charset=UTF-8"
)

type DiscoverClient struct {
	serverAddress string
	serverPort    int
	client        *http.Client
}

func NewDiscoverClient(serverAddress string, serverPort int) (*DiscoverClient, error) {
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	return &DiscoverClient{
		serverAddress: serverAddress,
		serverPort:    serverPort,
		client:        client,
	}, nil
}

func (client *DiscoverClient) Close() {
}

type rqPut struct {
	Key   string `json:"key"`
	Sub   string `json:"sub"`
	Value string `json:"value"`
}

type rsPut struct {
}

func (client *DiscoverClient) Put(key, sub, value string) error {
	rq := rqPut{
		Key:   key,
		Sub:   sub,
		Value: value,
	}
	rs := rsPut{}

	err := client.rest("put", &rq, &rs)
	if err != nil {
		return err
	}

	return nil
}

type rqGet struct {
	Key string `json:"key"`
}

type rsGetValue struct {
	Sub   string `json:"sub"`
	Value string `json:"value"`
}

type rsGet struct {
	ValueList []rsGetValue `json:"value_list"`
}

type RsGetValue struct {
	Sub   string
	Value string
}

func (client *DiscoverClient) Get(key string) ([]RsGetValue, error) {
	rq := rqGet{
		Key: key,
	}
	rs := rsGet{}

	err := client.rest("get", &rq, &rs)
	if err != nil {
		return nil, err
	}

	list := make([]RsGetValue, 0)
	for _, item := range rs.ValueList {
		list = append(list, RsGetValue{
			Sub:   item.Sub,
			Value: item.Value,
		})
	}

	return list, nil
}

func (client *DiscoverClient) rest(verb string, rq, rs interface{}) error {
	rqData, err := json.Marshal(rq)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://%s:%d/%s", client.serverAddress, client.serverPort, verb)

	resp, err := client.client.Post(url, contentTypeJson, bytes.NewReader(rqData))
	defer func() {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK || resp.Body == nil {
		return errors.New("error receiving response")
	}

	rsData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(rsData, rs)
	if err != nil {
		return err
	}

	return nil
}

/**
 * Flags
 */

type Flags struct {
	serverAddress string
	serverPort    int
}

/**
 * Main
 */

func main() {
	fmt.Printf("Hello this is simple discover client\n")

	// Parse flags
	var flags Flags

	flag.StringVar(&flags.serverAddress, "a", "", "Server address")
	flag.IntVar(&flags.serverPort, "s", 65001, "Server port")
	flag.Parse()

	if flags.serverAddress == "" {
		fmt.Printf("Error: please specify server address with -a address\n")
		os.Exit(1)
	}
	if flags.serverPort <= 0 || flags.serverPort > 65536 {
		fmt.Printf("Error: invalid server port %d\n", flags.serverPort)
		os.Exit(1)
	}

	// Construct the client
	serverAddress := flags.serverAddress
	serverPort := flags.serverPort

	client, err := NewDiscoverClient(serverAddress, serverPort)
	if err != nil {
		fatal("cannot create client", err)
	}
	defer client.Close()

	// Try a few operations
	if err = client.Put("kostya", "year-born", `1972`); err != nil {
		fatal("cannot put value", err)
	}
	if err = client.Put("kostya", "height", `6'2"`); err != nil {
		fatal("cannot put value", err)
	}
	if err = client.Put("kostya", "weight", `too much`); err != nil {
		fatal("cannot put value", err)
	}

	itemList, err := client.Get("kostya")
	if err != nil {
		fatal("cannot get values", err)
	}

	for _, item := range itemList {
		fmt.Printf("Value: %q -> %q\n", item.Sub, item.Value)
	}
}
