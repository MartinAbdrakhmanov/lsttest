package main

import (
	"html/template"
	"io"
	"log"
	"math"
	"net/http"
	"regexp"
	"sort"
	"strings"
)

type WordStats struct {
	Word string
	TF   int
	IDF  float64
}

func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/upload", uploadHandler)

	log.Println("Сервер запущен на :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Ошибка получения файла", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Ошибка чтения файла", http.StatusInternalServerError)
		return
	}

	text := string(data)
	cleanText := strings.ToLower(text)
	re := regexp.MustCompile(`[^\w\s]`)
	cleanText = re.ReplaceAllString(cleanText, " ")

	words := strings.Fields(cleanText)
	// tf
	freqMap := make(map[string]int)
	for _, word := range words {
		freqMap[word]++
	}

	totalWords := len(words)
	var stats []WordStats
	for word, tf := range freqMap {
		// Вычисляем idf как ln(totalWords / tf), так как файл всего 1 (хотя tf idf на одном файле неэффективен)
		idf := math.Log(float64(totalWords) / float64(tf))
		stats = append(stats, WordStats{Word: word, TF: tf, IDF: idf})
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].IDF > stats[j].IDF
	})

	if len(stats) > 50 {
		stats = stats[:50]
	}

	tmpl, err := template.ParseFiles("templates/result.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, stats)
}
