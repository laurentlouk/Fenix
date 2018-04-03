package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"gopkg.in/mgo.v2/bson"

	"github.com/codegangsta/negroni"

	"github.com/gorilla/mux"
	. "github.com/laurentlouk/fenix/auth"
	. "github.com/laurentlouk/fenix/config"
	. "github.com/laurentlouk/fenix/dao"
	. "github.com/laurentlouk/fenix/models"
)

var config = Config{}
var dao = MoviesDAO{}
var auth = Auth{}
var user = UserCredentials{}

// POST login JWT
func LoginEndPoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		respondWithError(w, http.StatusBadRequest, "Error in login request")
		return
	}

	if strings.ToLower(user.Username) != "someone" || user.Username == "" {
		respondWithError(w, http.StatusBadRequest, "Invalid credentials")
		return
	}
	if user.Password != "p@ssword" || user.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Invalid credentials")
		return
	}

	//Check for JWT token
	token, err := user.CheckForToken()

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error while extracting or signing the token")
		return
	}

	respondWithJson(w, http.StatusOK, map[string]string{"Token": token})
}

// GET list of movies
func AllMoviesEndPoint(w http.ResponseWriter, r *http.Request) {
	movies, err := dao.FindAll()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJson(w, http.StatusOK, movies)
}

// GET a movie by its ID
func FindMovieEndpoint(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	movie, err := dao.FindById(params["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Movie ID")
		return
	}
	respondWithJson(w, http.StatusOK, movie)
}

// POST a new movie
func CreateMovieEndPoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var movie Movie
	if err := json.NewDecoder(r.Body).Decode(&movie); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	movie.ID = bson.NewObjectId()
	if err := dao.Insert(movie); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJson(w, http.StatusCreated, movie)
}

// PUT update an existing movie
func UpdateMovieEndPoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var movie Movie
	if err := json.NewDecoder(r.Body).Decode(&movie); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if err := dao.Update(movie); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJson(w, http.StatusOK, map[string]string{"result": "success"})
}

// DELETE an existing movie
func DeleteMovieEndPoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var movie Movie
	if err := json.NewDecoder(r.Body).Decode(&movie); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if err := dao.Delete(movie); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJson(w, http.StatusOK, map[string]string{"result": "success"})
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJson(w, code, map[string]string{"error": msg})
}

func respondWithJson(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// Parse the configuration file 'config.toml', and establish a connection to DB
func init() {
	auth.InitKeys()
	config.Read()
	dao.Server = config.Server
	dao.Database = config.Database
	dao.Connect()
}

// Define HTTP request routes
func main() {
	r := mux.NewRouter().StrictSlash(false)

	// Middle validation of token
	acctBase := mux.NewRouter()
	r.PathPrefix("/account").Handler(negroni.New(
		negroni.NewRecovery(),
		negroni.HandlerFunc(ValidateTokenMiddleware),
		negroni.NewLogger(),
		negroni.Wrap(acctBase),
	))
	// Account subrouter - routes
	acct := acctBase.PathPrefix("/account").Subrouter()
	acct.Path("/movies").HandlerFunc(CreateMovieEndPoint).Methods("POST")
	acct.Path("/movies/{id}").HandlerFunc(FindMovieEndpoint).Methods("GET") // path can't fetch var {id}

	// Public routes
	r.HandleFunc("/login", LoginEndPoint).Methods("POST")
	r.HandleFunc("/movies", AllMoviesEndPoint).Methods("GET")
	//r.HandleFunc("/movies", CreateMovieEndPoint).Methods("POST")
	r.HandleFunc("/movies", UpdateMovieEndPoint).Methods("PUT")
	r.HandleFunc("/movies", DeleteMovieEndPoint).Methods("DELETE")
	//r.HandleFunc("/movies/{id}", FindMovieEndpoint).Methods("GET")
	if err := http.ListenAndServe(":3000", r); err != nil {
		log.Fatal(err)
	}
}
