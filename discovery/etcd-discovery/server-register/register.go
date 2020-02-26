/*
 * File : register.go
 * Author : arrowfeng
 * Date : 2020/2/26
 */

// 服务注册
package server_register

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"log"
	"sync"
	"time"
)

const (
	CHANGEBUFFERSIZE = 10 //状态改变的峰值
)

const (
	Added DeltaType = "Added"
	Deleted DeltaType = "Deleted"
)

type DeltaType string

type Service struct {
	ID int64
	IP string
}

type ServiceSet map[Service]struct{}
type empty struct {}

func (set ServiceSet) has(service Service) bool {
	 _, ok := set[service]
	return ok
}

func (set ServiceSet) insert(service Service) {
	set[service] = empty{}
}

func (set ServiceSet) delete(service Service) {
	delete(set, service)
}

type Delta struct {
	Type DeltaType
	Service Service
}


// server-register
type ServiceGroup struct {
	sync.Mutex
	services ServiceSet //去重
	GroupName string
	stopCh chan struct{} // 暂停信号
	changeCh chan Delta // 状态改变信号
	leaseid clientv3.LeaseID
	client *clientv3.Client
}


func NewServiceGroup(name string, endpoints []string) (*ServiceGroup, error){

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:            endpoints,
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		log.Fatal(err)
	}

	return &ServiceGroup{
		services:make(ServiceSet),
		GroupName: name,
		stopCh:     make(chan struct{}),
		changeCh:   make(chan Delta, CHANGEBUFFERSIZE),
		client:   cli,
	}, nil

}

// 提供服务信息， etcd 端点进行注册
func (s *ServiceGroup)AddService(service Service) {
	s.Lock()
	defer s.Unlock()
	s.services.insert(service)
	s.changeCh<- Delta{Added, service}
}


func (s *ServiceGroup)refreshService(delta Delta) {

	key := fmt.Sprintf("services/%s/%d",s.GroupName, delta.Service.ID)
	value, err := json.Marshal(delta.Service)
	if err != nil {
		log.Fatal(err)
		return
	}

	switch delta.Type {
	case Added:
		s.client.Put(context.Background(), key, string(value), clientv3.WithLease(s.leaseid))
	case Deleted:
		s.client.Delete(context.Background(), key)
	}
}

// 开始服务注册
func (s *ServiceGroup) Start() error {
	ch, err := s.keepAlive()
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case <-s.stopCh:
			s.revoke()
			return nil
		case delta := <-s.changeCh:
			go s.refreshService(delta)
		case <-s.client.Ctx().Done():
			return errors.New("server-register closed")
		case ka, ok := <-ch:
			if !ok {
				log.Println("keep alive channel closed")
				s.revoke()
				return nil
			} else {
				log.Printf("Recv reply from services: %s, ttl:%d", s.GroupName, ka.TTL)
			}
		}
	}

}


// 从服务组中移除某个服务
func (s *ServiceGroup) Remove(service Service) {
	s.Lock()
	defer s.Unlock()
	if s.services.has(service) {
		s.services.delete(service)
		s.changeCh <- Delta{Deleted, service}
	}
}

// 停止服务组
func (s *ServiceGroup) Stop() {
	close(s.stopCh)
}

// 吊销服务组 lease
func (s *ServiceGroup) revoke() error {
	_, err := s.client.Revoke(context.Background(), s.leaseid)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("servide:%s stop\n", s.GroupName)
	return err
}


// 续租
func (s *ServiceGroup) keepAlive() (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	// 获得 lease，同一组服务的 lease 相同
	resp, err := s.client.Grant(context.Background(), 5)
	if err != nil {
		log.Fatal(err)
	}

	for v, _ := range s.services {
		key := fmt.Sprintf("services/%s/%d",s.GroupName, v.ID)
		value, err := json.Marshal(v)
		if err != nil {
			log.Fatal(err)
		}

		_, err = s.client.Put(context.Background(), key, string(value), clientv3.WithLease(resp.ID))
		if err != nil {
			log.Fatal(err)
		}
	}
	s.leaseid = resp.ID
	return s.client.KeepAlive(context.Background(), resp.ID)
}