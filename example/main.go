package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/innovafour/ghttp"
)

type Logger struct{}

func (l Logger) Debug(v ...interface{}) {
	fmt.Println(v...)
}
func (l Logger) Error(v ...interface{}) {
	fmt.Println(v...)
}
func (l Logger) Info(v ...interface{}) {
	fmt.Println(v...)
}

func main() {
	ghttp.Init(map[string][]string{})

	// SSE example

	// request := ghttp.Request{
	// 	Method: "POST",
	// 	Url:    "https://api.openai.com/v1/chat/completions",
	// 	Body: map[string]interface{}{
	// 		"model":  "gpt-3.5-turbo-1106",
	// 		"stream": true,
	// 		"messages": []map[string]string{
	// 			{
	// 				"role":    "system",
	// 				"content": "Tu name es ia, eres la asistente virtual de TUL",
	// 			},
	// 			{
	// 				"role":    "user",
	// 				"content": "Hola, cual es tu name?",
	// 			},
	// 		},
	// 	},
	// 	Headers: map[string]string{
	// 		"Content-Type":  "application/json",
	// 		"Authorization": "Bearer " + os.Getenv("OPENAI_API_KEY"),
	// 	},
	// }

	// messageChannel := make(chan []byte)

	// go request.FetchSSE(messageChannel)

	// for message := range messageChannel {
	// 	fmt.Println(string(message))
	// }

	// fmt.Println(response.DataAsSlice())

	http.HandleFunc("/submit", formHandler)

	var wg sync.WaitGroup

	wg.Add(1)

	go http.ListenAndServe(":8000", nil)

	request := ghttp.Request{
		Method: "POST",
		Url:    "http://localhost:8000/submit",
		Body:   map[string]string{"name": "", "age": "30"},
		Headers: map[string][]string{
			"Content-Type": {"application/x-www-form-urlencoded"},
		},
		MaxRetries: 3,
	}

	response := request.Fetch()

	if response.Error != nil {
		fmt.Println(response.Error, response.DataAsMap())
		return
	}

	fmt.Println(string(response.Data))
	fmt.Println(response.Headers)

	wg.Wait()
}
