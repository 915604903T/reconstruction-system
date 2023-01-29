package handlers

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

func MakeFileHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reader, err := r.MultipartReader()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			fmt.Printf("FileName=[%s], FormName=[%s]\n", part.FileName(), part.FormName())
			if part.FileName() == "" { // this is FormData
				data, _ := ioutil.ReadAll(part)
				fmt.Printf("FormData=[%s]\n", string(data))
			} else { // This is FileData
				dst, _ := os.Create("./" + part.FileName())
				defer dst.Close()
				io.Copy(dst, part)
			}
		}
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("Upload file success!"))
	}
}
