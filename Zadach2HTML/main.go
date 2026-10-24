package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

// Структура для передачи данных в HTML
type PageData struct {
	InputN    string
	HasResult bool
	N         int
	Squares   []int
	Error     string
}

// --- ЛОГИКА ЗАДАЧИ 2 ---

// generate — первая горутина: генерирует числа от 1 до N и пишет их в канал out
func generate(n int, out chan<- int) {
	for i := 1; i <= n; i++ {
		out <- i
	}
	close(out) // Закрываем канал, чтобы вторая горутина поняла, что данные закончились
}

// square — вторая горутина: читает числа из in, возводит в квадрат и пишет в out
func square(in <-chan int, out chan<- int) {
	for num := range in {
		out <- num * num
	}
	close(out) // Закрываем выходной канал, чтобы main понял, что можно заканчивать
}

// Обработчик HTTP-запросов
func handler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		InputN: "10", // Значение по умолчанию
	}

	if r.Method == http.MethodPost {
		input := r.FormValue("n")
		data.InputN = input

		// Парсим N
		n, err := strconv.Atoi(input)
		if err != nil || n < 1 {
			data.Error = "Пожалуйста, введите корректное положительное число."
		} else if n > 1000 {
			data.Error = "Число слишком большое для наглядного вывода. Введите N до 1000."
		} else {
			data.HasResult = true
			data.N = n

			// Запускаем конвейер горутин
			numbers := make(chan int)
			squares := make(chan int)

			go generate(n, numbers)      // Горутина 1
			go square(numbers, squares)  // Горутина 2

			// Main собирает квадраты в срез
			for sq := range squares {
				data.Squares = append(data.Squares, sq)
			}
		}
	}

	// HTML-шаблон
	tmpl := `
<!DOCTYPE html>
<html lang="ru">
<head>
	<meta charset="UTF-8">
	<title>Задача 2: Конвейер горутин</title>
	<style>
		body { font-family: system-ui, sans-serif; max-width: 800px; margin: 3rem auto; padding: 0 1rem; background: #f4f4f9; color: #333; }
		h1 { color: #1a1a1a; }
		.task { background: #fff; padding: 1rem; border-left: 5px solid #007acc; border-radius: 4px; margin-bottom: 1.5rem; }
		.form-group { display: flex; gap: 0.5rem; margin-bottom: 1rem; }
		input[type="number"] { flex: 1; padding: 10px; font-size: 16px; border: 1px solid #ccc; border-radius: 4px; font-family: monospace; }
		button { padding: 10px 20px; font-size: 16px; background: #007acc; color: white; border: none; border-radius: 4px; cursor: pointer; }
		button:hover { background: #005a9e; }
		.result { background: #e6ffed; padding: 1rem; border-left: 5px solid #34d058; border-radius: 4px; }
		.error { background: #ffeef0; padding: 1rem; border-left: 5px solid #d73a49; border-radius: 4px; color: #b31d28; margin-bottom: 1rem; }
		.squares-list { display: flex; flex-wrap: wrap; gap: 8px; margin-top: 1rem; }
		.square-item { background: #fff; border: 1px solid #34d058; padding: 5px 10px; border-radius: 4px; font-family: monospace; }
	</style>
</head>
<body>
	<h1>Задача 2: Конвейер (генератор → квадратор)</h1>
	<div class="task">
		<p><strong>Условие:</strong> Первая горутина генерирует числа от 1 до N и пишет их в канал. Вторая горутина читает числа, возводит в квадрат и пишет в другой канал. Main собирает квадраты в срез.</p>
	</div>

	<form method="POST">
		<div class="form-group">
			<input type="number" name="n" value="{{.InputN}}" min="1" max="1000" placeholder="Введите N">
			<button type="submit">Сгенерировать</button>
		</div>
	</form>

	{{if .Error}}
		<div class="error"><strong>❌ {{.Error}}</strong></div>
	{{end}}

	{{if .HasResult}}
		<div class="result">
			<p><strong>Квадраты чисел от 1 до {{.N}}:</strong> (собрано из канала)</p>
			<div class="squares-list">
				{{range .Squares}}
					<span class="square-item">{{.}}</span>
				{{end}}
			</div>
		</div>
	{{end}}
</body>
</html>
`

	t, err := template.New("webpage").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, data)
}

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}