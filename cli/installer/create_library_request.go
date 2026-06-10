package installer

type createLibraryRequest struct {
	Name      string          `json:"name"`
	MediaType string          `json:"mediaType"`
	Folders   []libraryFolder `json:"folders"`
}
