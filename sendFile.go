package sendFile

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
)

func sendFile(directory, url string) {
	files, err := os.ReadDir(directory)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		name := file.Name()

		bodyBuffer := &bytes.Buffer{}
		bodyWriter := multipart.NewWriter(bodyBuffer)

		fileWriter, _ := bodyWriter.CreateFormFile("files", name)

		file, _ := os.Open(name)
		defer file.Close()

		io.Copy(fileWriter, file)

		contentType := bodyWriter.FormDataContentType()
		bodyWriter.Close()

		resp, _ := http.Post(url, contentType, bodyBuffer)
		defer resp.Body.Close()

		resp_body, _ := ioutil.ReadAll(resp.Body)

		log.Println(resp.Status)
		log.Println(string(resp_body))
	}

}
