package main

import (
	"io/ioutil"
	"net/http"
	"html/template"
	"regexp"
	"errors"
	"log"
	"strings"
)

// wikiのデータ構造
type Page struct {
	Title string
	Body  []byte
}

// パスのアドレスを設定して文字の長さを定数として持つ
const lenPath = len("/view/")

//テンプレートファイルの配列を作成
var templates = make(map[string]*template.Template)

var titleValidator = regexp.MustCompile("^[a-zA-Z0-9]+$")

const expend_string = ".txt"

//初期化関数
func init() {
	for _, tmpl := range[]string{"edit", "view"}{
		t := template.Must(template.ParseFiles(tmpl + ".html"))
		templates[tmpl] = t
	}
}

func getTitle(w http.ResponseWriter, r *http.Request)(title string, err error){
	title = r.URL.Path[lenPath:]
	if !titleValidator.MatchString(title){
		http.NotFound(w, r)
		err = errors.New("Invalid Page Title")
		log.Print(err)
	}
	return
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	title, err := getTitle(w, r)
	if err != nil {
		return
	}
	p, err := loadPage(title)
	if err != nil {
		//編集ページに飛ばす
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	title, err := getTitle(w, r)
	if err != nil {
		return
	}
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	title, err := getTitle(w, r)
	if err != nil {
		return
	}
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err = p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func topHandler(w http.ResponseWriter, r *http.Request) {
	files, err := ioutil.ReadDir("./")
	if err != nil {
		err = errors.New("所定のディレクトリ内にテキストファイルがありません")
		log.Print(err)
		return
	}

	var paths []string
	var fileName []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), expend_string) {
			fileName = strings.Split(string(file.Name()), expend_string)
			paths = append(paths, fileName[0])
		}
	}

	if paths == nil {
		err = errors.New("テキストファイルが存在しません")
		log.Print(err)
	}

	t := template.Must(template.ParseFiles("top.html"))
	err = t.Execute(w, paths)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		title := r.URL.Path[lenPath:]
		if !titleValidator.MatchString(title){
			http.NotFound(w, r)
			err := errors.New("Invalid Page Title")
			log.Print(err)
			return
		}
		fn(w, r, title)
	}
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates[tmpl].Execute(w, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// テキストファイルの保存メソッド
func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)

	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func main() {
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/top/", topHandler)
	http.ListenAndServe(":8082", nil)
}
