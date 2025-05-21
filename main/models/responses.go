package models

// ArrayResponse is a generic API response structure for returning a list of Anime objects.
// It includes status and message fields alongside the data.
type ArrayResponse struct {
	Data    []Anime `json:"data"`    // Slice of Anime objects, representing the main content of the response.
	Status  string  `json:"status"`  // Status code of the response (e.g., "200", "404"). Consider using integers for HTTP status.
	Message string  `json:"message"` // A human-readable message accompanying the response (e.g., "Success", "Not Found").
}

// SearchAnimeResponse is used for API responses that return a list of Anime objects
// resulting from a search operation, along with pagination information.
type SearchAnimeResponse struct {
	Data    []Anime     `json:"data"`    // Slice of Anime objects found by the search.
	Status  string      `json:"status"`  // Status code of the response.
	Message string      `json:"message"` // Accompanying message.
	Page    interface{} `json:"page"`    // Pagination information. Typically an integer representing the next page number,
	// or a special value (e.g., -1 or null) to indicate the last page.
	// Using interface{} allows flexibility but might require type assertion on the client side.
}

// Response is a generic wrapper for API responses where the data can be of any type.
// It includes status and message fields.
type Response struct {
	Data    interface{} `json:"data"`    // The main data payload of the response, can be any type (e.g., a single Anime object, a list of Servers).
	Status  string      `json:"status"`  // Status code of the response.
	Message string      `json:"message"` // Accompanying message.
}

// SlugResponse represents a named list of animes, typically used for categorized listings.
// The 'Data' field is an interface{} but is expected to contain a list of animes,
// which might be a design choice for flexibility or due to specific JSON marshalling needs.
type SlugResponse struct {
	Data interface{} `json:"animes"` // Expected to be a list of Anime objects, associated with the slug name.
	Name string      `json:"name"`   // The name or title of the slug/category (e.g., "Top Animes").
}

// SlugMainResponse is used for API responses that return multiple SlugResponse objects.
// This is useful for pages that display several categorized lists of anime.
type SlugMainResponse struct {
	Data    []SlugResponse `json:"data"`    // A list of SlugResponse objects, each representing a category of animes.
	Status  string         `json:"status"`  // Status code of the overall response.
	Message string         `json:"message"` // Accompanying message for the overall response.
}
