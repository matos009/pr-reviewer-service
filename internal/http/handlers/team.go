package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"pr-reviewer-service/internal/domain"
)

func (s *Server) PostTeamAdd(w http.ResponseWriter, r *http.Request) {
	var body PostTeamAddJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// OpenAPI модель → domain модель
	var members []domain.TeamMember
	for _, m := range body.Members {
		members = append(members, domain.TeamMember{
			ID:       m.UserId,
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}

	team := &domain.Team{
		Name:    body.TeamName,
		Members: members,
	}

	err := s.TeamService.CreateWithMembers(r.Context(), team)
	if err != nil {
		if errors.Is(err, domain.ErrTeamExists) {
			// открываем согласно openapi: 400 + код TEAM_EXISTS
			resp := PostTeamAdd400JSONResponse{
				Error: struct {
					Code    ErrorResponseErrorCode `json:"code"`
					Message string                 `json:"message"`
				}{
					Code:    TEAMEXISTS,
					Message: "team already exists",
				},
			}
			resp.VisitPostTeamAddResponse(w)
			return
		}

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := PostTeamAdd201JSONResponse{
		Team: &Team{
			TeamName: body.TeamName,
			Members:  body.Members,
		},
	}
	resp.VisitPostTeamAddResponse(w)
}

func (s *Server) GetTeamGet(w http.ResponseWriter, r *http.Request, params GetTeamGetParams) {
	team, err := s.TeamService.Get(r.Context(), params.TeamName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(team)
}
func (s *Server) PostTeamDeactivate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Team string `json:"team"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad JSON", http.StatusBadRequest)
		return
	}
	if req.Team == "" {
		http.Error(w, "team is required", http.StatusBadRequest)
		return
	}

	if err := s.TeamAdminService.DeactivateTeam(r.Context(), req.Team); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
