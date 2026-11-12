package handler

import (
	"aaa_bekend/internal/db"
	"aaa_bekend/internal/lmodel"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
)

type SubmitProjectReq struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type GradeProjectReq struct {
	Score    int    `json:"score"`
	Feedback string `json:"feedback"`
}

// POST /api/v1/projects

func SubmitProject(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := lmodel.UserFromContext(r.Context())
	var req SubmitProjectReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	projectID, err := db.CreateProject(r.Context(), req.Title, req.Description, userCtx.ID)
	if err != nil {
		http.Error(w, "Failed to create project", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"project_id": projectID})
}

func ListProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := db.GetAllProjects(r.Context())
	if err != nil {
		http.Error(w, "Failed to fetch projects", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(&projects)
}
func GetAllProjectGrades(w http.ResponseWriter, r *http.Request) {
	projects, err := db.GetAllProjectGrades(r.Context())
	if err != nil {
		http.Error(w, "Failed to fetch project grades", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(&projects)
}
func GradeProject(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := lmodel.UserFromContext(r.Context())

	var req GradeProjectReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	projectID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	project, err := db.GetProjectByID(r.Context(), projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to fetch project", http.StatusInternalServerError)
		return
	}

	if project.StudentID == userCtx.ID {
		http.Error(w, "Forbidden: you cannot grade your own project", http.StatusForbidden)
		return
	}

	err = db.CreateGrade(r.Context(), projectID, userCtx.ID, req.Score, req.Feedback)
	if err != nil {
		http.Error(w, "Failed to submit grade", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Grade Submitted"})
}

func GetProjectGrades(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := lmodel.UserFromContext(r.Context())
	projectID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	project, err := db.GetProjectByID(r.Context(), projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to fetch project", http.StatusInternalServerError)
		return
	}

	if userCtx.Role == "student" {
		if project.StudentID != userCtx.ID {
			http.Error(w, "Forbidden: you can only view grades for your own projects", http.StatusForbidden)
			return
		}

	}

	grades, err := db.GetGradesForProject(r.Context(), projectID)
	if err != nil {
		http.Error(w, "Failed to fetch grades", http.StatusInternalServerError)
		return
	}

	if grades == nil {
		grades = []lmodel.Grade{}
	}
	json.NewEncoder(w).Encode(grades)

}
