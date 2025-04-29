package api

import (
	"log"
	"net/http"
	"time"
)

func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	params := r.URL.Query()
	nowParam := params.Get("now")
	date := params.Get("date")
	repeat := params.Get("repeat")

	var now time.Time
	var err error

	if nowParam != "" {
		now, err = time.Parse(DateLayout, nowParam)
		if err != nil {
			log.Printf("Invalid 'now' format: %s", nowParam)
			http.Error(w, "Invalid 'now' format", http.StatusBadRequest)
			return
		}
	} else {
		now = time.Now().UTC()
	}

	nextDate, err := NextDate(now, date, repeat)
	if err != nil {
		log.Printf("Error computing next date: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(nextDate))
}
