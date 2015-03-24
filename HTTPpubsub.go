package main

import (
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

func main() {

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
	//
}

func unsubscribe(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	log.Println("request from " + r.RemoteAddr)
	log.Println("unsubscribe from " + params.ByName("topic") + "/" + params.ByName("username"))
}

func publish(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	log.Println("request from " + r.RemoteAddr)
	log.Println("publish into " + params.ByName("topic"))
}

func retrieve(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	log.Println("request from " + r.RemoteAddr)
	log.Println("retrieve from " + params.ByName("topic") + "/" + params.ByName("username"))
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
