package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"nexu-jllr/config"
	"os"
	"strings"
	"sync"

	"github.com/lib/pq"
)

var (
	db   *sql.DB
	once sync.Once
)

type Brand struct {
	ID           int      `json:"id"`
	Name         string   `json:"name"`
	AveragePrice *float64 `json:"average_price"`
}

type Model struct {
	ID           int      `json:"id"`
	Name         string   `json:"name"`
	AveragePrice *float64 `json:"average_price"`
	BrandID      int      `json:"brand_id"`
	BrandName    string   `json:"brand_name,omitempty"`
}

func InitDB() {
	once.Do(func() {
		connStr := fmt.Sprintf(
			"host=db port=5432 user=%s password=%s dbname=%s sslmode=disable",
			config.GetPostgresUser(),
			config.GetPostgresPassword(),
			config.GetPostgresDB(),
		)

		var err error
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			log.Fatalf("Failed to open DB connection: %v", err)
		}

		if err := db.Ping(); err != nil {
			log.Fatalf("Failed to connect to DB: %v", err)
		}

		log.Println("Connected to PostgreSQL successfully")
		initialize()
	})
}

func initialize() {
	createBrandsTable := `
		CREATE TABLE IF NOT EXISTS brands (
			id SERIAL PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			average_price DOUBLE PRECISION
		);`

	createModelsTable := `
		CREATE TABLE IF NOT EXISTS models (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			brand_id INTEGER NOT NULL REFERENCES brands(id) ON DELETE CASCADE,
			average_price DOUBLE PRECISION,
			UNIQUE (name, brand_id)
		);`

	if _, err := db.Exec(createBrandsTable); err != nil {
		log.Fatalf("Could not create brands table: %v", err)
	}

	if _, err := db.Exec(createModelsTable); err != nil {
		log.Fatalf("Could not create models table: %v", err)
	}

	loadInitialData()
}

func getNextModelID() (int, error) {
	var id int
	err := db.QueryRow("SELECT COALESCE(MAX(id), 0) + 1 FROM models").Scan(&id)
	return id, err
}

func loadInitialData() {
	data, err := os.ReadFile("config/json/models.json")
	if err != nil {
		log.Fatalf("Failed to read models.json: %v", err)
	}

	var models []Model
	if err := json.Unmarshal(data, &models); err != nil {
		log.Fatalf("Failed to parse models.json: %v", err)
	}

	brandIDs := map[string]int{}
	for _, model := range models {
		if _, exists := brandIDs[model.BrandName]; !exists {
			var brandID int
			err := db.QueryRow(`
				INSERT INTO brands (name) 
				VALUES ($1)
				ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
				RETURNING id`,
				model.BrandName,
			).Scan(&brandID)

			if err != nil {
				log.Fatalf("Failed to insert brand %s: %v", model.BrandName, err)
			}

			brandIDs[model.BrandName] = brandID
		}
	}

	for _, model := range models {
		brandID := brandIDs[model.BrandName]
		var exists bool

		if err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM models WHERE id = $1)", model.ID).Scan(&exists); err != nil {
			log.Printf("Error checking if model %d exists: %v", model.ID, err)
			continue
		}

		if !exists {
			_, err := db.Exec(`
				INSERT INTO models (id, name, brand_id, average_price) 
				VALUES ($1, $2, $3, $4)`,
				model.ID, model.Name, brandID, model.AveragePrice,
			)
			if err != nil {
				log.Printf("Could not insert model %s: %v", model.Name, err)
			}
		}
	}
}

func InsertBrand(brand Brand) (Brand, error) {
	var id int
	err := db.QueryRow("INSERT INTO brands (name) VALUES ($1) RETURNING id", brand.Name).Scan(&id)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return Brand{}, fmt.Errorf("brand name already exists")
		}
		return Brand{}, err
	}

	brand.ID = id
	return brand, nil
}

func InsertModel(m Model) (Model, error) {
	newID, err := getNextModelID()
	if err != nil {
		return Model{}, err
	}
	m.ID = newID

	query := `
		INSERT INTO models (id, name, brand_id, average_price)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	err = db.QueryRow(query, m.ID, m.Name, m.BrandID, m.AveragePrice).Scan(&m.ID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return Model{}, fmt.Errorf("model name already exists for this brand")
		}
		return Model{}, err
	}

	return m, nil
}

func GetAllBrands() ([]Brand, error) {
	rows, err := db.Query("SELECT id, name, average_price FROM brands ORDER BY name ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var brands []Brand
	for rows.Next() {
		var b Brand
		if err := rows.Scan(&b.ID, &b.Name, &b.AveragePrice); err != nil {
			return nil, err
		}
		brands = append(brands, b)
	}
	return brands, nil
}

func GetAllModels(greater, lower *float64) ([]Model, error) {
	query := `SELECT id, name, brand_id, average_price FROM models`
	var conditions []string
	var args []interface{}
	i := 1

	if greater != nil {
		conditions = append(conditions, fmt.Sprintf("average_price > $%d", i))
		args = append(args, *greater)
		i++
	}
	if lower != nil {
		conditions = append(conditions, fmt.Sprintf("average_price < $%d", i))
		args = append(args, *lower)
		i++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY average_price"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var models []Model
	for rows.Next() {
		var m Model
		if err := rows.Scan(&m.ID, &m.Name, &m.BrandID, &m.AveragePrice); err != nil {
			return nil, err
		}
		models = append(models, m)
	}

	return models, nil
}

func UpdateModel(m Model) (Model, error) {
	if m.AveragePrice == nil || *m.AveragePrice < 100000 {
		return Model{}, fmt.Errorf("average_price must be at least 100000")
	}

	query := `
		UPDATE models
		SET average_price = $1
		WHERE id = $2
		RETURNING id, name, brand_id, average_price`

	var updated Model
	err := db.QueryRow(query, *m.AveragePrice, m.ID).Scan(
		&updated.ID, &updated.Name, &updated.BrandID, &updated.AveragePrice,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return Model{}, fmt.Errorf("model not found")
		}
		return Model{}, err
	}

	return updated, nil
}

func GetModelsByBrandID(brandID int) ([]Model, error) {
	rows, err := db.Query(`
		SELECT id, name, brand_id, average_price 
		FROM models 
		WHERE brand_id = $1 
		ORDER BY name ASC`, brandID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var models []Model
	for rows.Next() {
		var m Model
		if err := rows.Scan(&m.ID, &m.Name, &m.BrandID, &m.AveragePrice); err != nil {
			return nil, err
		}
		models = append(models, m)
	}

	return models, nil
}
