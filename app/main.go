package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/redis"
	"github.com/codecrafters-io/redis-starter-go/app/redis/redisConfig"
)

var inputChannelQueue = make(chan chan []byte, 10)

func main() {
	port := flag.Int("port", 6379, "Port to listen on")
	dir := flag.String("dir", "", "Directory to store data files")
	dbfilename := flag.String("dbfilename", "", "Filename for the database dump")
	replicaOf := flag.String("replicaof", "", "'<Host> <port>' of the master node to replicate from (optional)")
	flag.Parse()

	selfDetails, masterDetails := CreateHostDetails(*port, *replicaOf)
	redisInstance, err := redis.NewRedisInstance(selfDetails, masterDetails)
	if err != nil {
		panic(err)
	}

	SetRdb(*dir, *dbfilename, redisInstance)

	redisInstance.ListenAndRun()
}

func SetRdb(dir, dbfilename string, redisInstance redis.RedisInstance) {
	if dir != "" && dbfilename != "" {
		redisInstance.GetConfig().Set(redisConfig.ConfigDir, dir)
		redisInstance.GetConfig().Set(redisConfig.ConfigDbFilename, dbfilename)
		redis.RestoreFromRdb(redisInstance)
	}
}

func CreateHostDetails(selfPort int, replicaOf string) (*redisConfig.HostDetails, *redisConfig.HostDetails) {
	selfDetails := &redisConfig.HostDetails{
		Host: "localhost",
		Port: selfPort,
	}
	var masterDetails *redisConfig.HostDetails = nil

	if replicaOf != "" {
		parts := strings.Split(replicaOf, " ")
		if len(parts) != 2 {
			fmt.Println("Invalid replicaOf format. Use '<host> <port>'")
			os.Exit(1)
		}
		masterPort, err := strconv.Atoi(parts[1])
		if err != nil {
			fmt.Println("Invalid port number:", parts[1])
			os.Exit(1)
		}
		masterDetails = &redisConfig.HostDetails{
			Host: parts[0],
			Port: masterPort,
		}
	}

	return selfDetails, masterDetails
}
