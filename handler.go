package gokakera

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
)

func (k *Kakera) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /upload/session", k.handleCreateSession)
	mux.HandleFunc("PUT /upload/{sessionID}/chunk/{index}", k.handleUploadChunk)
	mux.HandleFunc("POST /upload/{sessionID}/complete", k.handleComplete)
	mux.HandleFunc("GET /upload/{sessionID}/status", k.handleStatus)

}

func (k *Kakera) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Filename  string `json:"filename"`
		TotalSize int64  `json:"total_size"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
	}

	session, err := k.CreateSession(r.Context(), req.Filename, req.TotalSize)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"session_id":   session.ID,
		"total_chunks": session.TotalChunks,
		"chunk_size":   k.config.ChunkSize,
		"expires_at":   session.ExpiresAt,
	})
}

func (k *Kakera) handleUploadChunk(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("sessionID")

	index, err := strconv.Atoi(r.PathValue("index"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid chunk index")
		return
	}

	checksum := r.Header.Get("X-Checksum")
	if checksum == "" {
		writeError(w, http.StatusBadRequest, "missing X-Checksum header")
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read chunk data")
		return
	}
	defer r.Body.Close()

	if err := k.UploadChunk(r.Context(), sessionID, index, data, checksum); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message": "chunk uploaded",
		"index":   index,
	})
}

func (k *Kakera) handleComplete(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("sessionID")

	complete, err := k.IsUploadComplete(r.Context(), sessionID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if !complete {
		writeError(w, http.StatusBadRequest, "upload not complete, missing chunks")
		return
	}

	if err := k.AssembleFile(r.Context(), sessionID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message": "file assembled successfully",
	})
}

func (k *Kakera) handleStatus(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("sessionID")

	session, err := k.GetSession(r.Context(), sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"session_id":      session.ID,
		"filename":        session.Filename,
		"status":          session.Status,
		"total_chunks":    session.TotalChunks,
		"received_chunks": len(session.ReceivedChunks),
		"percentage":      float64(len(session.ReceivedChunks)) / float64(session.TotalChunks) * 100,
		"expires_at":      session.ExpiresAt,
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}
