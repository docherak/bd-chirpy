package main

import (
	"net/http"
)

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.env != "dev" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Reset is only allowed in dev env."))
		return
	}
	cfg.fileserverHits.Store(0)
	err := cfg.db.DeleteAllUsers(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to reset the user database: " + err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0 and user database reset to initial state."))
}
