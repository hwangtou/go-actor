package etcd

import (
	"context"
	"errors"
	etcClient "go.etcd.io/etcd/clientv3"
	"time"
)

const (
	dialTimeout = 2
)

var (
	ErrKeyExisted = errors.New("key existed")
	ErrValue = errors.New("invalid value")
)

type client struct {
	cli *etcClient.Client
	ctx context.Context
	ctxCancel context.CancelFunc
}

type value struct {
	cli *client
	kv *etcClient.KV
	key string
	resp *etcClient.GetResponse
}

func (m *value) Value() string {
	return string(m.resp.Kvs[0].Value)
}

func (m *value) Update(value string) (err error) {
	kv := etcClient.NewKV(m.cli.cli)
	_, err = kv.Put(m.cli.ctx, m.key, value)
	return
}

func (m *value) Delete() (err error) {
	kv := etcClient.NewKV(m.cli.cli)
	_, err = kv.Delete(m.cli.ctx, m.key)
	return
}

func NewClient(endpoints []string) (cli *client, err error) {
	m := &client{}
	m.cli, err = etcClient.New(etcClient.Config{
		Endpoints:            endpoints,
		DialTimeout:          dialTimeout * time.Second,
	})
	m.ctx, m.ctxCancel = context.WithCancel(context.Background())
	return m, err
}

func (m *client) Create(key, value string) (*value, error) {
	val, err := m.Get(key)
	if err != nil {
		return nil, err
	}
	if val != nil {
		return nil, ErrKeyExisted
	}
	if err := val.Update(value); err != nil {
		return nil, err
	}
	return val, nil
}

func (m *client) Get(key string) (*value, error) {
	kv := etcClient.NewKV(m.cli)
	resp, err := kv.Get(m.ctx, key)
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, ErrValue
	}
	return &value{m, &kv, key, resp}, nil
}

func (m *client) Close() error {
	m.ctxCancel()
	return m.cli.Close()
}
