package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"rockets/unsplash"

	b64 "encoding/base64"
	"rockets/seam"

	"github.com/flosch/pongo2"
)

const (
	Image_URL string = "https://bit.ly/2QGPDkr"
)

type Task struct {
	Position int
	URL      string
}

type TaskResult struct {
	Position int
	Resized  []byte
	Err      error
}

var spacexTemplate = pongo2.Must(pongo2.FromFile("spacex.html"))

func main() {
	fmt.Println("Ready for liftoff! Checkout\nhttp://localhost:3000/occupymars\nhttp://localhost:3000/spacex\nhttp://localhost:3000/spacex_seams")

	http.HandleFunc("/occupymars", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("resize") > "" {
			resized, err := seam.ContentAwareResize(Image_URL)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "image/jpeg")
			io.Copy(w, bytes.NewReader(resized))
		} else {
			fmt.Fprintf(w, "<html><div>Original image:</div> <img src=\"%s\" /><br/><a href=\"?resize=1\">Resize using Seam Carving</a></html>", Image_URL)
		}
	})

	http.HandleFunc("/spacex", func(w http.ResponseWriter, r *http.Request) {
		response, err := unsplash.LoadRockets()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		err = spacexTemplate.ExecuteWriter(pongo2.Context{"response": response}, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/spacex_seams", func(w http.ResponseWriter, r *http.Request) {
		response, err := unsplash.LoadRockets()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		resultChannel := make(chan TaskResult)
		taskChannel := make(chan Task)
		imagesToResize := 8

		// start 4 workers
		for w := 1; w <= 4; w++ {
			go worker(w, taskChannel, resultChannel)
		}

		// write to the taskChannel and close it when we're done
		go func() {
			for i, r := range response.Results[:imagesToResize] {
				taskChannel <- Task{i, r.URLs["small"]}
			}
			close(taskChannel)
		}()

		// start listening for results in a separate goroutine
		for a := 1; a <= imagesToResize; a++ {
			taskResult := <-resultChannel
			if taskResult.Err != nil {
				log.Printf("Image %d failed to resize", taskResult.Position)
			} else {
				sEnc := b64.StdEncoding.EncodeToString(taskResult.Resized)
				response.Results[taskResult.Position].Resized = sEnc
			}
		}

		err = spacexTemplate.ExecuteWriter(pongo2.Context{"response": response}, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	log.Fatal(http.ListenAndServe("localhost:3000", nil))
}

func worker(id int, taskChannel <-chan Task, resultChannel chan<- TaskResult) {
	for j := range taskChannel {
		fmt.Println("worker", id, "started job", j.Position)
		resized, err := seam.ContentAwareResize(j.URL)
		resultChannel <- TaskResult{j.Position, resized, err}
	}
}
