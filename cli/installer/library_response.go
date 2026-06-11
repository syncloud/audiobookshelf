package installer

type libraryResponse struct {
	ID       string `json:"id"`
	LastScan *int64 `json:"lastScan"`
}
