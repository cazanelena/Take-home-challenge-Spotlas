package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "secret"
	dbname   = "postgres"
)

type Spot struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Website     string  `json:"website"`
	Description string  `json:"description"`
	Rating      float64 `json:"rating"`
	Coordinates string  `json:"coordinates"`
}

func main() {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error connecting to the database: ", err)
	}
	defer db.Close()

	http.HandleFunc("/spots", func(w http.ResponseWriter, r *http.Request) {
		latitudeStr := r.URL.Query().Get("latitude")
		longitudeStr := r.URL.Query().Get("longitude")
		radiusStr := r.URL.Query().Get("radius")
		spotType := r.URL.Query().Get("type")

		// Parse the parameters to float64
		latitude, err := strconv.ParseFloat(latitudeStr, 64)
		if err != nil {
			http.Error(w, "Invalid latitude parameter", http.StatusBadRequest)
			return
		}

		longitude, err := strconv.ParseFloat(longitudeStr, 64)
		if err != nil {
			http.Error(w, "Invalid longitude parameter", http.StatusBadRequest)
			return
		}

		radius, err := strconv.ParseFloat(radiusStr, 64)
		if err != nil {
			http.Error(w, "Invalid radius parameter", http.StatusBadRequest)
			return
		}

		spots := getSpotsInArea(db, latitude, longitude, radius, spotType)
		json.NewEncoder(w).Encode(spots)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getSpotsInArea(db *sql.DB, latitude, longitude, radius float64, spotType string) []Spot {
	var query string

	switch spotType {
	case "circle":
		query = `
		WITH my_table_with_distance AS (
			SELECT 
				id, name, website, description, rating, 
				ST_AsText(coordinates) AS coordinates,
				ST_Distance(coordinates, ST_SetSRID(ST_MakePoint($1, $2), 4326)) AS distance
			FROM "MY_TABLE"
			WHERE ST_DWithin(coordinates, ST_SetSRID(ST_MakePoint($1, $2), 4326), $3)
		)
		SELECT 
			id, name, website, description, rating, coordinates, distance
		FROM my_table_with_distance
		ORDER BY 
			CASE 
				WHEN distance < 50 THEN rating
				ELSE distance
			END,
			CASE 
				WHEN distance < 50 THEN NULL
				ELSE rating
			END NULLS LAST;
		`
	case "square":
		query = `
		SELECT 
			id, name, website, description, rating, coordinates, distance
		FROM (
			SELECT 
				id, name, website, description, rating,
				ST_AsText(coordinates) AS coordinates,
				ST_Distance(coordinates, ST_SetSRID(ST_MakePoint($1, $2), 4326)) AS distance
			FROM "MY_TABLE"
			WHERE ST_Distance(coordinates, ST_SetSRID(ST_MakePoint($1, $2), 4326)) <= $3
		) AS subquery
		ORDER BY 
			CASE
				WHEN distance < 50 THEN rating
				ELSE distance
			END,
			rating NULLS LAST;
		`
	default:
		// If the spotType is invalid, return an empty slice
		return []Spot{}
	}

	rows, err := db.Query(query, longitude, latitude, radius)
	if err != nil {
		log.Println("Error querying database:", err)
		return []Spot{}
	}
	defer rows.Close()

	var spots []Spot
	for rows.Next() {
		var spot Spot
		var throwaway float64
		err := rows.Scan(
			&spot.ID, &spot.Name, &spot.Website, &spot.Description, &spot.Rating, &spot.Coordinates, &throwaway,
		)
		if err != nil {
			log.Println("Error scanning row:", err)
			continue
		}
		spots = append(spots, spot)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error iterating over rows:", err)
		return []Spot{}
	}

	return spots
}
