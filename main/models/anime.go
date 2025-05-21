package models

// Anime represents the detailed information for an anime series.
// It includes metadata such as title, poster image, synopsis, genre, and episode count.
type Anime struct {
	ID       string   `json:"id"`       // Unique identifier for the anime (often a slug).
	Name     string   `json:"name"`     // The main title of the anime.
	Poster   string   `json:"poster"`   // URL to the poster image for the anime.
	Type     string   `json:"type"`     // The type of anime (e.g., "TV", "Movie", "OVA").
	Synopsis string   `json:"synopsis"` // A brief summary or description of the anime's plot.
	State    string   `json:"state"`    // Current airing state (e.g., "Airing", "Finished Airing", "Not yet aired").
	Genre    []string `json:"genre"`    // A list of genres associated with the anime (e.g., "Action", "Comedy").
	Episodes string   `json:"episodes"` // The number of episodes, typically as a string (e.g., "12", "24", "Unknown").
}

// Episode represents a single episode of an anime series.
// It primarily holds an identifier and the episode number or title.
type Episode struct {
	Id      string `json:"id"`      // Unique identifier for the episode (could be a slug or a specific ID).
	Episode string `json:"episode"` // The episode number or title (e.g., "1", "S1E1", "The First Adventure").
}

// Server represents a video streaming server that hosts an anime episode.
// It contains the name of the server and the direct URL to the video.
type Server struct {
	Name string `json:"name"` // Name of the video server (e.g., "Mega", "YourUpload", "Streamtape (HD)").
	Url  string `json:"url"`  // Direct URL to the video file or streaming page.
}

// Slug is used to group a list of animes under a specific category name or slug.
// For example, a slug could be "trending-anime" with a list of Anime objects.
type Slug struct {
	Name   string  `json:"name"`   // The name or title of the category (e.g., "Top Animes", "Latest Episodes").
	Animes []Anime `json:"animes"` // A list of Anime objects belonging to this category/slug.
}
