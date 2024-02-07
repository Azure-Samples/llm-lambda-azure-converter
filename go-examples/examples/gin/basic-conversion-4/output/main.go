package main

import (
	"math/rand"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

const (
	EnvVarAzureFunctionPort = "FUNCTIONS_PORT"
)

var jokes []Joke

type Joke struct {
	ID   int    `json:"id"`
	Joke string `json:"joke"`
}

func init() {
	jokes = []Joke{
		{ID: 1, Joke: "Why don't scientists trust atoms? Because they make up everything!"},
		{ID: 2, Joke: "Did you hear about the mathematician who's afraid of negative numbers? He will stop at nothing to avoid them!"},
		{ID: 3, Joke: "Why don't skeletons fight each other? They don't have the guts!"},
		{ID: 4, Joke: "Why did the chicken go to the séance? To get to the other side!"},
		{ID: 5, Joke: "What do you call a fake noodle? An impasta!"},
		{ID: 6, Joke: "What did the grape do when he got stepped on? He let out a little wine!"},
		{ID: 7, Joke: "I wouldn't buy anything with velcro. It's a total rip-off."},
		{ID: 8, Joke: "The shovel was a ground-breaking invention."},
		{ID: 9, Joke: "Dad, can you put my shoes on? No, I don't think they'll fit me."},
		{ID: 10, Joke: "Did you hear about the restaurant on the moon? Great food, no atmosphere."},
	}
}

func JokeHandler(ctx *gin.Context) {
	joke := getRandomJoke()
	ctx.JSON(http.StatusOK, &joke)
}

func getRandomJoke() *Joke {
	index := rand.Intn(len(jokes))
	return &jokes[index]
}

func main() {
	r := gin.Default()
	r.Handle(http.MethodPost, "/JokeHandler", JokeHandler)

	port := os.Getenv(EnvVarAzureFunctionPort)
	if port == "" {
		port = "8080"
	}
	host := fmt.Sprintf("0.0.0.0:%s", port)
	fmt.Printf("Go server Listening...on port: %s\n", port)
	log.Fatal(r.Run(host))
}
