


setup routes - ```(SetupRoutes func)```


```
	router.HandleFunc("/jobs", s.handleCreateJob).Methods(http.MethodPost)
	router.HandleFunc("/jobs", s.handleGetJobs).Methods(http.MethodGet)
	router.HandleFunc("/jobs/{id}", s.handleGetJob).Methods(http.MethodGet)
	router.HandleFunc("/jobs/{id}", s.handleCancelJob).Methods(http.MethodDelete)
```
