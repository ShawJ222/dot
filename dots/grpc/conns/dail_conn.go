package conns

import (
	"fmt"
	"github.com/scryinfo/dot/dot"
	"github.com/scryinfo/dot/dots/grpc/lb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
)

const (
	DotTypeId = "7bf0a017-ef0c-496a-b04c-b1dc262abc8d"
)

//grpc的连接， 支持一个 Scheme， 下有多个服务，每个服务可以有多个地址（客户端的负载均衡）
type Conns struct {
	conns map[string]*grpc.ClientConn
	//ctx context.Context
	config Config
	lid dot.LiveId
}

func (c *Conns) SetTypeId(tid dot.TypeId, lid dot.LiveId) {
	c.lid = lid
}

func NewDailConnections(conf interface{}) (dot.Dot, error) {
	var err error = nil
	var bs []byte = nil
	if bt, ok := conf.([]byte); ok {
		bs = bt
	} else {
		return nil, dot.SError.Parameter
	}
	dconf := &Config{}
	err = dot.UnMarshalConfig(bs, dconf)
	if err != nil {
		return nil, err
	}

	d := &Conns{
		config: *dconf,
	}

	return d, err
}

func (c *Conns) Create(l dot.Line) error {
	var err error = nil
	sa := make(map[string][]string, len(c.config.Services))
	{
		for i := range c.config.Services {
			s := c.config.Services[i]
			sa[s.Name] = s.Addrs
		}
	}
	resolver.Register(lb.NewClientBuilder(c.config.Scheme, sa))
	c.conns = make(map[string]*grpc.ClientConn, len(c.config.Services))

	for i := range c.config.Services {
		var e1 error = nil
		s := c.config.Services[i]
		c.conns[s.Name], e1 = grpc.Dial(fmt.Sprintf("%s:///%s", c.config.Scheme, s.Name), lb.Balance(s.Balance), grpc.WithInsecure())
		if e1 != nil {
			if err != nil { //log the err
				dot.Logger().Errorln(err.Error())
			}
			err = e1
		}
	}

	return err
}

func (c *Conns) Stop(ignore bool) error {
	var err error = nil
	if len(c.conns) > 0 {
		conns := c.conns
		c.conns = nil
		for _, conn := range conns {
			if conn != nil {
				e1 := conn.Close()
				if e1 != nil { //do not return , close all connection
					if err != nil { //log the err
						dot.Logger().Errorln(err.Error())
					}
					err = e1
				}
			}
		}
	}

	//todo This function is for testing only,
	resolver.UnregisterForTesting(c.config.Scheme)

	return err
}

func (c *Conns) DefaultClientConn() *grpc.ClientConn {
	var conn *grpc.ClientConn = nil
	if len(c.conns) == 1 {
		for _, v := range c.conns {
			conn = v
		}
	}

	return conn
}

func (c *Conns) ClientConn(serviceName string) *grpc.ClientConn {
	var conn *grpc.ClientConn = nil
	if len(c.conns) > 0 {
		if c, ok := c.conns[serviceName]; ok {
			conn = c
		}
	}
	return conn
}

func (c *Conns) ServiceName() []string {
	var sn []string = nil
	if len(c.config.Services) > 0 {
		sn = make([]string, 0, len(c.config.Services))
		for i := range c.config.Services {
			sn = append(sn, c.config.Services[i].Name)
		}
	}
	return sn
}

func (c *Conns) SchemeName() string {
	return c.config.Scheme
}
