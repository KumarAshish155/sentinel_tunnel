package main

import (
	"fmt"
	"github.com/KumarAshish155/sentinel_tunnel/config"
	"github.com/KumarAshish155/sentinel_tunnel/st_sentinel_connection"
	"go.uber.org/zap"
	"io"
	"net"
	"time"
)


type SentinelTunnellingClient struct {
	configuration       config.Configuration
	sentinel_connection *st_sentinel_connection.Sentinel_connection
}

type get_db_address_by_name_function func(db_name string) (string, error)

func NewSentinelTunnellingClient(cfg config.Configuration, l *zap.Logger) *SentinelTunnellingClient {
	Tunnelling_client := SentinelTunnellingClient{}
	Tunnelling_client.configuration= cfg

	conn,err:=st_sentinel_connection.NewSentinelConnection(Tunnelling_client.configuration.SentinelAddress)
	Tunnelling_client.sentinel_connection=conn
	if err != nil {
		l.Fatal("an error has occur, ",zap.Error(err))
	}

	l.Info( "done initializing Tunnelling")

	return &Tunnelling_client
}

func createTunnelling(conn1 net.Conn, conn2 net.Conn) {
	io.Copy(conn1, conn2)
	conn1.Close()
	conn2.Close()
}

func handleConnection(c net.Conn, db_name string, get_db_address_by_name get_db_address_by_name_function, l *zap.Logger) {
	db_address, err := get_db_address_by_name(db_name)
	if err != nil {
		l.Error("cannot get db address for "+ db_name,zap.Error(err))
		c.Close()
		return
	}
	db_conn, err := net.Dial("tcp", db_address)
	if err != nil {
		l.Error("cannot connect to db "+ db_name,zap.Error(err))
		c.Close()
		return
	}
	go createTunnelling(c, db_conn)
	go createTunnelling(db_conn, c)
}

func handleSigleDbConnections(listening_port string, db_name string, get_db_address_by_name get_db_address_by_name_function, l *zap.Logger) {

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", listening_port))
	if err != nil {
		l.Fatal("cannot listen to port "+ listening_port, zap.Error(err))
	}

	l.Info("listening on port "+ listening_port+" for connections to database: "+db_name)
	for {
		conn, err := listener.Accept()
		if err != nil {
			l.Fatal("cannot accept connections on port "+listening_port,zap.Error(err))
		}
		go handleConnection(conn, db_name, get_db_address_by_name,l)
	}

}

func (st_client *SentinelTunnellingClient) Start( l *zap.Logger) {
	for _, db_conf := range st_client.configuration.DB {
		go handleSigleDbConnections(db_conf.Port, db_conf.Name,
			st_client.sentinel_connection.GetAddressByDbName,l)
	}
}

func main() {
	l,_  := zap.NewProduction()
	defer func() {
		if err := l.Sync(); err != nil {
			l.Fatal("couldn't flush zap logger", zap.Error(err))
		}
	}()
	cfg := config.Init();
	fmt.Println(cfg)
	st_client := NewSentinelTunnellingClient(cfg,l)
	st_client.Start(l)
	for {
		time.Sleep(1000 * time.Millisecond)
	}
}
