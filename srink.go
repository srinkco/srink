package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	args := os.Args
	if len(args) == 1 {
		fmt.Println("srink:", "no argument given")
		os.Exit(1)
	}
	switch opt := strings.ToLower(args[1]); opt {
	case "server":
		if len(args) >= 3 && strings.ToLower(args[2]) == "update" {
			updateServerConf(args[3:])
			fmt.Println("srink:", "updated config successfully")
			return
		}
		server := newServer(log.Default())
		server.start()
	default:
		if len(args) >= 2 && strings.ToLower(args[1]) == "update" {
			updateClientConf(args[2:])
			fmt.Println("srink:", "updated config successfully")
			return
		}
		var hash string
		if len(args) >= 3 {
			hash = args[2]
		}
		client := newClient(log.Default())
		client.updateApiUrl("http://0.0.0.0:7837")
		surl, err := client.shortenUrl(hash, opt)
		if err != nil {
			fmt.Println("srink:", err)
			os.Exit(1)
		}
		fmt.Println("srink:", `shortened your url to`, surl)
	}
}

func updateServerConf(args []string) {
	if len(args) <= 1 {
		fmt.Println("srink:", "provide me the key and its value to update")
		os.Exit(1)
	}
	conf := readUserConfig("server-conf.yml", log.Default())
	switch strings.ToLower(args[0]) {
	case "token", "auth-token", "api-key":
		conf.add("token", args[1])
	case "port", "api-port":
		port, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Println("srink:", "port needs to be a valid integer")
			os.Exit(1)
		}
		conf.add("port", port)
	}
	conf.write()
}

func updateClientConf(args []string) {
	if len(args) <= 1 {
		fmt.Println("srink:", "provide me the key and its value to update")
		os.Exit(1)
	}
	conf := readUserConfig("client-conf.yml", log.Default())
	switch strings.ToLower(args[0]) {
	case "token", "auth-token", "api-key":
		conf.add("token", args[1])
	case "api-url", "api", "url":
		conf.add("api-url", args[1])
	}
	conf.write()
}
