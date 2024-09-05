package handlers

import "net/http"

func getRandomQuote(w http.ResponseWriter, r *http.Request) {
	var quote string
	err := db.Get(&quote, "SELECT quote FROM quotes ORDER BY RANDOM() LIMIT 1")
	if err != nil {
		http.Error(w, "Could not fetch quote", http.StatusInternalServerError)
		return
	}
	w.Write([]byte(quote))
}
