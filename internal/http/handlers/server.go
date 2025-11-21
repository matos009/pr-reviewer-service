package handlers

import (
	"encoding/json"
	"net/http"
	"pr-reviewer-service/internal/service"
)

type Server struct {
	TeamService      *service.TeamService
	UserService      *service.UserService
	PRService        *service.PRService
	TeamAdminService *service.TeamAdminService
}

func NewServer(
	ts *service.TeamService,
	us *service.UserService,
	prs *service.PRService,
	admin *service.TeamAdminService,
) *Server {
	return &Server{
		TeamService:      ts,
		UserService:      us,
		PRService:        prs,
		TeamAdminService: admin,
	}
}
func (s *Server) GetStats(w http.ResponseWriter, r *http.Request) {
	rv, st, err := s.PRService.Stats(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"reviewerAssignments": rv,
		"prStatus":            st,
	})
}
