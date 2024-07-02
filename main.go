package main

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var templates = template.Must(template.ParseFiles("index.html"))

type FileInfo struct {
	Name  string
	Path  string
	IsDir bool
}

type PageData struct {
	Path  string
	Dirs  []FileInfo
	Files []FileInfo
	Query string
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := templates.ExecuteTemplate(w, tmpl+".html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func fileHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[1:] // Remove the leading "/"
	spath := r.URL.Query().Get("p")
	query := r.URL.Query().Get("q")

	if path == "" {
		path = "./files"
	}
	/* Hier wird geprüft, ob eine Suche gestartet wurde, oder nicht */
	if spath != "" {
		path = spath
		//fmt.Println("Search::Path=", path)
		fileInfo, err := os.Stat(path)
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}

		var dirs []FileInfo
		var files []FileInfo

		if fileInfo.IsDir() {

			err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if p == path {
					return nil
				}
				if query == "" || strings.Contains(strings.ToLower(info.Name()), strings.ToLower(query)) {
					file := FileInfo{
						Name:  info.Name(),
						Path:  p,
						IsDir: info.IsDir(),
					}
					if info.IsDir() {
						dirs = append(dirs, file)
					} else {
						files = append(files, file)
					}
				}
				return nil
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			http.ServeFile(w, r, path)
			return
		}

		pageData := PageData{
			Path:  path,
			Dirs:  dirs,
			Files: files,
			Query: query,
		}
		renderTemplate(w, "index", pageData)
	} else {
		//fmt.Println("Browse::Path=", path)
		fileInfo, err := os.Stat(path)
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}

		var dirs []FileInfo
		var files []FileInfo

		if fileInfo.IsDir() {
			entries, err := os.ReadDir(path)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			for _, entry := range entries {
				info, err := entry.Info()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				if query == "" || strings.Contains(info.Name(), query) {
					file := FileInfo{
						Name:  info.Name(),
						Path:  filepath.Join(path, info.Name()),
						IsDir: info.IsDir(),
					}
					if info.IsDir() {
						dirs = append(dirs, file)
					} else {
						files = append(files, file)
					}
				}
			}
		} else {
			http.ServeFile(w, r, path)
			return
		}

		pageData := PageData{
			Path:  path,
			Dirs:  dirs,
			Files: files,
			Query: query,
		}
		renderTemplate(w, "index", pageData)
	}

}

func main() {
	http.HandleFunc("/", fileHandler)
	http.ListenAndServe(":8080", nil)
}
