package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

var (
	usersMu sync.RWMutex
	users   []User
	tmpl    *template.Template
)


var allowedDirs = map[string]bool{
	"text":  true,
	"photo": true,
	"other": true,
}

func safePath(dir, name string) (string, error) {
	if !allowedDirs[dir] {
		return "", errors.New("недопустимая папка")
	}
	name = filepath.Base(name)
	if name == "." || name == ".." || name == "/" || name == "" {
		return "", errors.New("недопустимое имя файла")
	}
	full := filepath.Join("storage", dir, name)
	absRoot, _ := filepath.Abs("storage")
	absFull, _ := filepath.Abs(full)
	if !strings.HasPrefix(absFull, absRoot+string(os.PathSeparator)) {
		return "", errors.New("путь за пределами хранилища")
	}
	return full, nil
}


var (
	textExts  = map[string]bool{".txt": true, ".md": true, ".pdf": true, ".doc": true, ".docx": true, ".csv": true}
	photoExts = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
)

func categoryOf(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if textExts[ext] {
		return "text"
	}
	if photoExts[ext] {
		return "photo"
	}
	return "other"
}


func loadUsers(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	usersMu.Lock()
	defer usersMu.Unlock()
	err = json.Unmarshal(data, &users)
	if err != nil {
		return err
	}

	for i := range users {
		if !strings.HasPrefix(users[i].Password, "$2a$") {
			hash, _ := bcrypt.GenerateFromPassword([]byte(users[i].Password), bcrypt.DefaultCost)
			users[i].Password = string(hash)
			log.Printf("Сгенерирован хеш для пользователя: %s", users[i].Login)
		}
	}
	return nil
}

func findUser(login string) (User, bool) {
	usersMu.RLock()
	defer usersMu.RUnlock()
	for _, u := range users {
		if u.Login == login {
			return u, true
		}
	}
	return User{}, false
}

type Session struct {
	Login string
	Role  string
}

var (
	sessMu   sync.RWMutex
	sessions = map[string]Session{}
)

func newToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func setSession(w http.ResponseWriter, login, role string) {
	token := newToken()
	sessMu.Lock()
	sessions[token] = Session{Login: login, Role: role}
	sessMu.Unlock()
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func getSession(r *http.Request) (Session, bool) {
	c, err := r.Cookie("session")
	if err != nil {
		return Session{}, false
	}
	sessMu.RLock()
	defer sessMu.RUnlock()
	s, ok := sessions[c.Value]
	return s, ok
}

func clearSession(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie("session"); err == nil {
		sessMu.Lock()
		delete(sessions, c.Value)
		sessMu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{
		Name: "session", Value: "", Path: "/", MaxAge: -1,
	})
}


type handlerFunc func(w http.ResponseWriter, r *http.Request, s Session)

func auth(next handlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s, ok := getSession(r)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r, s)
	}
}

func authView(next handlerFunc) http.HandlerFunc {
	return auth(func(w http.ResponseWriter, r *http.Request, s Session) {
		if s.Role != "viewer" && s.Role != "admin" {
			http.Error(w, "403 Доступ запрещен (только просмотр)", http.StatusForbidden)
			return
		}
		next(w, r, s)
	})
}

func authDownload(next handlerFunc) http.HandlerFunc {
	return auth(func(w http.ResponseWriter, r *http.Request, s Session) {
		if s.Role != "downloader" && s.Role != "admin" {
			http.Error(w, "403 Доступ запрещен (только скачивание)", http.StatusForbidden)
			return
		}
		next(w, r, s)
	})
}

func authAdmin(next handlerFunc) http.HandlerFunc {
	return auth(func(w http.ResponseWriter, r *http.Request, s Session) {
		if s.Role != "admin" {
			http.Error(w, "403 Доступ запрещен (только админ)", http.StatusForbidden)
			return
		}
		next(w, r, s)
	})
}


func loginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl.ExecuteTemplate(w, "login.html", nil)
		return
	}
	login := r.FormValue("login")
	password := r.FormValue("password")
	u, ok := findUser(login)
	if !ok || bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) != nil {
		tmpl.ExecuteTemplate(w, "login.html", map[string]string{"Error": "Неверный логин или пароль"})
		return
	}
	setSession(w, u.Login, u.Role)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func logout(w http.ResponseWriter, r *http.Request) {
	clearSession(w, r)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func home(w http.ResponseWriter, r *http.Request, s Session) {
	dirs := []map[string]any{}
	for _, d := range []string{"text", "photo", "other"} {
		entries, _ := os.ReadDir(filepath.Join("storage", d))
		dirs = append(dirs, map[string]any{
			"Name":  d,
			"Count": len(entries),
		})
	}
	tmpl.ExecuteTemplate(w, "browse.html", map[string]any{
		"User": s,
		"Dirs": dirs,
	})
}

func browse(w http.ResponseWriter, r *http.Request, s Session) {
	dir := r.URL.Query().Get("dir")
	if !allowedDirs[dir] {
		http.Error(w, "Папка не найдена", 404)
		return
	}
	entries, err := os.ReadDir(filepath.Join("storage", dir))
	if err != nil {
		http.Error(w, "Ошибка чтения папки", 500)
		return
	}
	files := []map[string]any{}
	for _, e := range entries {
		if !e.IsDir() {
			info, _ := e.Info()
			files = append(files, map[string]any{
				"Name": e.Name(),
				"Size": info.Size(),
			})
		}
	}
	tmpl.ExecuteTemplate(w, "files.html", map[string]any{
		"User": s,
		"Dir":  dir,
		"Files": files,
	})
}

func viewFile(w http.ResponseWriter, r *http.Request, s Session) {
	dir := r.URL.Query().Get("dir")
	name := r.URL.Query().Get("file")
	path, err := safePath(dir, name)
	if err != nil {
		http.Error(w, "Недопустимый путь", 400)
		return
	}
	w.Header().Set("Content-Disposition", "inline; filename=\""+name+"\"")
	http.ServeFile(w, r, path)
}

func downloadFile(w http.ResponseWriter, r *http.Request, s Session) {
	dir := r.URL.Query().Get("dir")
	name := r.URL.Query().Get("file")
	path, err := safePath(dir, name)
	if err != nil {
		http.Error(w, "Недопустимый путь", 400)
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename=\""+name+"\"")
	http.ServeFile(w, r, path)
}

func uploadFile(w http.ResponseWriter, r *http.Request, s Session) {
	if r.Method == "GET" {
		tmpl.ExecuteTemplate(w, "upload.html", map[string]any{"User": s})
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 50<<20) // 50 MB limit
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Файл слишком большой или произошла ошибка", 400)
		return
	}
	defer file.Close()

	category := categoryOf(header.Filename)
	safeName := filepath.Base(header.Filename)
	if safeName == "." || safeName == ".." || safeName == "" {
		safeName = "uploaded_file"
	}
	fullPath := filepath.Join("storage", category, safeName)
	
	if _, err := os.Stat(fullPath); err == nil {
		safeName = "copy_" + safeName
		fullPath = filepath.Join("storage", category, safeName)
	}

	out, err := os.Create(fullPath)
	if err != nil {
		http.Error(w, "Ошибка сохранения файла", 500)
		return
	}
	defer out.Close()

	_, err = out.ReadFrom(file)
	if err != nil {
		http.Error(w, "Ошибка записи файла", 500)
		return
	}
	http.Redirect(w, r, "/browse?dir="+category, http.StatusSeeOther)
}


func main() {
	if err := loadUsers("users.json"); err != nil {
		log.Fatal("Ошибка загрузки users.json: ", err)
	}

	var err error
	tmpl, err = template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatal("Ошибка загрузки шаблонов: ", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/login", loginPage)
	mux.HandleFunc("/logout", logout)
	
	mux.HandleFunc("/", auth(home))
	mux.HandleFunc("/browse", auth(browse))
	mux.HandleFunc("/view", authView(viewFile))
	mux.HandleFunc("/download", authDownload(downloadFile))
	mux.HandleFunc("/upload", authAdmin(uploadFile))

	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}