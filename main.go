package main

import (
	"aegis/poc/db"
	"aegis/poc/handlers"
	"context"
	"fmt"
	"net/http"
)

func main() {
	redisClient := db.SetupRedis()
	if redisClient == nil {
		fmt.Println("Failed to setup Redis")
		return
	}
	redisClient.FlushAll(context.Background())

	handlerDeps := handlers.NewHandlerDependencies(redisClient)

	// Serve static files and setup handlers
	fs := http.FileServer(http.Dir("./static"))
	http.HandleFunc("/queue", handlerDeps.QueueHandler)
	http.HandleFunc("/random", handlers.RandomHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			fs.ServeHTTP(w, r)
			return
		}
		w.Write([]byte("Assalamualaikum warahmatullahi wabarakatuh, selamat menjalankan ibadah puasa."))
	})

	fmt.Println("Listening on : http://127.0.0.1:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println(err)
	}
}
