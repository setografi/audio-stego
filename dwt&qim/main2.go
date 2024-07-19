package main

import (
    "bytes"
    "fmt"
    "io/ioutil"
    "net/http"
    "github.com/go-audio/audio"
    "github.com/go-audio/wav"
    "github.com/go-audio/mp3"
    "dwt&qim/dwt"
    "dwt&qim/qim"
)

func main() {
    http.HandleFunc("/", indexHandler)
    http.HandleFunc("/embed", embedHandler)
    http.ListenAndServe(":8080", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "templates/index.html")
}

func embedHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodPost {
        r.ParseMultipartForm(10 << 20) // 10 MB limit

        message := r.FormValue("message")
        audioFile, _, err := r.FormFile("audioFile")
        if err != nil {
            http.Error(w, "Unable to get audio file", http.StatusBadRequest)
            return
        }
        defer audioFile.Close()

        // Read audio file
        var audioData []byte
        audioData, err = ioutil.ReadAll(audioFile)
        if err != nil {
            http.Error(w, "Unable to read audio file", http.StatusInternalServerError)
            return
        }

        // Process audio file
        var audioSignal []float64
        // Assuming WAV format for simplicity
        decoder := wav.NewDecoder(bytes.NewReader(audioData))
        for {
            buf, err := decoder.PCMBuffer()
            if err != nil {
                break
            }
            audioSignal = append(audioSignal, buf.Data...)
        }

        // Apply DWT
        waveletCoeffs := dwt.Transform(audioSignal)

        // Apply QIM
        modCoeffs := qim.ModifyCoefficients(waveletCoeffs, message)

        // Reconstruct Signal
        encodedSignal := dwt.Reconstruct(modCoeffs)

        // Output or save the encoded audio
        // ...

        fmt.Fprintln(w, "Embedding successful")
    } else {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
    }
}
