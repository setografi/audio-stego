package main

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
)

const delimiter = "\x00\x01\x02\x03" // Delimiter unik untuk menandai akhir pesan

func main() {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/embed", handleEmbed)
	http.HandleFunc("/extract", handleExtract)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func handleEmbed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, header, err := r.FormFile("audio")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	message := r.FormValue("message")

	ext := filepath.Ext(header.Filename)
	if ext != ".mp3" && ext != ".wav" {
		http.Error(w, "Only MP3 and WAV files are supported", http.StatusBadRequest)
		return
	}

	audioData, err := readAudio(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	stegoData, err := embedMessage(audioData, message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=stego"+ext)
	w.Header().Set("Content-Type", "audio/"+ext[1:])
	w.Write(stegoData)
}

func handleExtract(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("audio")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	audioData, err := readAudio(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	message, err := extractMessage(audioData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(message))
}

func readAudio(file io.Reader) ([]byte, error) {
	return ioutil.ReadAll(file)
}

func embedMessage(audioData []byte, message string) ([]byte, error) {
	messageWithDelimiter := message + delimiter
	messageBytes := []byte(messageWithDelimiter)
	if len(messageBytes)*8 > len(audioData) {
		return nil, errors.New("message is too long for the given audio file")
	}

	result := make([]byte, len(audioData))
	copy(result, audioData)

	bitIndex := 0
	for _, messageByte := range messageBytes {
		for i := 7; i >= 0; i-- {
			messageBit := (messageByte >> uint(i)) & 1
			result[bitIndex] = (result[bitIndex] & 0xFE) | messageBit
			bitIndex++
		}
	}

	return result, nil
}

func extractMessage(audioData []byte) (string, error) {
	var extractedBytes []byte
	var currentByte byte
	bitCount := 0

	for _, audioByte := range audioData {
		currentByte = (currentByte << 1) | (audioByte & 1)
		bitCount++

		if bitCount == 8 {
			extractedBytes = append(extractedBytes, currentByte)
			currentByte = 0
			bitCount = 0

			// Cek apakah kita telah mencapai delimiter
			if len(extractedBytes) >= len(delimiter) {
				if bytes.HasSuffix(extractedBytes, []byte(delimiter)) {
					return string(extractedBytes[:len(extractedBytes)-len(delimiter)]), nil
				}
			}
		}
	}

	return "", errors.New("no hidden message found or message is corrupted")
}
