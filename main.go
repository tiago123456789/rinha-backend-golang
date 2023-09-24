package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"rinha-backend-golang/cache"
	"rinha-backend-golang/database"
	"rinha-backend-golang/repository"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()
	db := database.Start()

	client, ctx := cache.Start()
	pong, err := client.Ping(ctx).Result()
	fmt.Println(pong, err)

	r := mux.NewRouter()

	r.HandleFunc("/pessoas/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		peopleCached, _ := client.Get(ctx, id).Result()
		if len(peopleCached) > 0 {
			fmt.Fprint(w, peopleCached)
			return
		}

		people, err := repository.GetPeopleById(db, id)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		peopleToString, _ := json.Marshal(people)
		client.SetEX(ctx, id, peopleToString, time.Second*3)
		json.NewEncoder(w).Encode(people)
	}).Methods("GET")

	r.HandleFunc("/contagem-pessoas", func(w http.ResponseWriter, r *http.Request) {
		keyCached := "total"
		totalCached, _ := client.Get(ctx, keyCached).Result()
		if len(totalCached) > 0 {
			fmt.Fprint(w, totalCached)
			return
		}

		total, err := repository.GetTotalPeoples(db)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		client.SetEX(ctx, keyCached, string(total), time.Second*2)
		fmt.Fprint(w, total)
	}).Methods("GET")

	r.HandleFunc("/pessoas", func(w http.ResponseWriter, r *http.Request) {
		term := r.URL.Query().Get("t")

		keyCached := term
		value, _ := client.Get(ctx, keyCached).Result()
		if len(value) > 0 {
			fmt.Fprint(w, value)
			return
		}

		peoples, err := repository.GetAllByTerm(db, term)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		jsonResponse, _ := json.Marshal(peoples)
		client.SetEX(ctx, keyCached, jsonResponse, time.Second*3)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	}).Methods("GET")

	r.HandleFunc("/pessoas", func(w http.ResponseWriter, r *http.Request) {
		var p repository.Pessoa
		json.NewDecoder(r.Body).Decode(&p)

		_, err := repository.Insert(db, &p)
		if err != nil {
			log.Fatal(err)
		}

		w.WriteHeader(http.StatusCreated)
	}).Methods("POST")

	srv := &http.Server{
		Handler: r,
		Addr:    "127.0.0.1:8000",
	}

	fmt.Println("Server is running")
	log.Fatal(srv.ListenAndServe())
}
