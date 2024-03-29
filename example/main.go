package main

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/txodds/twunproxy"
	"strings"
	"time"
)

// The path to the Twemproxy configuration file and the name of the pool from which to get connections.
const (
	confPath string = "./nutcracker.yml"
	poolName string = "alpha"
)

// TwunPool is a wrapper for the Redigo Redis pool; it satisfies the twunproxy.ConnGetter interface.
// In this way, Twunproxy is agnostic with regard to the Redis client library. Just wrap whatever you're using like this.
type twunPool struct {
	wrapped *redis.Pool
}

func (p *twunPool) Get() twunproxy.Conn {
	return p.wrapped.Get()
}

// Instantiates connection pools based on the entries in the Twemproxy configuration file.
// A different method could also be defined to instantiate pools for entries describing Unix domain sockets.
var getTwunPool twunproxy.CreatePool = func(desc string, auth string) twunproxy.ConnGetter {
	tok := strings.Split(strings.Split(desc, " ")[0], ":")
	fmt.Println("Creating pool.")
	return &twunPool{wrapped: newPool(strings.Join(tok[:2], ":"), auth)}
}

// From: https://godoc.org/github.com/garyburd/redigo/redis#Pool
// MaxIdle for this pool is 0. This prevents any persistent connections.
func newPool(server, password string) *redis.Pool {
	return &redis.Pool{
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			/*
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			*/
			return c, err
		},
	}
}

// Instantiates a Twunproxy connection based on our Twemproxy configuration file and BLPOPs a list indefinitely.
func main() {
	proxy, err := twunproxy.NewProxyConn(confPath, poolName, 0, getTwunPool)
	if err != nil {
		panic(err)
	}

	fmt.Println("Waiting for list items...")

	for {
		if v, err := proxy.BLPop("test:list", 20*time.Second); err == nil {
			fmt.Println(v)
		} else {
			panic(err)
		}
	}
}
