package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
)

// Serie representa una serie de TV/Anime en la base de datos
type Serie struct {
	ID            int    `json:"id"`
	Title         string `json:"title"`
	Status        string `json:"status"`
	LastEpisode   int    `json:"last_episode"`
	TotalEpisodes int    `json:"total_episodes"`
	Ranking       int    `json:"ranking"`
}

var db *sql.DB

func main() {
	// Conexion a la base de datos
	var err error
	db, err = sql.Open("postgres", "host=db user=postgres password=postgres dbname=series sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Configurar el router
	router := mux.NewRouter()

	// Endpoints principales
	router.HandleFunc("/api/series", getSeries).Methods("GET")
	router.HandleFunc("/api/series/{id}", getSerie).Methods("GET")
	router.HandleFunc("/api/series", createSerie).Methods("POST")
	router.HandleFunc("/api/series/{id}", updateSerie).Methods("PUT")
	router.HandleFunc("/api/series/{id}", deleteSerie).Methods("DELETE")

	// Endpoints adicionales
	router.HandleFunc("/api/series/{id}/status", updateStatus).Methods("PATCH")
	router.HandleFunc("/api/series/{id}/episode", incrementEpisode).Methods("PATCH")
	router.HandleFunc("/api/series/{id}/upvote", upvoteSerie).Methods("PATCH")
	router.HandleFunc("/api/series/{id}/downvote", downvoteSerie).Methods("PATCH")

	// Servir documentacion Swagger
	router.PathPrefix("/docs").Handler(http.StripPrefix("/docs", http.FileServer(http.Dir("./docs"))))

	// Configurar CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
	})

	// Iniciar el servidor
	handler := c.Handler(router)
	srv := &http.Server{
		Handler:      handler,
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	fmt.Println("Servidor iniciado en el puerto 8080")
	log.Fatal(srv.ListenAndServe())
}

// getSeries obtiene todas las series
func getSeries(w http.ResponseWriter, r *http.Request) {
	statusFilter := r.URL.Query().Get("status")
	order := r.URL.Query().Get("order")

	query := "SELECT id, title, status, last_episode, total_episodes, ranking FROM series"
	if statusFilter != "" && statusFilter != "All" {
		query += " WHERE status = $1"
	}

	if order == "Descending" {
		query += " ORDER BY ranking DESC"
	} else {
		query += " ORDER BY ranking ASC"
	}

	var rows *sql.Rows
	var err error

	if statusFilter != "" && statusFilter != "All" {
		rows, err = db.Query(query, statusFilter)
	} else {
		rows, err = db.Query(query)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var series []Serie
	for rows.Next() {
		var s Serie
		if err := rows.Scan(&s.ID, &s.Title, &s.Status, &s.LastEpisode, &s.TotalEpisodes, &s.Ranking); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		series = append(series, s)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(series)
}

// getSerie obtiene una serie por ID
func getSerie(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	var s Serie
	err = db.QueryRow("SELECT id, title, status, last_episode, total_episodes, ranking FROM series WHERE id = $1", id).
		Scan(&s.ID, &s.Title, &s.Status, &s.LastEpisode, &s.TotalEpisodes, &s.Ranking)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Serie no encontrada", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}

// createSerie crea una nueva serie
func createSerie(w http.ResponseWriter, r *http.Request) {
	var s Serie
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := db.QueryRow(
		"INSERT INTO series (title, status, last_episode, total_episodes, ranking) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		s.Title, s.Status, s.LastEpisode, s.TotalEpisodes, s.Ranking,
	).Scan(&s.ID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(s)
}

// updateSerie actualiza una serie existente
func updateSerie(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	var s Serie
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = db.Exec(
		"UPDATE series SET title = $1, status = $2, last_episode = $3, total_episodes = $4, ranking = $5 WHERE id = $6",
		s.Title, s.Status, s.LastEpisode, s.TotalEpisodes, s.Ranking, id,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// deleteSerie elimina una serie
func deleteSerie(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("DELETE FROM series WHERE id = $1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// updateStatus actualiza el estado de una serie
func updateStatus(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validar el estado
	validStatuses := map[string]bool{
		"Plan to Watch": true,
		"Watching":      true,
		"Dropped":       true,
		"Completed":     true,
	}
	if !validStatuses[req.Status] {
		http.Error(w, "Estado inválido", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("UPDATE series SET status = $1 WHERE id = $2", req.Status, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// incrementEpisode incrementa el episodio actual
func incrementEpisode(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("UPDATE series SET last_episode = last_episode + 1 WHERE id = $1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// upvoteSerie aumenta la puntuacion de una serie
func upvoteSerie(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("UPDATE series SET ranking = ranking + 1 WHERE id = $1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// downvoteSerie disminuye la puntuacion de una serie
func downvoteSerie(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("UPDATE series SET ranking = ranking - 1 WHERE id = $1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
