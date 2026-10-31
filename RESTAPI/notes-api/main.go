package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password,omitempty"` 
}

type Note struct {
	ID      int    `json:"id"`
	UserID  int    `json:"user_id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

var jwtSecret = []byte("super_secret_key_12345")

var (
	users      = make(map[int]User)
	notes      = make(map[int]Note)
	nextUserID = 1
	nextNoteID = 1
	dbMu       sync.RWMutex
)

func generateToken(user User) (string, error) {
	claims := jwt.MapClaims{
		"sub":   strconv.Itoa(user.ID), 
		"email": user.Email,            
		"exp":   time.Now().Add(15 * time.Minute).Unix(),
		"iat":   time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Требуется токен авторизации", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Неверный формат заголовка Authorization", http.StatusUnauthorized)
			return
		}

		tokenStr := parts[1]
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("неожиданный метод подписи")
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Недействительный токен", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Ошибка чтения токена", http.StatusUnauthorized)
			return
		}

		subStr, ok := claims["sub"].(string)
		if !ok {
			http.Error(w, "Неверный ID в токене", http.StatusUnauthorized)
			return
		}
		userID, _ := strconv.Atoi(subStr)

		ctx := context.WithValue(r.Context(), "userID", userID)
		next(w, r.WithContext(ctx))
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный JSON", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email и пароль обязательны", http.StatusBadRequest)
		return
	}

	dbMu.Lock()
	defer dbMu.Unlock()

	for _, u := range users {
		if u.Email == req.Email {
			http.Error(w, "Пользователь с таким email уже существует", http.StatusConflict)
			return
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Ошибка хеширования пароля", http.StatusInternalServerError)
		return
	}

	user := User{
		ID:       nextUserID,
		Email:    req.Email,
		Password: string(hash),
	}
	users[nextUserID] = user
	nextUserID++

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Пользователь успешно создан"})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный JSON", http.StatusBadRequest)
		return
	}

	dbMu.RLock()
	defer dbMu.RUnlock()

	var foundUser User
	for _, u := range users {
		if u.Email == req.Email {
			foundUser = u
			break
		}
	}

	if foundUser.ID == 0 {
		http.Error(w, "Неверный email или пароль", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(req.Password)); err != nil {
		http.Error(w, "Неверный email или пароль", http.StatusUnauthorized)
		return
	}

	token, err := generateToken(foundUser)
	if err != nil {
		http.Error(w, "Ошибка создания токена", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func getNotesHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)

	dbMu.RLock()
	defer dbMu.RUnlock()

	var userNotes []Note
	for _, n := range notes {
		if n.UserID == userID {
			userNotes = append(userNotes, n)
		}
	}

	if userNotes == nil {
		userNotes = []Note{} 
	}

	json.NewEncoder(w).Encode(userNotes)
}

func createNoteHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)

	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный JSON", http.StatusBadRequest)
		return
	}

	dbMu.Lock()
	defer dbMu.Unlock()

	note := Note{
		ID:      nextNoteID,
		UserID:  userID,
		Title:   req.Title,
		Content: req.Content,
	}
	notes[nextNoteID] = note
	nextNoteID++

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(note)
}

func getNoteByIDHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)
	idStr := r.PathValue("id") 
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}

	dbMu.RLock()
	defer dbMu.RUnlock()

	note, exists := notes[id]
	if !exists {
		http.Error(w, "Заметка не найдена", http.StatusNotFound)
		return
	}

	if note.UserID != userID {
		http.Error(w, "Доступ запрещен", http.StatusForbidden)
		return
	}

	json.NewEncoder(w).Encode(note)
}

func deleteNoteHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}

	dbMu.Lock()
	defer dbMu.Unlock()

	note, exists := notes[id]
	if !exists {
		http.Error(w, "Заметка не найдена", http.StatusNotFound)
		return
	}

	if note.UserID != userID {
		http.Error(w, "Доступ запрещен", http.StatusForbidden)
		return
	}

	delete(notes, id)
	w.WriteHeader(http.StatusNoContent)
}



func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /register", registerHandler)
	mux.HandleFunc("POST /login", loginHandler)

	mux.HandleFunc("GET /notes", authMiddleware(getNotesHandler))
	mux.HandleFunc("POST /notes", authMiddleware(createNoteHandler))
	mux.HandleFunc("GET /notes/{id}", authMiddleware(getNoteByIDHandler))
	mux.HandleFunc("DELETE /notes/{id}", authMiddleware(deleteNoteHandler))

	log.Println("API сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}