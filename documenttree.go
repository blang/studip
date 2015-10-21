package studip

// DocumentTree represents a tree of all semesters and its documents.
// API Endpoints:
// - api/studip-client-core/documenttree/
type DocumentTree []struct {
	SemesterID  string `json:"semester_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Courses     []struct {
		CourseID string   `json:"course_id"`
		CourseNr string   `json:"course_nr"`
		Title    string   `json:"title"`
		Folders  []Folder `json:"folders"`
	} `json:"courses"`
}

// Folder represents a folder in studips documents.
type Folder struct {
	FolderID    string `json:"folder_id"`
	Name        string `json:"name"`
	Mkdate      string `json:"mkdate"`
	Chdate      string `json:"chdate"`
	Permissions struct {
		Visible    bool `json:"visible"`
		Writable   bool `json:"writable"`
		Readable   bool `json:"readable"`
		Extendable bool `json:"extendable"`
	} `json:"permissions"`
	Subfolders []Folder `json:"subfolders"`
	Files      []struct {
		DocumentID string `json:"document_id"`
		Name       string `json:"name"`
		Mkdate     string `json:"mkdate"`
		Chdate     string `json:"chdate"`
		Filename   string `json:"filename"`
		Filesize   string `json:"filesize"`
		Protection string `json:"protection"`
		MimeType   string `json:"mime_type"`
	} `json:"files"`
}
