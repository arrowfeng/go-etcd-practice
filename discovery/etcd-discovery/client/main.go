/*
 * File : main.go
 * Author : arrowfeng
 * Date : 2020/2/26
 */
package main

import (
	"fmt"
	server_discovery "github.com/arrowfeng/etcd-demo/discovery/etcd-discovery/server-discovery"
	"log"
	"time"
)

var endpoints = []string{"localhost:32773", "localhost:32771", "localhost:32769"}
var COMMON_NAME = "services/"

func main() {

	payWatcher, err := server_discovery.NewServiceWatcher(endpoints, COMMON_NAME + "pay")
	if err != nil {
		log.Fatal(err)
	}

	orderWatcher, err := server_discovery.NewServiceWatcher(endpoints, COMMON_NAME + "order")
	if err != nil {
		log.Fatal(err)
	}
	go payWatcher.Watch()
	go orderWatcher.Watch()

	ticker := time.NewTicker(3 * time.Second)
	for {
		select {
			case <- ticker.C:
				fmt.Println(payWatcher.Services)
				fmt.Println(orderWatcher.Services)
		}
	}

}
