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

//TODO -abstract exist check
//-when you make a publish, the post created should always be one ahead but with no content
//-rewrite all of the usr
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

	router.POST("/:topic/:username", subscribe(RedisClient))
	router.DELETE("/:topic/:username", unsubscribe(RedisClient))
	router.POST("/:topic", publish(RedisClient))
	router.GET("/:topic/:username", retrieve(RedisClient))

	http.Handle("/", router)

	log.Println("Listening...")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func subscribe(RedisClient *redis.Client) func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

		log.Println("request from " + r.RemoteAddr)
		log.Println("subscribe from " + params.ByName("topic") + "/" + params.ByName("username"))

		var postNo int

		//Get the current PostN, if the channel doesn't exist yet make it.
		//check if the channel key has been set in redis
		channelExists, err := RedisClient.Cmd("EXISTS", params.ByName("topic")).Int()
		errLog(err)
		fmt.Println("does channel exist: ", channelExists)
		//if channel !exists
		if channelExists == 0 {
			//make channel:0 and post0:usrCount=0
			RedisClient.Cmd("SET", params.ByName("topic"), "0")
			RedisClient.Cmd("HMSET", params.ByName("topic")+"0", "usrCount", "0")
			postNo = 0
		} else {
			//else read channel:value int
			postNo, _ = RedisClient.Cmd("GET", params.ByName("topic")).Int()
		}

		//add channel to usr profile with channel post no, doesn't over-ride
		//previos subscriptions to the same channel
		writen, err := RedisClient.Cmd("HSETNX", params.ByName("username"), params.ByName("topic"), postNo).Int()
		errLog(err)
		fmt.Println("was it writen (1) or was user already subscribed to this channel (0): ", writen)

		if writen == 1 {
			//increment usrCount on channel post n
			result := RedisClient.Cmd("HINCRBY", params.ByName("topic")+string(postNo), "usrCount", "1")
			fmt.Println("usrCount incremented by: ", result)
		}

		//return responce code 200
		w.WriteHeader(200)
		fmt.Println("-----------------------------------")
	}
}

//test the HDEL and HINCRBY
func unsubscribe(RedisClient *redis.Client) func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		log.Println("request from " + r.RemoteAddr)
		log.Println("unsubscribe from " + params.ByName("topic") + "/" + params.ByName("username"))

		usrExists, _ := RedisClient.Cmd("HEXISTS", params.ByName("username"), params.ByName("topic")).Int()
		if usrExists == 0 {
			w.WriteHeader(404)
		} else {
			postNo, _ := RedisClient.Cmd("HGET", params.ByName("username"), params.ByName("topic")).Int()

			//then remove the channel from usr profile
			RedisClient.Cmd("HDEL", params.ByName("username"), params.ByName("topic"))

			//decrement usr count on that post
			RedisClient.Cmd("HINCRBY", params.ByName("topic")+string(postNo), "usrCount", -1)

			w.WriteHeader(200)

		}
		fmt.Println("-----------------------------------")
		//*unsubChannelClean* go unsubChannelClean

	}
}

//may need to lock the channel:value while doing a publish
func publish(RedisClient *redis.Client) func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

		log.Println("Publish requested from " + r.RemoteAddr + " into Channel " + params.ByName("topic"))
		defer r.Body.Close()
		bodyByte, _ := ioutil.ReadAll(r.Body)
		body := string(bodyByte)
		log.Println("Content: " + body)

		var postNo int

		//Get the current PostN, if the channel doesn't exist yet make it.
		//check if the channel key has been set in redis
		channelExists, err := RedisClient.Cmd("EXISTS", params.ByName("topic")).Int()
		errLog(err)
		fmt.Println("does channel exist: ", channelExists)
		//if channel !exists
		if channelExists == 0 {
			//make channel:0 and post0:usrCount=0
			RedisClient.Cmd("SET", params.ByName("topic"), "0")
			RedisClient.Cmd("HMSET", params.ByName("topic")+"0", "usrCount", "0")
			postNo = 0
		} else {
			//else read channel:value int
			postNo, _ = RedisClient.Cmd("GET", params.ByName("topic")).Int()
		}

		//increment the channel:value
		test := RedisClient.Cmd("INCR", params.ByName("topic"))
		fmt.Println("current post number: ", postNo)
		fmt.Println("next post: ", test)

		//add content to the current post
		result := RedisClient.Cmd("HMSET", params.ByName("topic")+string(postNo), "Content", body)
		fmt.Println("content added to current post: ", result)

		//make the next post with no content
		result = RedisClient.Cmd("HMSET", params.ByName("topic")+string(postNo+1), "usrCount", "0")
		fmt.Println("made a empty next post: ", result)

		//return responce code 200
		w.WriteHeader(200)
		fmt.Println("-----------------------------------")
	}
}

