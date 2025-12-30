package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"gitclone/internal/app/repos"
)

// handleRepoIssues handles GET/POST /api/repos/:id/issues
func (s *Server) handleRepoIssues(w http.ResponseWriter, r *http.Request, repoID string) {
	if _, err := repos.ResolveRepoPath(s.repoBase, repoID); err != nil {
		RespondJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	if r.Method == http.MethodGet {
		issues, err := s.LoadIssues(repoID)
		if err != nil {
			RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}
		RespondJSON(w, http.StatusOK, issues)
	} else if r.Method == http.MethodPost {
		var req CreateIssueRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			RespondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
			return
		}

		if req.Title == "" {
			RespondJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Issue title is required"})
			return
		}

		authorEmail := req.Author
		if authorEmail == "" {
			authorEmail = "system"
		}
		avatarURL := fmt.Sprintf("https://api.dicebear.com/7.x/initials/svg?seed=%s", url.QueryEscape(authorEmail))

		issue := Issue{
			ID:           fmt.Sprintf("%s-%d", repoID, time.Now().UnixNano()),
			Title:        req.Title,
			Body:         req.Body,
			Status:       "open",
			Priority:     req.Priority,
			Labels:       req.Labels,
			Author:       authorEmail,
			AuthorAvatar: avatarURL,
			CreatedAt:    time.Now(),
			CommentCount: 0,
		}

		if err := s.SaveIssue(repoID, issue); err != nil {
			RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}

		RespondJSON(w, http.StatusCreated, issue)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleIssue handles GET/PATCH /api/repos/:id/issues/:issueId
func (s *Server) handleIssue(w http.ResponseWriter, r *http.Request, repoID, issueID string) {
	if _, err := repos.ResolveRepoPath(s.repoBase, repoID); err != nil {
		RespondJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	if r.Method == http.MethodGet {
		issues, err := s.LoadIssues(repoID)
		if err != nil {
			RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}

		for _, issue := range issues {
			if issue.ID == issueID {
				RespondJSON(w, http.StatusOK, issue)
				return
			}
		}

		RespondJSON(w, http.StatusNotFound, ErrorResponse{Error: "Issue not found"})
	} else if r.Method == http.MethodPatch || r.Method == http.MethodPut {
		issues, err := s.LoadIssues(repoID)
		if err != nil {
			RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}

		var found bool
		for i := range issues {
			if issues[i].ID == issueID {
				found = true
				var updateReq struct {
					Status string `json:"status,omitempty"`
					Body   string `json:"body,omitempty"`
				}
				_ = json.NewDecoder(r.Body).Decode(&updateReq)

				if updateReq.Status != "" {
					issues[i].Status = updateReq.Status
				} else {
					if issues[i].Status == "open" {
						issues[i].Status = "closed"
					} else {
						issues[i].Status = "open"
					}
				}

				if updateReq.Body != "" {
					issues[i].Body = updateReq.Body
				}
				break
			}
		}

		if !found {
			RespondJSON(w, http.StatusNotFound, ErrorResponse{Error: "Issue not found"})
			return
		}

		db := s.metaStore.GetDB()
		if db == nil {
			RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Database not available"})
			return
		}

		key := fmt.Sprintf("repo:%s:issues", repoID)
		data, err := json.Marshal(issues)
		if err != nil {
			RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}

		if err := db.Put(key, data); err != nil {
			RespondJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}

		for _, issue := range issues {
			if issue.ID == issueID {
				RespondJSON(w, http.StatusOK, issue)
				return
			}
		}
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
