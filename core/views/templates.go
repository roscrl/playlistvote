package views

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"app/config"
	"github.com/fsnotify/fsnotify"
)

const (
	DirTemplate      = "templates"
	DirTemplateSlash = DirTemplate + "/"

	PathTemplates = "core/views/" + DirTemplate
	PathViews     = "core/views"
)

var PathConfigDevBrowser = ""

func init() {
	_, filename, _, _ := runtime.Caller(0)
	PathConfigDevBrowser = filepath.Dir(filename) + "/.dev.browser.mock"
}

//go:embed templates/*.tmpl templates/**/*.tmpl
var tmplFS embed.FS

type Views struct {
	env config.Environment

	templates *template.Template
	funcMap   template.FuncMap
}

func New(env config.Environment) *Views {
	funcMap := template.FuncMap{
		"formatNumberInK": func(num int64) string {
			const K = 1000
			if num >= K {
				quotient := num / K

				return fmt.Sprintf("%dk", quotient)
			}

			return fmt.Sprintf("%d", num)
		},
		"safeURL": func(s string) template.URL {
			return template.URL(s) //nolint:gosec
		},
		"rawHTML": func(s string) template.HTML {
			return template.HTML(s) //nolint:gosec
		},
	}
	views := &Views{env: env, funcMap: funcMap}

	if env == config.DEV {
		templates := findAndParseTemplates(os.DirFS(PathTemplates), funcMap)

		views.templates = templates
		watchDevTemplates(views)
	} else {
		tmplFS, err := fs.Sub(tmplFS, DirTemplate)
		if err != nil {
			log.Fatal(err)
		}

		templates := findAndParseTemplates(tmplFS, funcMap)

		views.templates = templates
	}

	log.Println(views.templates.DefinedTemplates())

	return views
}

func (v *Views) Stream(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", TurboStreamMIME)
	v.Render(w, name, data)
}

func (v *Views) Render(w io.Writer, name string, data any) {
	tmpl := template.Must(v.templates.Clone())

	if v.env == config.DEV {
		tmpl = template.Must(tmpl.ParseGlob(PathTemplates + "/" + name))
	} else {
		tmpl = template.Must(tmpl.ParseFS(tmplFS, DirTemplateSlash+name))
	}

	err := tmpl.ExecuteTemplate(w, name, data)
	if err != nil {
		if strings.Contains(err.Error(), "broken pipe") {
			return
		}

		log.Printf("failed to render template %s: %v, defined templates %v", name, err, tmpl.DefinedTemplates())
		_ = tmpl.ExecuteTemplate(w, "error.tmpl", err)
	}
}

func (v *Views) RenderStandardError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	v.Render(w, "error.tmpl", map[string]any{})
}

func (v *Views) RenderError(w http.ResponseWriter, msg string, statusCode int) {
	w.WriteHeader(statusCode)
	v.Render(w, "error.tmpl", map[string]any{"error": msg})
}

func findAndParseTemplates(filesys fs.FS, funcMap template.FuncMap) *template.Template {
	rootTemplate := template.New("")

	root, err := filesys.Open(".")
	if err != nil {
		log.Fatal(err)
	}

	rootStat, err := root.Stat()
	if err != nil {
		log.Fatal(err)
	}

	err = fs.WalkDir(tmplFS, rootStat.Name(), func(path string, info fs.DirEntry, err error) error {
		if !info.IsDir() && strings.HasSuffix(path, ".tmpl") {
			if err != nil {
				return err
			}

			strippedTemplateDirTemplatePath := strings.TrimPrefix(path, DirTemplateSlash)

			templateContent, err := fs.ReadFile(filesys, strippedTemplateDirTemplatePath)
			if err != nil {
				return err
			}

			tmpl := rootTemplate.New(strippedTemplateDirTemplatePath).Funcs(funcMap)
			_, err = tmpl.Parse(string(templateContent))
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	return rootTemplate
}

func watchDevTemplates(views *Views) {
	watcher, err := fsnotify.NewWatcher() // leaks but only used for dev
	if err != nil {
		log.Fatal(err)
	}

	addWatchers := func(path string) error {
		err := watcher.Add(path)
		if err != nil {
			return err
		}

		// Walk the directory and add watchers for subdirectories
		return filepath.Walk(path, func(subpath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return watcher.Add(subpath)
			}

			return nil
		})
	}

	err = addWatchers("./" + PathViews)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("watching templates")

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if event.Has(fsnotify.Chmod) {
					continue
				}

				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
					if !strings.HasSuffix(event.Name, ".tmpl") {
						continue
					}

					log.Printf("%s changed %s, reloading~", event.Name, event.Op)

					templates := findAndParseTemplates(os.DirFS(PathTemplates), views.funcMap)

					views.templates = templates // not thread safe but only used for dev
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}

				log.Fatal(err)
			}
		}
	}()
}
