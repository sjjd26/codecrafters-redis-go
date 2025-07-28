package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/redis"
	"github.com/codecrafters-io/redis-starter-go/app/redis/redisConfig"
	"github.com/codecrafters-io/redis-starter-go/app/redis/store"
)

var inputChannelQueue = make(chan chan []byte, 10)

func main() {
	port := flag.Int("port", 6379, "Port to listen on")
	dir := flag.String("dir", "/data", "Directory to store data files")
	dbfilename := flag.String("dbfilename", "dump.rdb", "Filename for the database dump")
	replicaOf := flag.String("replicaof", "", "'<Host> <port>' of the master node to replicate from (optional)")
	flag.Parse()

	initConfig(dir, dbfilename, replicaOf, *port)

	redisInstance := redis.NewRedisInstance()
	redisInstance.ListenAndRun(*port)
}

func initConfig(dir, dbfilename, replicaOf *string, port int) {
	config := redisConfig.NewRedisConfig()
	config.Set(redisConfig.ConfigDir, *dir)
	config.Set(redisConfig.ConfigDbFilename, *dbfilename)

	replicationDetails, err := redisConfig.NewReplicationDetails(redisConfig.RoleMaster, port)
	if err != nil {
		panic(err)
	}
	if *replicaOf != "" {
		masterHost := strings.Split(*replicaOf, " ")
		if len(masterHost) != 2 {
			fmt.Println("Invalid replicaOf format. Use '<host> <port>'")
			os.Exit(1)
		}
		masterPort, err := strconv.Atoi(masterHost[1])
		if err != nil {
			fmt.Println("Invalid port number:", masterHost[1])
			os.Exit(1)
		}
		replicationDetails.MasterDetails = &redisConfig.HostDetails{
			Host: masterHost[0],
			Port: masterPort,
		}
		replicationDetails.Role = redisConfig.RoleSlave

		if _, err := replicationDetails.SendHandshake(); err != nil {
			panic(err)
		}
	}
	config.SetReplicationDetails(replicationDetails)

	redisStore := store.NewRedisStore()
	err = redisStore.RdbRestore()
	if err != nil {
		fmt.Println(err)
	}
}
