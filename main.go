package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"path"
)

var MAIN_PAGE = []byte(`
<html>
  <head>
    <title>Uploader</title>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/dropzone/4.3.0/dropzone.css" integrity="sha256-0Z6mOrdLEtgqvj7tidYQnCYWG3G2GAIpatAWKhDx+VM=" crossorigin="anonymous" />
    <script src="https://cdnjs.cloudflare.com/ajax/libs/dropzone/4.3.0/dropzone.js" integrity="sha256-vnXjg9TpLhXuqU0OcVO7x+DpR/H1pCeVLLSeQ/I/SUs=" crossorigin="anonymous"></script>
  </head>
  <body>
    <div style="position: absolute; top: 25%; left: 25%; width: 50%; height: 50%; text-align: center; margin: 0 auto;">
      <form action="/" method="post" class="dropzone" id="my-awesome-dropzone" enctype="multipart/form-data">
        <!-- Nothing to see here. -->
      </form>
    </div>
  </body>
</html>
`)

func saveUploadedFiles(w http.ResponseWriter, r *http.Request) {
	log.Printf("Started uploading files from %s", r.RemoteAddr)

	// Allow a maximum upload length of 64MB.
	if err := r.ParseMultipartForm(1 << 16); err != nil {
		log.Printf("Blocked attempted upload a file larger than 64MB from %s.", r.RemoteAddr)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Loop over the metadata for each uploaded file.
	for _, upload := range r.MultipartForm.File["file"] {
		log.Printf("Processing upload of '%s' from %s", upload.Filename, r.RemoteAddr)

		source, err := upload.Open()

		if err != nil {
			log.Printf("Upload of '%s' from %s failed with error %+v", upload.Filename, r.RemoteAddr, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer source.Close()

		destination, err := os.Create(path.Join(storage, upload.Filename))

		if err != nil {
			log.Printf("Saving of '%s' from %s failed with error %+v", upload.Filename, r.RemoteAddr, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer destination.Close()

		if _, err := io.Copy(destination, source); err != nil {
			log.Printf("Writing of '%s' from %s failed with error %+v", upload.Filename, r.RemoteAddr, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("Upload of '%s' from %s completed successfully!", upload.Filename, r.RemoteAddr)
	}

	log.Printf("Finished uploading files from %s", r.RemoteAddr)
}

func showUploadPage(w http.ResponseWriter, r *http.Request) {
	w.Write(MAIN_PAGE)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		showUploadPage(w, r)
	case "POST":
		saveUploadedFiles(w, r)
	}
}

var storage = "./data"

func main() {
	// Allow overriding the data storage path on the command line.
	flag.StringVar(&storage, "data", storage, "Data storage directory.")
	flag.Parse()

	// Attempt to create a the storage directory, or fail.
	if _, err := os.Stat(storage); os.IsNotExist(err) {
		if err := os.Mkdir(storage, os.ModePerm); err != nil {
			log.Println("Storage directory does not exist and cannot be created!", err)
			return
		}
	}

	// Don't try to serve a favicon.
	http.HandleFunc("/favicon.ico", http.NotFound)

	// Handle uploads as normal.
	http.HandleFunc("/", uploadHandler)

	// Launch an HTTP server on :8080.
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("http.ListenAndServe: ", err)
	}
}
