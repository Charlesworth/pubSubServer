package redisClient

import (
	"github.com/fzzy/radix/redis"
	"log"
)

type RedisPSClient struct {
	client *redis.Client
}

//test code remove main and package main
//func main() {

//	r := Make("178.62.74.225:6379")
//	r.TopicNew("poo")
//	r.Close()
//}

//Make opens a connection to the Redis instance at ipAdress.
//Remember to defer clientname.close in the main program!
func Make(ipAdress string) *RedisPSClient {
	Redis, err := redis.Dial("tcp", ipAdress)
	errLog(err)

	foo, err := Redis.Cmd("PING").Str()
	errLog(err)
	log.Println("Redis Connection to " + ipAdress + " gives Reply: " + foo + " (connection accepted)")

	r := &RedisPSClient{Redis}
	return r
}

//*************Topic*************************************
func (r *RedisPSClient) TopicNew(topic string) {
	r.client.Cmd("SET", topic, "0")
}

func (r *RedisPSClient) TopicExists(topic string) (exists bool) {
	e, _ := r.client.Cmd("EXISTS", topic).Int()
	if e == 1 {
		return true
	}
	return false
}

func (r *RedisPSClient) TopicGetN(topic string) (postN int) {
	postN, _ = r.client.Cmd("GET", topic).Int()
	return postN
}

func (r *RedisPSClient) TopicIncN(topic string) {
	r.client.Cmd("INCR", topic)
}

//*************Post*************************************
func (r *RedisPSClient) PostNew(post string, content string) { //todo***************************
	foo, err := r.client.Cmd("PING").Str()
	errLog(err)
	log.Println("Redis Connection Reply: " + foo + " (connection accepted)")
}

func (r *RedisPSClient) PostExists(post string) bool {
	e, err := r.client.Cmd("EXISTS", post, "Content").Int()
	errLog(err)
	if e == 1 {
		return true
	}
	return false
}

func (r *RedisPSClient) PostIncUsrCount(post string, incAmount int) {
	_, err := r.client.Cmd("HINCRBY", post, "usrCount", incAmount).Int()
	errLog(err)
}

func (r *RedisPSClient) PostDelete(post string) {
	r.client.Cmd("DEL", post)
}

func (r *RedisPSClient) PostGetContent(post string) string {
	content, err := r.client.Cmd("HGET", post, "Content").Str()
	errLog(err)
	return content
}

func (r *RedisPSClient) PostGetUsrCount(post string) int {
	usrCount, err := r.client.Cmd("HGET", post, "usrCount").Int()
	errLog(err)
	return usrCount
}

//*************User*************************************
func (r *RedisPSClient) UserNew(name string, topic string, postNo int) {
	_, err := r.client.Cmd("HSETNX", name, topic, postNo).Int()
	errLog(err)
}

func (r *RedisPSClient) UserAddTopic(name string, topic string, postNo int) {
	foo, err := r.client.Cmd("PING").Str()
	errLog(err)
	log.Println("Redis Connection Reply: " + foo + " (connection accepted)")
}

func (r *RedisPSClient) UserDelTopic(name string, topic string) {
	foo, err := r.client.Cmd("PING").Str()
	errLog(err)
	log.Println("Redis Connection Reply: " + foo + " (connection accepted)")
}

func (r *RedisPSClient) UserIncPostN(name string, topic string) {
	foo, err := r.client.Cmd("PING").Str()
	errLog(err)
	log.Println("Redis Connection Reply: " + foo + " (connection accepted)")
}

func (r *RedisPSClient) UserGetPostN(name string, topic string) (postNo int) {
	foo, err := r.client.Cmd("PING").Str()
	errLog(err)
	log.Println("Redis Connection Reply: " + foo + " (connection accepted)")
	return 1
}

//*************Util***************************************
func (r *RedisPSClient) Close() {
	r.client.Close()
}

//*************Errors*************************************
func errFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func errLog(err error) {
	if err != nil {
		log.Print(err)
	}
}
