/*
 * File : discovery.go
 * Author : arrowfeng
 * Date : 2020/2/26
 */

// 服务发现
package server_discovery

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/coreos/etcd/clientv3"
	"log"
	"strings"
	"sync"
	"time"
)

const (
	CHANGEBUFFERSIZE = 10 //状态改变的峰值
)

const (
	Puted DeltaType = "Puted"
	Deleted DeltaType = "Deleted"
)

type DeltaType string

type Delta struct {
	Type DeltaType
	*clientv3.Event
}


type ServiceWatcher struct {
	sync.Mutex
	Path string // 监听的路径
	Services Services
	Client *clientv3.Client
	stopCh chan struct{}
	changeCh chan Delta
}

type Services map[string]*Service

type Service struct {
	FullName string
	ID int64
	IP string
}

func NewServiceWatcher(endpoints []string, watchPath string) (*ServiceWatcher, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:            endpoints,
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	watcher := &ServiceWatcher{
		Path:   watchPath,
		Services:  make(map[string]*Service),
		stopCh: make(chan struct{}),
		changeCh: make(chan Delta, CHANGEBUFFERSIZE),
		Client: cli,
	}

	go watcher.Watch()

	return watcher, nil
}

func (s *ServiceWatcher) add(event *clientv3.Event) {
	service := &Service{FullName:strings.Replace(string(event.Kv.Key), "/", "-", -1)}
	err := json.Unmarshal(event.Kv.Value, &service)
	if err != nil {
		log.Fatal(err)
	}
	s.Lock()
	defer s.Unlock()
	s.Services[service.FullName] = service
}

func (s *ServiceWatcher) del(event *clientv3.Event) {
	s.Lock()
	defer s.Unlock()
	delete(s.Services, strings.Replace(string(event.Kv.Key), "/", "-", -1))
}

func (s *ServiceWatcher) Stop() {
	close(s.stopCh)
}

// 监听节点事件
func (s *ServiceWatcher) watch() {
	for {
		select {
			case delta := <- s.changeCh:
				switch delta.Type {
				case Puted:
					s.add(delta.Event)
				case Deleted:
					s.del(delta.Event)
				}
			case <- s.stopCh:
				return
		}
	}
}

// 监听 etcd 服务变更信号
func (s *ServiceWatcher) Watch() {
	wch := s.Client.Watch(context.Background(), s.Path, clientv3.WithPrefix())

	go s.watch()

	for {
		select {
			case events := <-wch :
			for _, ev := range events.Events {
				switch ev.Type {
					case clientv3.EventTypePut:
						s.changeCh <- Delta{Type:  Puted, Event: ev}
					case clientv3.EventTypeDelete:
						s.changeCh <- Delta{Type: Deleted, Event: ev}
				}
			}
			case <-s.stopCh:
				return
		}
	}
}

func (s Services) String() string {

	buffer := bytes.NewBufferString("service:    ")

	for _, v := range s {
		buffer.WriteString(v.FullName)
		buffer.WriteString(":")
		buffer.WriteString(v.IP)
		buffer.WriteString("            ")
	}

	return buffer.String()
}