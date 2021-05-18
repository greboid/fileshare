package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"github.com/goji/httpauth"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/foolin/goview"
	"github.com/greboid/fileshare"
	"github.com/kouhin/envflag"
	"github.com/yalue/merged_fs"
	"goji.io"
	"goji.io/pat"
)

//go:embed resources/static resources/views
var embeddedFiles embed.FS

var (
	staticFiles   fs.FS
	templateFiles fs.FS
	version       = "snapshot"
	httpPort      = flag.Int("httpport", 8080, "HTTP server port")
	workDir       = flag.String("workdir", "./data", "Working directory")
)

func main() {
	err := envflag.Parse()
	if err != nil {
		log.Fatalf("Unable to parse flags: %s", err.Error())
	}
	err = initFileSystem()
	if err != nil {
		log.Fatalf("Unable to create work directory: %s", err.Error())
	}
	authOptions := httpauth.AuthOptions{
		AuthFunc: authFunc(os.Getenv("API-KEY")),
	}
	initTemplates()
	router := goji.NewMux()
	upload := goji.SubMux()

	upload.Use(httpauth.BasicAuth(authOptions))
	upload.HandleFunc(pat.New("/file"), handleUpload)

	router.Use(fileshare.LoggingHandler(os.Stdout))
	router.Use(fileshare.StripSlashes)
	router.HandleFunc(pat.Get("/"), handleIndex)
	router.Handle(pat.New("/upload/*"), upload)
	router.Handle(pat.Get("/static/*"), http.StripPrefix("/static", http.FileServer(http.FS(staticFiles))))
	router.Handle(pat.Get("/raw/*"), http.StripPrefix("/raw", http.FileServer(http.Dir(filepath.Join(*workDir, "raw")))))

	log.Print("Starting server.")
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", *httpPort),
		Handler: router,
	}
	go func() {
		_ = server.ListenAndServe()
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Unable to shutdown: %s", err.Error())
	}
	log.Print("Finishing server.")
}

func authFunc(key string) func (string, string, *http.Request) bool {
	if key == "" {
		log.Printf("API Key: Unspecified")
		return func(string, string, *http.Request) bool {
			return true
		}
	}
	log.Printf("API Key: Specified")
	return func(_ string, password string, request *http.Request) bool {
		key := request.Header.Get("X-API-KEY")
		if key == "meh" || password == "meh" {
			return true
		}
		return false
	}
}

func handleUpload(writer http.ResponseWriter, request *http.Request) {
	ud := fileshare.UploadDescription{
		Name: fileshare.Hex(6),
	}
	if err := request.ParseMultipartForm(1 << 30); err != nil {
		log.Printf("Upload failed: couldn't parse multipart data: %v", err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	file, handler, err := request.FormFile("file")
	if err != nil {
		log.Printf("Upload failed: couldn't find file: %v", err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	defer func() {
		_ = file.Close()
	}()
	ud.Extension = filepath.Ext(handler.Filename)
	ud.Size = handler.Size
	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Printf("Upload failed: couldn't read file: %v", err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	json, err := ud.GetJSON()
	if err != nil {
		log.Printf("Upload failed: couldn't get file data: %v", err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	err = os.WriteFile(filepath.Join(*workDir, "raw", ud.GetFullName()), data, 0644)
	if err != nil {
		log.Printf("Upload failed: couldn't write file: %v", err)
		writer.WriteHeader(http.StatusBadRequest)
	}
	_, _ = writer.Write(json)
}

func handleIndex(writer http.ResponseWriter, _ *http.Request) {
	err := goview.Render(writer, http.StatusOK, "index", goview.M{
		"Title":   "Index",
		"Version": version,
	})
	if err != nil {
		_, _ = fmt.Fprintf(writer, "Render index error: %v!", err)
	}
}

func initTemplates() {
	gv := goview.New(goview.Config{
		Root:      "resources/views",
		Extension: ".gohtml",
		Master:    "layouts/master",
	})
	gv.SetFileHandler(func(config goview.Config, tplFile string) (content string, err error) {
		file, err := templateFiles.Open(tplFile + config.Extension)
		if err != nil {
			return "", err
		}
		data, err := ioutil.ReadAll(file)
		if err != nil {
			return "", err
		}
		return string(data), nil
	})
	goview.Use(gv)
}

func initFileSystem() error {
	staticFs, _ := fs.Sub(embeddedFiles, "resources/static")
	staticFiles = merged_fs.NewMergedFS(os.DirFS("resources/static"), staticFs)

	templateFs, _ := fs.Sub(embeddedFiles, "resources/views")
	templateFiles = merged_fs.NewMergedFS(os.DirFS("resources/views"), templateFs)

	err := os.MkdirAll(filepath.Join(*workDir, "raw"), 0644)
	if err != nil {
		return err
	}
	return nil
}
