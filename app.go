package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"strconv"
)

type App struct {
	Router        *mux.Router
	DatabaseType  string
	DatabaseSQL   *sql.DB
	DatabaseNoSQL *redis.Client
}

func (a *App) Initialize(db string, user, password, dbname string) {
	a.DatabaseType = db

	if db == "postgresql" {
		connectionString := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", user, password, dbname)
		var err error
		a.DatabaseSQL, err = sql.Open("postgres", connectionString)
		if err != nil {
			log.Fatal(err)
		}
	} else if db == "redis" {
		a.DatabaseNoSQL = redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: password, // no password set
			DB:       0,  // use default DB
		})
	} else {
		log.Fatal(fmt.Printf("Unknown DB type: %s\n", db))
	}

	a.Router = mux.NewRouter()

	a.initializeRoutes()
}
func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(":8000", a.Router))
}

func (a *App) getShortByShort(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortCurrent := vars["short"]
	if shortCurrent == "" {
		respondWithError(w, http.StatusBadRequest, "Invalid short")
		return
	}

	s := short{Short: shortCurrent}
	if err := s.getShortByShort(a.DatabaseSQL); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Short not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, s)
}

func (a *App) getShort(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid short ID")
		return
	}

	s := short{ID: uint64(id)}
	if err := s.getShort(a.DatabaseSQL); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Short not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, s)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (a *App) getShorts(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))

	if count > 10 || count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}

	shorts, err := getShorts(a.DatabaseSQL, start, count)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, shorts)
}

func (a *App) createShort(w http.ResponseWriter, r *http.Request) {
	var s short
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&s); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if err := s.createShort(a.DatabaseSQL); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, s)
}

func (a *App) updateShort(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid short ID")
		return
	}

	var s short
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&s); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid resquest payload")
		return
	}
	defer r.Body.Close()
	s.ID = uint64(id)

	if err := s.updateShort(a.DatabaseSQL); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, s)
}

func (a *App) deleteShort(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid short ID")
		return
	}

	s := short{ID: uint64(id)}
	if err := s.deleteShort(a.DatabaseSQL); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/shorts", a.getShorts).Methods("GET")
	a.Router.HandleFunc("/shorts", a.createShort).Methods("POST")
	a.Router.HandleFunc("/shorts/{id:[0-9]+}", a.getShort).Methods("GET")
	a.Router.HandleFunc("/shorts/{id:[0-9]+}", a.updateShort).Methods("PUT")
	a.Router.HandleFunc("/shorts/{id:[0-9]+}", a.deleteShort).Methods("DELETE")
	a.Router.HandleFunc("/shorts/{short:[0-9]*[a-z]*[A-Z]*_*}", a.getShortByShort).Methods("GET")
}