//this is used by routing table when management with redis

/*
package main

import "fmt"
import "github.com/go-redis/redis"
import "time"


var rclient *redis.Client
var recentTimeKey string

const KEY_FORMAT = "20060102_150405"

func init () {
	 rclient = redis.NewClient(&redis.Options{
                Addr:     "localhost:6379",
                Password: "", // no password set
                DB:       0,  // use default DB
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
        })
	 rclient.FlushDB()
}

func redisPing(rclient *redis.Client) (b bool){
	 _, err = rclient.Ping().Result()
	if err == redis.Nil || if err == nil {
		b = true
	}
	return
}

func setPrefixToRedis(rclient *redis.Client, json string) (result string, err error){
	timeKey := fmt.Sprint(time.Now().Format(KEY_FORMAT))
	recentTimeKey = timeKey
	cmd := redis.NewStringCmd("JSON.SET", timeKey, ".", json, (24 * 30 * time.Hour))
	rclient.Process(cmd)
	result, err := cmd.Result()
	return
}

func getRecentPrefixFromRedis(rclient *redis.Client) (result string, err error){
	result, err := client.Get(recentTimeKey).Result()
	return
}
*/