//retrieve
func retrieve(RedisClient *redis.Client) func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

		log.Println("request from " + r.RemoteAddr)
		log.Println("retrieve from " + params.ByName("topic") + "/" + params.ByName("username"))
		//check if subscribed
		isSubscribed, _ := RedisClient.Cmd("HEXISTS", params.ByName("username"), params.ByName("topic")).Int()
		if isSubscribed == 0 {
			//subscription doesn't exsist
			fmt.Println("404: no subscription")
			w.WriteHeader(404)
		} else {
			//if subscribed then look at channel post no on usr profile
			//postNo, err := RedisClient.Cmd("HGET", params.ByName("username"), params.ByName("topic")).Str()
			postNo, err := RedisClient.Cmd("HGET", "charlie", "hello").Int()
			errLog(err)
			fmt.Println("usr is on postN: ", postNo)

			//check if there is any post content to read, they may have subscribed and
			//no one has posted yet
			postHasContent, err := RedisClient.Cmd("HEXISTS", params.ByName("topic")+string(postNo), "Content").Int()
			errLog(err)
			fmt.Println("post number has content: ", postHasContent)
			if postHasContent == 0 {
				fmt.Println("204: no posts for you, your up to date")
				//if not then return 204
				w.WriteHeader(204)
			} else {
				//else if channel post no is not last, retrieve post content
				content, _ := RedisClient.Cmd("HGET", params.ByName("topic")+string(postNo), "Content").Str()
				fmt.Println("content retrieved: ", content)

				//then move usr count to next post
				result := RedisClient.Cmd("HINCRBY", params.ByName("topic")+string(postNo), "usrCount", -1)
				fmt.Println("minus from current post: ", result)
				result = RedisClient.Cmd("HINCRBY", params.ByName("topic")+string(postNo+1), "usrCount", 1)
				fmt.Println("to next post: ", result)

				//increment the post number on the user profile
				result = RedisClient.Cmd("HINCRBY", params.ByName("username"), params.ByName("topic"), 1)
				fmt.Println("increment postN on usr profile: ", result)

				fmt.Fprintf(w, content)
				w.WriteHeader(200)
			}

			//*channelClean* go channelClean
		}
		fmt.Println("-----------------------------------")
	}
}

//clean the channel of any posts that have been read by everyone, kicked off after a usrCount decrement
//should run differently depending on retrieve and unsub: unsub will scan whole load until a usrCount is found,
//while retrieve will scan up to channelPost(N + 1)
func postClean(channel string, postNo int) bool {
	//if post[postNo].usrcount = 0 then
	usrCount, _ := RedisClient.Cmd("HGET", channel+string(postNo), "usrCount").Int()
	if usrCount == 0 {
		//	if post[postNo - 1] !exists
		postHasContent, _ := RedisClient.Cmd("HEXISTS", channel+string(postNo-1), "Content").Int()
		if postHasContent == 0 {
			//delete the post and return true
			RedisClient.Cmd("HDEL", channel+string(postNo), "Content", "usrCount")
			return true
		}
	}
	//else return false
	return false
}

func channelClean(channel string, postNo int) {
	deleted := postClean(channel, postNo)
	if deleted {
		for i := postNo + 1; ; i++ {
			//get the usrcount
			usrCount, _ := RedisClient.Cmd("HGET", channel+string(i), "usrCount").Int()
			if usrCount == 0 {
				//	delete and reloop
				RedisClient.Cmd("HDEL", channel+string(i), "Content", "usrCount")
			} else {
				return
			}
		}
	}
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
