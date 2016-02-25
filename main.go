package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

type Conf struct {
	Java     string `envconfig:"JAVA"     default:"java"`
	Disunity string `envconfig:"DISUNITY" default:"./disunity.jar"`
	Convert  string `envconfig:"CONVERT"  default:"convert"`
	Addr     string `envconfig:"ADDR"     default:":19300"`
	TmpDir   string `envconfig:"TMP_DIR"  default:"/tmp"`
}

var conf Conf

func which(path string) bool {
	err := exec.Command("which", path).Run()
	return err == nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func init() {
	if err := envconfig.Process("disunity-app", &conf); err != nil {
		log.Println(err)
		os.Exit(1)
	}

	if !which(conf.Java) {
		log.Println("not found JAVA:", conf.Java)
		os.Exit(1)
	}

	if !exists(conf.Disunity) {
		log.Println("not found DISUNITY:", conf.Disunity)
		os.Exit(1)
	}

	if !which(conf.Convert) {
		log.Println("not found CONVERT:", conf.Convert)
		os.Exit(1)
	}
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			postHandleFunc(w, r)
		} else {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	})
	log.Fatal(http.ListenAndServe(conf.Addr, nil))
}

func postHandleFunc(w http.ResponseWriter, r *http.Request) {
	uploadFile, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer uploadFile.Close()

	tmpFile, err := ioutil.TempFile(conf.TmpDir, "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		tmpFile.Close()
		errLog(os.RemoveAll(tmpFile.Name() + ".unity3d"))
		errLog(os.RemoveAll(tmpFile.Name()))
	}()

	_, err = io.Copy(tmpFile, uploadFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpFile.Close(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := os.Rename(tmpFile.Name(), tmpFile.Name()+".unity3d"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := exec.Command(conf.Java, "-jar", conf.Disunity, "extract", tmpFile.Name()+".unity3d").Run(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !exists(tmpFile.Name()) {
		http.Error(w, "extract failed", http.StatusInternalServerError)
		return
	}

	var tgaPath string
	filepath.Walk(tmpFile.Name(), func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".tga") {
			tgaPath = path
			return err
		}
		return err
	})

	if tgaPath == "" {
		http.Error(w, "tga not found", http.StatusInternalServerError)
		return
	}

	pngPath := strings.TrimSuffix(tgaPath, ".tga") + ".png"

	if err := exec.Command(conf.Convert, tgaPath, pngPath).Run(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.ServeFile(w, r, pngPath)
}

func errLog(err error) {
	if err != nil {
		_, fn, line, _ := runtime.Caller(1)
		log.Printf("[error] %s:%d %v", fn, line, err)
	}
}
