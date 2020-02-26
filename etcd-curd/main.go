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
	"time"
)

var endpoints = []string{"localhost:32773", "localhost:32771", "localhost:32769"}
var client *clientv3.Client

func main() {
	//clientv3.SetLogger(grpclog.NewLoggerV2(os.Stderr, os.Stderr, os.Stderr))
	client, err := clientv3.New(clientv3.Config{
		Endpoints:            endpoints,
		DialTimeout: 5 * time.Second,
	})

	if err != nil {

	}

	defer client.Close()


	ctx, cancel := context.WithTimeout(context.Background(), 2 * time.Second)
	_, err = client.Put(ctx, "sample_key", "sample_value")
	if err != nil {

	}

	resp1, err := client.Get(ctx, "sample_key")
	if err != nil {

	}
	cancel()

	fmt.Println(resp1.Kvs)

}
