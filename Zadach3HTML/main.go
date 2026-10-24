package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// Структура для передачи данных в HTML
type PageData struct {
	InputNumbers string
	HasResult    bool
	Original     []int
	Processed    []int
	Error        string
}

// --- ЛОГИКА ЗАДАЧИ 3 ---

// worker — воркер, который читает числа из jobs,
// умножает их на 2 и пишет в results.
func worker(id int, jobs <-chan int, results chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	for num := range jobs {
		// Имитируем небольшую задержку для наглядности параллельной работы
		// (в реальном коде здесь может быть сложная логика)
		results <- num * 2
	}
}

// Обработчик HTTP-запросов
func handler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		InputNumbers: "1, 2, 3, 4, 5, 6, 7, 8, 9, 10", // Значение по умолчанию
	}

	if r.Method == http.MethodPost {
		input := r.FormValue("numbers")
		data.InputNumbers = input

		// Парсим строку в срез чисел
		var nums []int
		parts := strings.Split(input, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			num, err := strconv.Atoi(p)
			if err != nil {
				data.Error = "Ошибка: вводите только числа, разделенные запятой."
				break
			}
			nums = append(nums, num)
		}

		if data.Error == "" && len(nums) > 0 {
			data.HasResult = true
			data.Original = nums

			// --- ЗАПУСК WORKER POOL ---
			jobs := make(chan int, len(nums))
			results := make(chan int, len(nums))

			const numWorkers = 3
			var wg sync.WaitGroup

			// Запускаем 3 воркера
			for i := 1; i <= numWorkers; i++ {
				wg.Add(1)
				go worker(i, jobs, results, &wg)
			}

			// Отправляем задачи в канал jobs
			for _, n := range nums {
				jobs <- n
			}
			close(jobs) // Важно: закрываем канал задач, чтобы воркеры завершились

			// Ждем завершения всех воркеров в отдельной горутине, затем закрываем results
			go func() {
				wg.Wait()
				close(results)
			}()

			// Собираем результаты в срез
			for res := range results {
				data.Processed = append(data.Processed, res)
			}

			// Для красоты на сайте отсортируем результаты,
			// так как воркеры завершаются в случайном порядке.
			sort.Ints(data.Processed)
		} else if len(nums) == 0 && data.Error == "" {
			data.Error = "Ошибка: введите хотя бы одно число."
		}
	}

	// HTML-шаблон
	tmpl := `
<!DOCTYPE html>
<html lang="ru">
<head>
	<meta charset="UTF-8">
	<title>Задача 3: Пул из 3 воркеров</title>
	<style>
		body { font-family: system-ui, sans-serif; max-width: 700px; margin: 3rem auto; padding: 0 1rem; background: #f4f4f9; color: #333; }
		h1 { color: #1a1a1a; }
		.task { background: #fff; padding: 1rem; border-left: 5px solid #007acc; border-radius: 4px; margin-bottom: 1.5rem; }
		.form-group { display: flex; gap: 0.5rem; margin-bottom: 1rem; }
		input[type="text"] { flex: 1; padding: 10px; font-size: 16px; border: 1px solid #ccc; border-radius: 4px; font-family: monospace; }
		button { padding: 10px 20px; font-size: 16px; background: #007acc; color: white; border: none; border-radius: 4px; cursor: pointer; }
		button:hover { background: #005a9e; }
		.result { background: #e6ffed; padding: 1rem; border-left: 5px solid #34d058; border-radius: 4px; margin-bottom: 1rem; }
		.error { background: #ffeef0; padding: 1rem; border-left: 5px solid #d73a49; border-radius: 4px; color: #b31d28; margin-bottom: 1rem; }
		.mono { font-family: monospace; }
		.badge { background: #007acc; color: white; padding: 2px 8px; border-radius: 12px; font-size: 12px; margin-left: 5px; vertical-align: middle; }
		.squares-list { display: flex; flex-wrap: wrap; gap: 8px; margin-top: 1rem; }
		.square-item { background: #fff; border: 1px solid #007acc; padding: 5px 10px; border-radius: 4px; font-family: monospace; }
	</style>
</head>
<body>
	<h1>Задача 3: Пул из 3 воркеров <span class="badge">Worker Pool</span></h1>
	<div class="task">
		<p><strong>Условие:</strong> Есть срез чисел. Запущено 3 воркера. Каждый получает числа из канала задач, умножает на 2 и отправляет в канал результатов. Main собирает результаты в срез.</p>
	</div>

	<form method="POST">
		<div class="form-group">
			<input type="text" name="numbers" value="{{.InputNumbers}}" placeholder="Числа через запятую">
			<button type="submit">Запустить воркеры</button>
		</div>
	</form>

	{{if .Error}}
		<div class="error"><strong>❌ {{.Error}}</strong></div>
	{{end}}

	{{if .HasResult}}
		<div class="result">
			<p><strong>Исходные числа:</strong> <br><span class="mono">{{.Original}}</span></p>
			<hr>
			<p><strong>Результаты (число × 2):</strong> <br>
			<small><em>(отсортированы для удобства, так как воркеры работают параллельно)</em></small></p>
			<div class="squares-list">
				{{range .Processed}}
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