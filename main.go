package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Task struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	Duration  string    `json:"duration"`
	Result    string    `json:"result"`
}

var tasks = make(map[string]*Task)
var mu sync.Mutex

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	http.HandleFunc("/tasks", createTaskHandler)
	http.HandleFunc("/tasks/", getTaskHandler)
	http.HandleFunc("/tasks/delete/", deleteTaskHandler)

	log.Println("Сервер запущен на http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Println("Error starting the server:", err)
	}
}

func createTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Только POST", http.StatusMethodNotAllowed)
		return
	}

	id := strconv.FormatInt(time.Now().UnixNano(), 10)
	task := &Task{
		ID:        id,
		Status:    "pending",
		CreatedAt: time.Now(),
	}
	mu.Lock()
	tasks[id] = task
	mu.Unlock()

	go runTask(id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)

}

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Только GET", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Path[len("/tasks/"):]
	mu.Lock()
	task, ok := tasks[id]
	mu.Unlock()

	if !ok {
		http.Error(w, "Задача не найдена", http.StatusNotFound)
		return
	}

	task.Duration = time.Since(task.CreatedAt).Truncate(time.Second).String()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Только DELETE", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Path[len("/tasks/delete/"):]
	mu.Lock()
	_, ok := tasks[id]
	if ok {
		delete(tasks, id)
	}
	mu.Unlock()

	if !ok {
		http.Error(w, "Задача не найдена", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)

}

func runTask(id string) {
	mu.Lock()
	task, ok := tasks[id]
	if ok {
		task.Status = "running"
	}
	mu.Unlock()

	if !ok {
		return
	}

	time.Sleep(time.Duration(180+rand.Intn(120)) * time.Second)

	mu.Lock()
	if task, ok := tasks[id]; ok {
		task.Status = "done"
		task.Result = "Задача выполнена!"
	}
	mu.Unlock()

}
