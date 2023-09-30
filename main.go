package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"rinha-backend-golang/cache"
	"rinha-backend-golang/database"
	"rinha-backend-golang/repository"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

type MessageValidation struct {
	Message string `json:'message'`
}

const MAX_GOROTINES = 10

func main() {
	godotenv.Load()
	db := database.Start()

	client, ctx := cache.Start()
	pong, err := client.Ping(ctx).Result()
	fmt.Println(pong, err)

	guard := make(chan struct{}, MAX_GOROTINES)
	r := mux.NewRouter()

	r.HandleFunc("/pessoas/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		if len(id) == 0 {
			fmt.Fprint(w, MessageValidation{
				Message: "Id is required",
			})
		}

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

		if len(term) == 0 {
			fmt.Fprint(w, MessageValidation{
				Message: "Term is required",
			})
			return
		}

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

		if len(p.Apelido) == 0 {
			fmt.Fprint(w, MessageValidation{
				Message: "Apelido is required",
			})
			return
		}

		if len(p.Nome) == 0 {
			fmt.Fprint(w, MessageValidation{
				Message: "Nome is required",
			})
			return
		}

		if len(p.Nascimento) == 0 {
			fmt.Fprint(w, MessageValidation{
				Message: "Nascimento is required",
			})
			return
		}

		if len(p.Apelido) > 36 {
			fmt.Fprint(w, MessageValidation{
				Message: "Apelido can't more than 36 characters",
			})
			return
		}

		if len(p.Nome) > 100 {
			fmt.Fprint(w, MessageValidation{
				Message: "Nome can't more than 100 characters",
			})
			return
		}

		isValidDate, _ := regexp.MatchString("([0-9]){4}-([0-9]){2}-([0-9]){2}", p.Nascimento)
		if !isValidDate {
			fmt.Fprint(w, MessageValidation{
				Message: "Nascimento is invalid. The correct format YYY-MM-DD",
			})
			return
		}

		if len(p.Stack) > 0 {
			for _, value := range p.Stack {
				if len(value) > 32 {
					fmt.Fprint(w, MessageValidation{
						Message: "Stack has one item has more than 32 characters",
					})
					return
				}
			}
		}

		guard <- struct{}{}
		go func(db *sql.DB, p *repository.Pessoa) {
			_, err := repository.Insert(db, p)
			if err != nil {
				log.Fatal(err)
			}
			<-guard
		}(db, &p)
		w.WriteHeader(http.StatusCreated)
	}).Methods("POST")

	srv := &http.Server{
		Handler: r,
		Addr:    "0.0.0.0:8000",
	}

	fmt.Println("Server is running")
	log.Fatal(srv.ListenAndServe())
}
