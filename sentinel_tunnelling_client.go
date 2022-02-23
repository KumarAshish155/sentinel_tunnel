package main

import (
	// "bufio"
	"encoding/json"
	"fmt"
	"github.com/KumarAshish155/sentinel_tunnel/st_sentinel_connection"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"net"
	"time"
)

type SentinelTunnellingDbConfig struct {
	Name       string
	Local_port string
}

type SentinelTunnellingConfiguration struct {
	Sentinels_addresses_list []string
	Databases                []SentinelTunnellingDbConfig
}

type SentinelTunnellingClient struct {
	configuration       SentinelTunnellingConfiguration
	sentinel_connection *st_sentinel_connection.Sentinel_connection
}

type get_db_address_by_name_function func(db_name string) (string, error)

func NewSentinelTunnellingClient(config_file_location string, l *zap.Logger) *SentinelTunnellingClient {
	data, err := ioutil.ReadFile(config_file_location)
	if err != nil {
		l.Fatal("an error has occur during configuration read",zap.Error(err))
	}

	Tunnelling_client := SentinelTunnellingClient{}
	err = json.Unmarshal(data, &(Tunnelling_client.configuration))
	if err != nil {
		l.Fatal("an error has occur during configuration read,",zap.Error(err))
	}

	Tunnelling_client.sentinel_connection, err =
		st_sentinel_connection.NewSentinelConnection(Tunnelling_client.configuration.Sentinels_addresses_list)
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
	for _, db_conf := range st_client.configuration.Databases {
		go handleSigleDbConnections(db_conf.Local_port, db_conf.Name,
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
	st_client := NewSentinelTunnellingClient("/home/kumar/go/src/github.com/KumarAshish155/sentinel_tunnel/sentinel_tunnel_configuration_example.json",l)
	st_client.Start(l)
	for {
		time.Sleep(1000 * time.Millisecond)
	}
}
