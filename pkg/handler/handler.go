package handler

import (
	"encoding/json"
	"net/http"
	"nexu-jllr/pkg/db"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type Error struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func GetAllBrands(w http.ResponseWriter, r *http.Request) {
	setJSONHeader(w)

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	brands, err := db.GetAllBrands()
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if len(brands) == 0 {
		sendError(w, http.StatusNotFound, "There are no brands")
		return
	}

	_ = json.NewEncoder(w).Encode(brands)
}

func GetAllModels(w http.ResponseWriter, r *http.Request) {
	setJSONHeader(w)

	query := r.URL.Query()

	var greater, lower *float64

	if val := query.Get("greater"); val != "" {
		if g, err := strconv.ParseFloat(val, 64); err == nil {
			greater = &g
		} else {
			http.Error(w, "Invalid greater value", http.StatusBadRequest)
			return
		}
	}

	if val := query.Get("lower"); val != "" {
		if l, err := strconv.ParseFloat(val, 64); err == nil {
			lower = &l
		} else {
			http.Error(w, "Invalid lower value", http.StatusBadRequest)
			return
		}
	}

	models, err := db.GetAllModels(greater, lower)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Failed to filter models")
		return
	}

	if len(models) == 0 {
		sendError(w, http.StatusNotFound, "There are no models")
		return
	}

	json.NewEncoder(w).Encode(models)
}

func CreateBrand(w http.ResponseWriter, r *http.Request) {
	setJSONHeader(w)

	var brand db.Brand
	if err := json.NewDecoder(r.Body).Decode(&brand); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	created, err := db.InsertBrand(brand)
	if err != nil {
		sendError(w, http.StatusConflict, err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func CreateModelByBrandID(w http.ResponseWriter, r *http.Request) {
	setJSONHeader(w)

	brandID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid brand ID", http.StatusBadRequest)
		return
	}

	var model db.Model
	if err := json.NewDecoder(r.Body).Decode(&model); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	model.BrandID = brandID
	created, err := db.InsertModel(model)
	if err != nil {
		sendError(w, http.StatusConflict, err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func UpdateModel(w http.ResponseWriter, r *http.Request) {
	setJSONHeader(w)

	modelID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid model ID", http.StatusBadRequest)
		return
	}

	var update db.Model
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	update.ID = modelID

	updated, err := db.UpdateModel(update)
	if err != nil {
		sendError(w, http.StatusConflict, err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(updated)
}

func GetModelsByBrandID(w http.ResponseWriter, r *http.Request) {
	setJSONHeader(w)

	brandID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid brand ID", http.StatusBadRequest)
		return
	}

	models, err := db.GetModelsByBrandID(brandID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Could not retrieve models")
		return
	}

	if len(models) == 0 {
		sendError(w, http.StatusNotFound, "There are no models")
		return
	}

	json.NewEncoder(w).Encode(models)
}

func setJSONHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

func sendError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(Error{
		Status:  code,
		Message: message,
	})
}
