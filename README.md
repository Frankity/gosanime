# Gosanime

Gosanime is a kind of web scrapper written in [Go](https://go.dev/), the purpose of it is read the content of [Jkanime](https://jkanime.net/) site and serve as an API.

**API Endpoints:**

Here's a detailed description of the available API endpoints:

### Root

*   **Endpoint:** `/`
*   **Method:** `GET`
*   **Purpose:** Provides a welcome message indicating the API is running.
*   **Request Parameters:** None.
*   **Response Format:**
    ```json
    {
      "message": "Welcome to Gosanime API"
    }
    ```

### Main Page Content

*   **Endpoint:** `/api/v1/main`
*   **Method:** `GET`
*   **Purpose:** Fetches the main page content, typically a list of trending or recently added anime.
*   **Request Parameters:** None.
*   **Response Format:** `models.ArrayResponse`
    ```json
    {
      "data": [
        {
          "id": "anime-id",
          "name": "Anime Name",
          "poster": "image-url.jpg",
          "type": "TV",
          "synopsis": "Brief description of the anime.",
          "state": "Airing",
          "genre": ["action", "adventure"],
          "episodes": "12"
        }
        // ... more anime objects
      ],
      "status": "200",
      "message": "Success"
    }
    ```

### OVAs

*   **Endpoint:** `/api/v1/ovas`
*   **Method:** `GET`
*   **Purpose:** Fetches a list of OVAs (Original Video Animations).
*   **Request Parameters:** None.
*   **Response Format:** `models.ArrayResponse` (Same structure as `/api/v1/main`)

### Search Anime

*   **Endpoint:** `/api/v1/search`
*   **Method:** `GET`
*   **Purpose:** Searches for anime based on a query string and provides paginated results.
*   **Request Parameters:**
    *   `anime` (string, required): The name or keyword of the anime to search for (e.g., `naruto`).
    *   `page` (integer, optional, default: `1`): The page number for pagination.
*   **Response Format:** `models.SearchAnimeResponse`
    ```json
    {
      "data": [
        // ... list of anime objects (models.Anime)
      ],
      "status": "200",
      "message": "Success",
      "page": 2 // Next page number, or -1 if it's the last page
    }
    ```

### Anime Details

*   **Endpoint:** `/api/v1/anime`
*   **Method:** `GET`
*   **Purpose:** Fetches detailed information about a specific anime using its ID.
*   **Request Parameters:**
    *   `id` (string, required): The ID of the anime (e.g., `naruto`).
*   **Response Format:** `models.Response` containing `models.Anime`
    ```json
    {
      "data": {
        "id": "naruto",
        "name": "Naruto",
        "poster": "image-url.jpg",
        "type": "TV",
        "synopsis": "Detailed synopsis of Naruto.",
        "state": "Finished Airing",
        "genre": ["action", "adventure", "comedy"],
        "episodes": "220"
      },
      "status": "200",
      "message": "Success"
    }
    ```

### Video Servers

*   **Endpoint:** `/api/v1/video`
*   **Method:** `GET`
*   **Purpose:** Fetches available video server links for a specific episode of an anime.
*   **Request Parameters:**
    *   `anime` (string, required): The slug of the anime (e.g., `spy-x-family/`). Note the trailing slash.
    *   `episode` (string, required): The episode number (e.g., `1`).
*   **Response Format:** `models.Response` containing a list of `models.Server`
    ```json
    {
      "data": [
        {
          "name": "Server Name (e.g., JKServer)",
          "url": "decoded-video-url"
        }
        // ... more server objects
      ],
      "status": "200",
      "message": "Success"
    }
    ```

### Anime by Tag

*   **Endpoint:** `/api/v1/tags`
*   **Method:** `GET`
*   **Purpose:** Fetches anime associated with a specific tag and provides paginated results.
*   **Request Parameters:**
    *   `tag` (string, required): The tag name to filter by (e.g., `shounen`).
    *   `page` (integer, optional, default: `1`): The page number for pagination.
*   **Response Format:** `models.SearchAnimeResponse` (Same structure as `/api/v1/search`)

Insomnia collection [here](Insomnia_2022-07-15.json)

## Architecture Overview

Gosanime is a web scraper and API server built with Go. It fetches data from Jkanime and exposes it through a RESTful API.

### Core Components

*   **`main.go`**: This is the entry point of the application. It initializes the server (from `main/server`) and starts the HTTP listener.
*   **`main/server/server.go`**: Responsible for setting up the HTTP router using `gorilla/mux`. It defines all the API routes and maps them to their respective handler functions.
*   **`main/server/*.go`** (e.g., `anime.go`, `ovas.go`, `videos.go`): These files contain the specific handler logic for each API endpoint. This includes fetching data from the target website (Jkanime), parsing the HTML content (often using libraries like `soup`), and preparing the response.
*   **`main/models/`**: This directory holds the Go struct definitions that represent the data structures used throughout the application, such as API responses (`Anime`, `Server`, `ArrayResponse`, etc.).
*   **`main/config/`**: This directory is intended for application configuration, such as base URLs for scraping or server settings.
*   **`main/utils/`**: Contains utility functions, like the HTTP client wrapper (`http_utils.go`), used by various parts of the application.

### Data Flow

A typical request to the Gosanime API follows this flow:

1.  An HTTP request is received by the Go `http.Server` instance started in `main.go`.
2.  The `gorilla/mux` router, configured in `main/server/server.go`, directs the request to the appropriate handler function based on the URL path and HTTP method.
3.  The designated handler function (located in one of the `main/server/*.go` files) executes. This typically involves:
    *   Making HTTP requests to Jkanime to fetch raw HTML data.
    *   Parsing the HTML to extract the required information.
    *   Structuring this data into the defined Go types from `main/models/`.
4.  The handler function sends the structured data back to the client as a JSON response.

### External Dependencies

*   **`gorilla/mux`**: A powerful URL router and dispatcher for Go. Used to manage API routing.
*   **`github.com/anaskhan96/soup`**: A library for web scraping, used to parse HTML content from Jkanime.
*   **`github.com/go-resty/resty/v2`**: A feature-rich HTTP client for Go, likely used for making requests to Jkanime (though `utils.NewHTTPClient()` might be custom).

## Wanna Try It? Project Setup and Usage

Here's how you can set up and run Gosanime on your local machine:

### Prerequisites

*   **Go:** You need to have Go installed on your system. If you haven't installed it yet, you can find the official installation guide at [https://go.dev/doc/install](https://go.dev/doc/install).

### Installation Steps

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/Frankity/gosanime.git
    ```

2.  **Navigate to the project directory:**
    ```bash
    cd gosanime
    ```

3.  **Install dependencies:**
    The `go get ./` command will download the necessary dependencies. Alternatively, for Go modules-based projects like this one, you might prefer using `go mod tidy` which ensures your `go.mod` file matches the source code and downloads dependencies.
    ```bash
    go get ./
    # or
    # go mod tidy
    ```

### Running the Application

1.  **Run the application:**
    This command will compile and run the Go application.
    ```bash
    go run main.go
    ```

2.  **Access the API:**
    The server will start and listen on port **3000** by default (e.g., `http://localhost:3000`).
    You can access the API endpoints using a web browser (for GET requests), or tools like `curl`, Postman, or Insomnia.

    For example, to check if the API is running, open your browser and navigate to `http://localhost:3000/`.
    To explore other functionalities, refer to the **API Endpoints** section above for detailed information on each endpoint, including request parameters and response formats.

## Code Documentation (GoDoc)

GoDoc is a tool that extracts and generates documentation from Go source code comments. This project includes GoDoc comments for public types, functions, and methods.

### Generating and Viewing Documentation

1.  **Install GoDoc (if necessary):**
    GoDoc is part of the standard Go toolchain and should be available with your Go installation. If you need to install it manually or want to ensure you have the latest version for running a local documentation server, you can use:
    ```bash
    go install golang.org/x/tools/cmd/godoc@latest
    ```
    For older Go versions (before Go 1.16), you might use `go get` instead:
    ```bash
    # For Go versions < 1.16
    # go get golang.org/x/tools/cmd/godoc
    ```

2.  **Run the GoDoc HTTP Server:**
    To view the documentation in your browser, you can run a local GoDoc server. Open your terminal and execute the following command (you can choose a different port if `:6060` is occupied):
    ```bash
    godoc -http=:6060
    ```

3.  **Access Project Documentation:**
    Once the server is running, open your web browser and navigate to the following URL:
    [http://localhost:6060/pkg/xyz.frankity/gosanime/](http://localhost:6060/pkg/xyz.frankity/gosanime/)

    This will display the documentation for the `xyz.frankity/gosanime` package and its sub-packages (like `main/server`, `main/models`, etc.), rendered from the GoDoc comments in the source code.

---
**To Do:**

- Decode correctly another video server due the only one i played with is the jk one.
- Fix the pagination logic.
- Paginate tags correctly.
- Increase the search response range.
- Add more endpoints for other pages of the site.

willing to help?, fork and send some pr's.
