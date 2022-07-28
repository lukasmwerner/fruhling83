package main

func (s *server) ConfigureRoutes() {
	s.router.Use(loggingMiddleware)
	s.router.HandleFunc("/{key}", corsHandler(s.GetBoard)).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/{key}", corsHandler(s.ChangeBoardContent)).Methods("PUT", "OPTIONS")
	s.router.HandleFunc("/", corsHandler(s.LandingPage)).Methods("GET", "OPTIONS")
}
