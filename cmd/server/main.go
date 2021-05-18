package main

import (
	"context"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/foolin/goview"
	"github.com/goji/httpauth"
	"github.com/greboid/fileshare"
	"github.com/kouhin/envflag"
	"github.com/tidwall/buntdb"
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
	dbFile        = flag.String("db-file", "data/meta.db", "Path to the meta database")
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
	db, err := buntdb.Open(*dbFile)
	if err != nil {
		log.Fatalf("Unable to open the database: %s", err.Error())
	}
	defer func() {
		_ = db.Close()
	}()
	backgroundPrune(db)
	authOptions := httpauth.AuthOptions{
		AuthFunc: authFunc(os.Getenv("API-KEY")),
	}
	initTemplates()
	router := goji.NewMux()
	upload := goji.SubMux()
	files := goji.SubMux()

	upload.Use(httpauth.BasicAuth(authOptions))
	upload.HandleFunc(pat.Post("/file"), handleUpload(db))

	files.Use(checkExpiry(db))
	files.Handle(pat.New("/*"), http.StripPrefix("/raw", http.FileServer(http.Dir(filepath.Join(*workDir, "raw")))))

	router.Use(fileshare.LoggingHandler(os.Stdout))
	router.Use(fileshare.StripSlashes)
	router.HandleFunc(pat.Get("/"), handleIndex)
	router.Handle(pat.New("/upload/*"), upload)
	router.Handle(pat.Get("/static/*"), http.StripPrefix("/static", http.FileServer(http.FS(staticFiles))))
	router.Handle(pat.Get("/raw/*"), files)

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
	if err = server.Shutdown(ctx); err != nil {
		log.Fatalf("Unable to shutdown: %s", err.Error())
	}
	log.Print("Finishing server.")
}

func checkExpiry(db *buntdb.DB) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ud, err := getFile(db, strings.TrimPrefix(r.URL.Path, "/raw/"))
			if err == nil {
				checkFile(db, *ud)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func backgroundPrune(db *buntdb.DB) {
	ticker := time.NewTicker(1 * time.Minute)
	pruneFiles(db)
	go func() {
		for {
			select {
			case <-ticker.C:
				pruneFiles(db)
			}
		}
	}()
}

func pruneFiles(db *buntdb.DB) {
	var uploads []fileshare.UploadDescription
	_ = db.View(func(tx *buntdb.Tx) error {
		err := tx.Ascend("", func(key, value string) bool {
			ud := fileshare.UploadDescription{}
			_ = json.Unmarshal([]byte(value), &ud)
			uploads = append(uploads, ud)
			return true
		})
		return err
	})
	for index := range uploads {
		checkFile(db, uploads[index])
	}
}

func getFile(db *buntdb.DB, fullname string) (*fileshare.UploadDescription, error) {
	var dbValue string
	ud := &fileshare.UploadDescription{}
	var err error
	err = db.View(func(tx *buntdb.Tx) error {
		dbValue, err = tx.Get(fullname)
		return err
	})
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(dbValue), ud)
	if err != nil {
		return nil, err
	}
	return ud, nil
}

func checkFile(db *buntdb.DB, ud fileshare.UploadDescription) {
	if time.Now().After(ud.Expiry) {
		log.Printf("Removing: %s", ud.GetFullName())
		err := os.Remove(filepath.Join(*workDir, "raw", ud.GetFullName()))
		if err != nil {
			log.Printf("Error removing file %s: %s", ud.GetFullName(), err.Error())
		}
		_ = db.Update(func(tx *buntdb.Tx) error {
			_, err = tx.Delete(ud.GetFullName())
			return err
		})
	} else {
		log.Printf("Not removing: %s", ud.GetFullName())
	}
}

func authFunc(key string) func(string, string, *http.Request) bool {
	if key == "" {
		return func(string, string, *http.Request) bool {
			return true
		}
	}
	return func(_ string, password string, request *http.Request) bool {
		key := request.Header.Get("X-API-KEY")
		if key == "meh" || password == "meh" {
			return true
		}
		return false
	}
}

func handleUpload(db *buntdb.DB) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
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
		expiry := request.FormValue("expiry")
		if expiry != "0" {
			duration, err := time.ParseDuration(expiry)
			if err == nil {
				ud.Expiry = time.Now().Add(duration)
			}
		}
		ud.Extension = filepath.Ext(handler.Filename)
		ud.Size = handler.Size
		jsonData, err := json.Marshal(ud)
		err = db.Update(func(tx *buntdb.Tx) error {
			_, _, err = tx.Set(ud.GetFullName(), string(jsonData), nil)
			return err
		})
		if err != nil {
			log.Printf("Upload failed: unable to write meta: %v", err)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		jsonData, err = ud.GetJSON()
		if err != nil {
			log.Printf("Upload failed: couldn't get file data: %v", err)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		data, err := ioutil.ReadAll(file)
		if err != nil {
			log.Printf("Upload failed: couldn't read file: %v", err)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		err = os.WriteFile(filepath.Join(*workDir, "raw", ud.GetFullName()), data, 0644)
		if err != nil {
			log.Printf("Upload failed: couldn't write file: %v", err)
			writer.WriteHeader(http.StatusBadRequest)
		}
		_, _ = writer.Write(jsonData)
	}
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
