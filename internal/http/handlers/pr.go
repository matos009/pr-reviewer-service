package handlers

import (
	"encoding/json"
	"net/http"
)

func (s *Server) PostPullRequestCreate(w http.ResponseWriter, r *http.Request) {
	var body PostPullRequestCreateJSONRequestBody

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad JSON", http.StatusBadRequest)
		return
	}

	pr, err := s.PRService.Create(
		r.Context(),
		body.PullRequestId,
		body.PullRequestName,
		body.AuthorId,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(pr)
}

func (s *Server) PostPullRequestMerge(w http.ResponseWriter, r *http.Request) {
	var body PostPullRequestMergeJSONRequestBody

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad JSON", http.StatusBadRequest)
		return
	}

	pr, err := s.PRService.Merge(r.Context(), body.PullRequestId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(pr)
}

func (s *Server) PostPullRequestReassign(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID                string `json:"id"`
		ReviewerToReplace string `json:"reviewerId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad JSON", http.StatusBadRequest)
		return
	}

	pr, newID, err := s.PRService.ReassignReviewer(r.Context(), req.ID, req.ReviewerToReplace)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := struct {
		NewReviewer string      `json:"newReviewer"`
		PR          interface{} `json:"pr"`
	}{
		NewReviewer: newID,
		PR:          pr,
	}

	json.NewEncoder(w).Encode(resp)
}

func (s *Server) GetUsersGetReview(w http.ResponseWriter, r *http.Request, params GetUsersGetReviewParams) {
	list, err := s.PRService.GetUserReviews(r.Context(), params.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(list)
}
