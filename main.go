package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strconv"
	"sync"
)

// Test endpoints
// curl -X GET http://localhost:8080/notes
// curl -X POST http://localhost:8080/notes  -H "Content-Type: application/json"  -d '{ "title": "My Note Title", "description": "This is the note description." }'
// curl -X PATCH http://localhost:8080/notes/6  -H "Content-Type: application/json" -d '{"title": "Updated title","description": "Updated description text"}'
// curl -X DELETE http://localhost:8080/notes/6

type Note struct {
	Id          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

var (
	fileName   = "./notes.json"
	notesMutex = &sync.RWMutex{}
)

func getNotes(w http.ResponseWriter, r *http.Request) {
	// Indicamos que la respuesta es un json
	w.Header().Set("Content-Type", "application/json")

	// Evitar colisiones con otros procesos
	notesMutex.RLock()
	content, err := os.ReadFile(fileName)
	defer notesMutex.RUnlock() // Usar defer para asegurarnos que siempre se haga

	if err != nil {
		http.Error(w, "Error reading the notes file.", http.StatusInternalServerError)
		return
	}

	var data []Note
	err = json.Unmarshal(content, &data)
	if err != nil {
		http.Error(w, "Error during Unmarshal.", http.StatusInternalServerError)
		return
	}

	// Creamos el encoder para agregar indentación al json
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		http.Error(w, "Error encoding JSON.", http.StatusInternalServerError)
		return
	}
	// No w.WriteHeader(http.StatusOK) porque json.encoder lo pone solo
}

func getNoteById(w http.ResponseWriter, r *http.Request) {
	// Indicamos que la respuesta es un json
	w.Header().Set("Content-Type", "application/json")

	notesMutex.RLock()
	content, err := os.ReadFile(fileName)
	defer notesMutex.RUnlock()

	if err != nil {
		http.Error(w, "Error reading the notes file.", http.StatusInternalServerError)
		return
	}

	var data []Note
	err = json.Unmarshal(content, &data)
	if err != nil {
		http.Error(w, "Error during Unmarshal.", http.StatusInternalServerError)
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Error converting string to int", http.StatusBadRequest)
		return
	}

	index, found := findNoteById(id, &data)
	if !found {
		http.Error(w, "Error searching the note.", http.StatusNotFound)
		return
	}

	// Creamos el encoder para agregar indentación al json
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data[index]); err != nil {
		http.Error(w, "Error encoding JSON.", http.StatusInternalServerError)
		return
	}
}

// No devolvemos la nota y así es mejor para usarlo en todos los métodos
func findNoteById(id int, notes *[]Note) (int, bool) {
	for i, n := range *notes {
		if n.Id == id {
			return i, true
		}
	}
	return -1, false
}

