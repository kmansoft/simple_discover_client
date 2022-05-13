package simple_discover_client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

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
