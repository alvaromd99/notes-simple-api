# Go Notes API

A simple RESTful API written in Go that allows users to create, read, update, and delete notes, with data persisted to a local JSON file.

## Purpose

This project was created as a learning exercise to practice and improve Go programming skills, including working with file systems, JSON encoding/decoding, HTTP routing, and concurrency using mutexes.

## Features

- Handles CRUD operations for notes via REST endpoints.
- Stores data persistently in a local `notes.json` file.
- Uses mutex locks to ensure safe concurrent access to the data.
- Provides formatted JSON responses with proper status codes.
- Thread-safe implementation using Go’s `sync.RWMutex`.

## Requirements

- Go 1.22 or higher
- A `notes.json` file in the project root (can be empty array `[]` initially)

### Run the Project

```bash
go run main.go
```

The server will start on

```bash
localhost:8080
```

### Example Endpoints (using curl)

```bash
# Get all notes
curl -X GET http://localhost:8080/notes

# Create a new note
curl -X POST http://localhost:8080/notes \
  -H "Content-Type: application/json" \
  -d '{ "title": "My Note Title", "description": "This is the note description." }'

# Update a note
curl -X PATCH http://localhost:8080/notes/6 \
  -H "Content-Type: application/json" \
  -d '{"title": "Updated title","description": "Updated description text"}'

# Delete a note
curl -X DELETE http://localhost:8080/notes/6
```
