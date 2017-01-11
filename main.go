package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type Node struct {
	Certname  string `json:"certname"`
	Ipaddress string `json:"value"`
}

const (
	query = "facts { name='ipaddress' and nodes { facts { name='collectd_version' and value ~ '^5\\\\.[567]' } and resources { type='Class' and title='Collectd' } } }"
	port  = "1234"
	file  = "/tmp/prometheus-targets.yml"
	sleep = 5 * time.Second
)

func main() {
	client := &http.Client{}

	for {
		nodes, err := getNodes(client)
		if err != nil {
			fmt.Println(err)
			break
		}

		err = writeNodes(nodes)
		if err != nil {
			fmt.Println(err)
			break
		}

		fmt.Printf("Sleeping for %v\n", sleep)
		time.Sleep(sleep)
	}
}

func getNodes(client *http.Client) (nodes []Node, err error) {
	form := strings.NewReader(fmt.Sprintf("{\"query\":\"%s\"}", query))
	req, err := http.NewRequest("POST", "http://localhost:8080/pdb/query/v4", form)
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &nodes)
	return
}

func writeNodes(nodes []Node) (err error) {
	var buffer bytes.Buffer

	buffer.WriteString(" - targets:\n")
	for _, node := range nodes {
		buffer.WriteString(fmt.Sprintf("   - %s:%s\n", node.Ipaddress, port))
	}

	fmt.Printf("Writing %v targets to file %s\n", len(nodes), file)
	err = ioutil.WriteFile(file, buffer.Bytes(), 0644)
	return nil
}
