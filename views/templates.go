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
	"strings"

	"app/services/spotify"

	"app/config"

	"github.com/fsnotify/fsnotify"
)

const (
	TemplateDir      = "templates"
	TemplateDirSlash = TemplateDir + "/"

	TemplatesPath = "views/" + TemplateDir
)

//go:embed templates/*.tmpl templates/**/*.tmpl
var tmplFS embed.FS

type Views struct {
	env config.Environment

	templates *template.Template
	funcMap   template.FuncMap
}

func New(env config.Environment) *Views {
	funcMap := template.FuncMap{
		"formatNumberInK": func(n int64) string {
			if n >= 1000 {
				quotient := n / 1000
				return fmt.Sprintf("%dk", quotient)
			}
			return fmt.Sprintf("%d", n)
		},
		"safeURL": func(s string) template.URL {
			return template.URL(s)
		},
		"rawHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"stripSpotifyURI": func(s string) string {
			return strings.TrimPrefix(s, spotify.URIPlaylistPrefix)
		},
	}
	views := &Views{env: env, funcMap: funcMap}

	if env == config.DEV {
		templates, err := findAndParseTemplates(os.DirFS(TemplatesPath), funcMap)
		if err != nil {
			log.Fatal(err)
		}

		views.templates = templates
		watchDevTemplates(views)
	} else {
		tmplFS, err := fs.Sub(tmplFS, TemplateDir)
		if err != nil {
			log.Fatal(err)
		}

		templates, err := findAndParseTemplates(tmplFS, funcMap)
		if err != nil {
			log.Fatal(err)
		}

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
		tmpl = template.Must(tmpl.ParseGlob(TemplatesPath + "/" + name))
	} else {
		tmpl = template.Must(tmpl.ParseFS(tmplFS, TemplateDirSlash+name))
	}

	err := tmpl.ExecuteTemplate(w, name, data)
	if err != nil {
		log.Printf("failed to render template %s: %v, defined templates %v", name, err, tmpl.DefinedTemplates())
		tmpl.ExecuteTemplate(w, "error.tmpl", err)
	}
}

func (v *Views) RenderError(w io.Writer, msg string) {
	if msg != "" {
		v.Render(w, "error.tmpl", map[string]any{"error": msg})
	} else {
		v.Render(w, "error.tmpl", map[string]any{})
	}
}

func findAndParseTemplates(FS fs.FS, funcMap template.FuncMap) (*template.Template, error) {
	rootTemplate := template.New("")
	root, err := FS.Open(".")
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

			strippedTemplateDirTemplatePath := strings.TrimPrefix(path, TemplateDirSlash)

			templateContent, err := fs.ReadFile(FS, strippedTemplateDirTemplatePath)
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

	return rootTemplate, nil
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

	err = addWatchers("./views")
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
					templates, err := findAndParseTemplates(os.DirFS(TemplatesPath), views.funcMap)
					if err != nil {
						log.Fatal(err)
					}
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
