package api

import "net/http"

func routes(s *server, mux *http.ServeMux) {
	// FYI: I like all the routes in one place so I can easily see the whole surface of the API. Keep this
	// method clean and readable and sorted.

	mux.Handle("GET /api/events/{thread_id}", s.withContext(Events))
	mux.Handle("GET /api/version", s.api(Version))
}
