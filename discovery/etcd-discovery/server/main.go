/*
 * File : main.go
 * Author : arrowfeng
 * Date : 2020/2/26
 */
package main

import (
	server_register "github.com/arrowfeng/etcd-demo/discovery/etcd-discovery/server-register"
	"log"
	"strconv"
	"time"
)

var endpoints = []string{"localhost:32773", "localhost:32771", "localhost:32769"}

func main() {

	payGroup, err := server_register.NewServiceGroup("pay", endpoints)
	if err != nil {
		log.Fatal(err)
		return
	}

	orderGroup, err := server_register.NewServiceGroup("order", endpoints)
	if err != nil {
		log.Fatal(err)
		return
	}

	go payGroup.Start()
	go orderGroup.Start()

	time.Sleep(10 * time.Second)
	for i := 0; i < 10; i++ {
		time.Sleep(3 * time.Second)
		go payGroup.AddService(server_register.Service{
			ID: int64(i),
			IP: "192.168.0." + strconv.Itoa(i + 1),
		})

		go orderGroup.AddService(server_register.Service{
			ID: int64(i),
			IP: "192.199.0." + strconv.Itoa(i + 1),
		})
	}

	for i := 0; i < 4; i++ {
		time.Sleep(3 * time.Second)
		go payGroup.Remove(server_register.Service{
			ID: int64(i),
			IP: "192.88.0." + strconv.Itoa(i + 1),
		})

		go orderGroup.Remove(server_register.Service{
			ID: int64(i),
			IP: "192.199.0." + strconv.Itoa(i + 1),
		})
	}

	time.Sleep(10000*time.Second)
}
