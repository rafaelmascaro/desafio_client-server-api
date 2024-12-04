package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	consultaDolar(ctx)
}

func consultaDolar(ctx context.Context) /*([]byte, error)*/ {
	res, err := http.Get("https://economia.awesomeapi.com.br/json/last/USD-BRL")
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println("maiscaro ", body)
	//return body, nil
}