func addNote(w http.ResponseWriter, r *http.Request) {
	// Indicamos que la respuesta es un json
	w.Header().Set("Content-Type", "application/json")

	var newNote Note

	if err := json.NewDecoder(r.Body).Decode(&newNote); err != nil {
		http.Error(w, "Error invalid JSON payload", http.StatusBadRequest)
		return
	}

	if newNote.Title == "" || newNote.Description == "" {
		http.Error(w, "Invalid title or description", http.StatusBadRequest)
		return
	}

	notesMutex.Lock()
	content, err := os.ReadFile(fileName)
	defer notesMutex.Unlock()

	if err != nil {
		http.Error(w, "Error reading the notes file.", http.StatusInternalServerError)
		return
	}

	var data []Note
	err = json.Unmarshal(content, &data)
	if err != nil {
		http.Error(w, "Error during Unmarshal.", http.StatusInternalServerError)
		return

	}
	// Le damos valor al id de la nueva nota
	lastId := data[len(data)-1].Id
	newNote.Id = lastId + 1

	// Agregamos la nueva nota
	data = append(data, newNote)

	// Formateamos las notas y las escribimos otra vez en el archivo
	updatedNotes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		http.Error(w, "Error preparing data for saving", http.StatusInternalServerError)
		return
	}

	// Escribimos las nueva lista de notas
	if err = os.WriteFile(fileName, updatedNotes, 0644); err != nil {
		http.Error(w, "Error saving the new note", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	// Creamos el encoder para agregar indentación al json
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	res := map[string]string{
		"message": "Note created successfully.",
		"id":      strconv.Itoa(newNote.Id),
	}
	if err := encoder.Encode(res); err != nil {
		http.Error(w, "Error encoding JSON.", http.StatusInternalServerError)
		return
	}
}

func deleteNoteById(w http.ResponseWriter, r *http.Request) {
	// Indicamos que la respuesta es un json
	w.Header().Set("Content-Type", "application/json")

	// Como se hacen lecturas y escrituras tiene que ser Lock no RLock
	notesMutex.Lock()
	content, err := os.ReadFile(fileName)
	defer notesMutex.Unlock()

	if err != nil {
		http.Error(w, "Error reading the notes file.", http.StatusInternalServerError)
		return
	}

	var data []Note
	err = json.Unmarshal(content, &data)
	if err != nil {
		http.Error(w, "Error during Unmarshal.", http.StatusInternalServerError)
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Error converting string to int", http.StatusInternalServerError)
		return
	}

	index, found := findNoteById(id, &data)
	if !found {
		http.Error(w, "Error searching the note.", http.StatusNotFound)
		return
	}

	// Borramos la nota por su índice
	data = deleteNote(index, &data)

	// Formateamos las notas y las escribimos otra vez en el archivo
	updatedNotes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		http.Error(w, "Error preparing data for saving", http.StatusInternalServerError)
		return
	}

	// Escribimos las nueva lista de notas
	if err = os.WriteFile(fileName, updatedNotes, 0644); err != nil {
		http.Error(w, "Error saving the new note", http.StatusInternalServerError)
		return
	}

	// Creamos el encoder para agregar indentación al json
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	res := map[string]string{
		"message": "Note deleted successfully.",
		"id":      strconv.Itoa(id),
	}
	if err := encoder.Encode(res); err != nil {
		http.Error(w, "Error encoding JSON.", http.StatusInternalServerError)
		return
	}
}

func deleteNote(index int, notes *[]Note) []Note {
	return slices.Delete(*notes, index, index+1)
}

func modifyNote(w http.ResponseWriter, r *http.Request) {
	var updatedNote Note

	// Leemos y comprobamos el body
	if err := json.NewDecoder(r.Body).Decode(&updatedNote); err != nil {
		http.Error(w, "Error invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Comprobamos el id de la url
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Error converting string to int", http.StatusBadRequest)
		return
	}

	// Leemos las notas
	notesMutex.Lock()
	content, err := os.ReadFile(fileName)
	defer notesMutex.Unlock()

	if err != nil {
		http.Error(w, "Error reading the notes file.", http.StatusInternalServerError)
		return
	}

	var data []Note
	err = json.Unmarshal(content, &data)
	if err != nil {
		http.Error(w, "Error during Unmarshal.", http.StatusInternalServerError)
		return
	}

	// Buscamos la nota que queremos cambiar
	index, found := findNoteById(id, &data)
	if !found {
		http.Error(w, "Error searching the note.", http.StatusNotFound)
		return
	}

	// Modificamos los campos
	data[index].Title = updatedNote.Title
	data[index].Description = updatedNote.Description

	// Formateamos las notas y las escribimos otra vez en el archivo
	updatedNotes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		http.Error(w, "Error preparing data for saving", http.StatusInternalServerError)
		return
	}

	// Escribimos las nueva lista de notas
	if err = os.WriteFile(fileName, updatedNotes, 0644); err != nil {
		http.Error(w, "Error saving the new note", http.StatusInternalServerError)
		return
	}

	// Creamos el encoder para agregar indentación al json
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	res := map[string]string{
		"message": "Note updated successfully.",
		"id":      strconv.Itoa(data[index].Id),
	}
	if err := encoder.Encode(res); err != nil {
		http.Error(w, "Error encoding JSON.", http.StatusInternalServerError)
		return
	}
}

func main() {
	http.HandleFunc("GET /notes", getNotes)
	http.HandleFunc("GET /notes/{id}", getNoteById)
	http.HandleFunc("POST /notes", addNote)
	http.HandleFunc("PATCH /notes/{id}", modifyNote)
	http.HandleFunc("DELETE /notes/{id}", deleteNoteById)

	fmt.Println("Server starting on port :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Server failed to start: %v", err)
	}
}
