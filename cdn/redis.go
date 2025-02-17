package cdn

import (
	"HypernexCDN/api/data"
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var cts = context.Background()
var rdb *redis.Client

func StartRedisClient(address string, password string, db int) {
	rdb = redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       db,
	})
	_, err := rdb.Ping(cts).Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("Connected to redis!")
}

func GetUserData(id string) *data.UserData {
	result, err := rdb.Get(cts, "user/"+id).Result()
	if err != nil {
		return nil
	}
	var r data.UserData
	if err2 := json.Unmarshal([]byte(result), &r); err2 != nil {
		return nil
	}
	return &r
}
