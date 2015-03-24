package main

import (
	"fmt"
	"github.com/fzzy/radix/redis"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"log"
	"net/http"
)

var RedisClient redis.Client

//TODO abstract exist check
func main() {

	log.Println("               __    ___    ___")
	log.Println("      |__|    |__|  |__    |__")
	log.Println("      |  |TTP |  ub ___|ub ___|erver")
	log.Println("https://github.com/Charlesworth/pubSubServer")
	log.Println("--------------------------------------------")

	//**************    set up redis     ************************
	RedisClient, err := redis.Dial("tcp", "178.62.74.225:6379")
	errLog(err)
	defer RedisClient.Close()

	foo, err := RedisClient.Cmd("PING").Str()
	errLog(err)
	log.Println("Redis Connection Reply: " + foo + " (connection accepted)")

	_, err = RedisClient.Cmd("FLUSHALL").Str() //test code
	errLog(err)                                //test code

	//************** router set up       ************************
	router := httprouter.New()

	router.POST("/:topic/:username", subscribe)
	router.DELETE("/:topic/:username", unsubscribe)
	router.POST("/:topic", publish)
	router.GET("/:topic/:username", retrieve)

	http.Handle("/", router)

	log.Println("Listening...")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func subscribe(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	log.Println("request from " + r.RemoteAddr)
	log.Println("subscribe from " + params.ByName("topic") + "/" + params.ByName("username"))
	//check if usr already exists
	usrExists, _ := RedisClient.Cmd("EXISTS", params.ByName("username")).Int() //******double check exsists works with a hash
	if usrExists == 0 {
		//make the usr profile
		RedisClient.Cmd("HMSET", params.ByName("username"))
	}

	//check if channel exsists
	keyExists, _ := RedisClient.Cmd("EXISTS", params.ByName("topic")).Int()
	if keyExists == 0 {
		//no channel for that name
		log.Println("fail, no such channel exists")
		//**************************************Need code here for fail****************************
	} else {
		//get channel post no
		postNo, _ := RedisClient.Cmd("GET", params.ByName("topic")).Str()

		//add channel to usr profile with channel post no
		RedisClient.Cmd("HSETNX", params.ByName("username"), params.ByName("topic"), postNo)

		//increment usrCount on channel post n
		RedisClient.Cmd("HINCRBY", params.ByName("topic")+postNo, 1) //error possible here, post may not exists yet
		//may need to increment a channel waiting users int

		//return responce code 200
		w.WriteHeader(200)
	}

}

func unsubscribe(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	log.Println("request from " + r.RemoteAddr)
	log.Println("unsubscribe from " + params.ByName("topic") + "/" + params.ByName("username"))
	//check if channel exists
	usrExists, _ := RedisClient.Cmd("EXISTS", params.ByName("username")).Int() //******double check exsists works with a hash
	if usrExists == 0 {
		//profile doesn't exist, fail
		//**************************************Need code here for fail****************************
	}

	//check if channel is in usr profile
	//HEXISTS  yhash field1
	isSubscribed, _ := RedisClient.Cmd("HEXISTS", params.ByName("username"), params.ByName("topic")).Int()
	if isSubscribed == 0 {
		//subscription doesn't exsist
		w.WriteHeader(404)
	} else {
		//if it is get the post no
		postNo, _ := RedisClient.Cmd("HGET", params.ByName("username"), params.ByName("topic")).Str()

		//then remove the channel from usr profile
		RedisClient.Cmd("HDEL", params.ByName("username"), params.ByName("topic"))

		//decrement usr count on that post
		RedisClient.Cmd("HINCRBY", params.ByName("topic")+postNo, "usrCount", -1) //this may fail as post may
		//not exsist yet

		w.WriteHeader(200)
	}

	//*channelClean* go channelClean

}

func publish(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	log.Println("Publish requested from " + r.RemoteAddr + " into Channel " + params.ByName("topic"))
	defer r.Body.Close()
	bodyByte, _ := ioutil.ReadAll(r.Body)
	body := string(bodyByte)
	log.Println(body)

	var postNo string
	//if channel !exists
	keyExists, _ := RedisClient.Cmd("EXISTS", params.ByName("topic")).Int()
	if keyExists == 0 {
		//make the channel with value 1 and use postNumber = 0
		RedisClient.Cmd("SET", params.ByName("topic"), "1")
		postNo = "0"
	} else {
		//else read channel postNumber int and increment the value
		postNo, _ = RedisClient.Cmd("GET", params.ByName("topic")).Str()
		RedisClient.Cmd("INCR", params.ByName("topic"), "1")
	}

	//make channeln hash with content=body and UsrNO=0
	RedisClient.Cmd("HMSET", params.ByName("topic")+postNo, "usrCount", "0", "Content", body)
	//return responce code 200
	w.WriteHeader(200)
}

func retrieve(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	log.Println("request from " + r.RemoteAddr)
	log.Println("retrieve from " + params.ByName("topic") + "/" + params.ByName("username"))
	//check if subscribed
	isSubscribed, _ := RedisClient.Cmd("HEXISTS", params.ByName("username"), params.ByName("topic")).Int()
	if isSubscribed == 0 {
		//subscription doesn't exsist
		w.WriteHeader(404)
	} else {
		//if subscribed then look at channel post no on usr profile
		postNo, _ := RedisClient.Cmd("HGET", params.ByName("username"), params.ByName("topic")).Str()

		//check if there are any posts to read
		postExists, _ := RedisClient.Cmd("HEXISTS", params.ByName("username"), params.ByName("Content")).Int()
		if postExists == 0 {
			//if not then return 204
			w.WriteHeader(204)
		} else {
			//else if channel post no is not last, retrieve post content
			content, _ := RedisClient.Cmd("HGET", params.ByName("username"), params.ByName("Content")).Str()

			//then move usr count to next post (may need to make an empty post for this or make a "waiting users"
			//feild in the channel to carry users accross, in the case where they just read the last post)

			//increment the post number on the user profile
			RedisClient.Cmd("HINCRBY", params.ByName("username"), params.ByName("topic"), 1)

			fmt.Fprintf(w, content)
			w.WriteHeader(200)
		}

		//*cleanChannel*
	}
}

//clean the channel of any posts that have been read by everyone, kicked off after a usrCount decrement
//should run differently depending on retrieve and unsub: unsub will scan whole load until a usrCount is found,
//while retrieve will scan up to channelPost(N + 1)
func channelClean(channel string, postNo int) {
	//after a usr count decrement, clean the channel of any posts that have been read by everyone
}

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
