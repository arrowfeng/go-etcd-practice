/*
 * File : main.go
 * Author : arrowfeng
 * Date : 2020/2/26
 */
package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"log"
	"time"
)

var endpoints = []string{"localhost:32773", "localhost:32771", "localhost:32769"}
var client *clientv3.Client

func main(){
	client, err := clientv3.New(clientv3.Config{
		Endpoints: endpoints,
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Grant(context.Background(), 3)

	client.Put(context.Background(), "zdf", "zdf_value", clientv3.WithLease(resp.ID))

	s, _ := client.KeepAlive(context.Background(), resp.ID)

	for {
		_, ok := <-s
		fmt.Println(ok)
		time.Sleep(time.Second * 5)
	}

}
