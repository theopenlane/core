package config

// FileProvider is a struct that represents a file provider and generally "who" the file is from
type FileProvider struct {
	// URL is the web address that should be called to fetch the file
	URL string `json:"url,omitempty"`
	// File is the file name that should be used to save the file
	FileName string `json:"file"`
	// Path is the location on disk that should be used to store the file using the FileName
	Path string `json:"path"`
	// Hash is the hash of the file that should be used to verify the file
	Hash string `json:"hash"`
	// Category represents the type of file
	Category FileCategory
}

// FileCategory is a struct that represents the category of a file which would be roughly "what" the file is
type FileCategory struct {
	// Name is the name of the file category
	Name string `json:"name"`
}
