package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// Структура для передачи данных в HTML-шаблон
type PageData struct {
	NumbersInput string
	HasResult    bool
	Numbers      []int
	Sum1         int
	Sum2         int
	Total        int
	Error        string
}

// sumSlice — горутина, которая считает сумму своей половины среза
func sumSlice(numbers []int, ch chan int) {
	sum := 0
	for _, n := range numbers {
		sum += n
	}
	ch <- sum // отправляем результат в канал
}

// Обработчик главной страницы
func handler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		NumbersInput: "1, 2, 3, 4, 5, 6, 7, 8, 9, 10", // Значение по умолчанию
	}

	if r.Method == http.MethodPost {
		input := r.FormValue("numbers")
		data.NumbersInput = input

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
			data.Numbers = nums

			// --- ЛОГИКА ЗАДАЧИ 1 ---
			mid := len(nums) / 2
			ch := make(chan int)

			// Запускаем две горутины
			go sumSlice(nums[:mid], ch)
			go sumSlice(nums[mid:], ch)

			// Собираем результаты из канала
			data.Sum1 = <-ch
			data.Sum2 = <-ch
			data.Total = data.Sum1 + data.Sum2
		} else if len(nums) == 0 && data.Error == "" {
			data.Error = "Ошибка: введите хотя бы одно число."
		}
	}

	// HTML-шаблон (встроен прямо в код)
	tmpl := `
<!DOCTYPE html>
<html lang="ru">
<head>
	<meta charset="UTF-8">
	<title>Задача 1: Сумма через горутины</title>
	<style>
		body { font-family: system-ui, sans-serif; max-width: 600px; margin: 3rem auto; padding: 0 1rem; background: #f4f4f9; color: #333; }
		h1 { color: #1a1a1a; }
		.task { background: #fff; padding: 1rem; border-left: 5px solid #007acc; border-radius: 4px; margin-bottom: 1.5rem; }
		.form-group { display: flex; gap: 0.5rem; margin-bottom: 1rem; }
		input[type="text"] { flex: 1; padding: 10px; font-size: 16px; border: 1px solid #ccc; border-radius: 4px; font-family: monospace; }
		button { padding: 10px 20px; font-size: 16px; background: #007acc; color: white; border: none; border-radius: 4px; cursor: pointer; }
		button:hover { background: #005a9e; }
		.result { background: #e6ffed; padding: 1rem; border-left: 5px solid #34d058; border-radius: 4px; }
		.error { background: #ffeef0; padding: 1rem; border-left: 5px solid #d73a49; border-radius: 4px; color: #b31d28; }
		.mono { font-family: monospace; color: #555; }
	</style>
</head>
<body>
	<h1>Задача 1: Сумма через две горутины</h1>
	<div class="task">
		<p>Введите числа через запятую. Go-сервер запустит две горутины для подсчета суммы первой и второй половины, а затем соберет результаты через канал.</p>
	</div>

	<form method="POST">
		<div class="form-group">
			<input type="text" name="numbers" value="{{.NumbersInput}}" placeholder="Например: 1, 2, 3, 4, 5">
			<button type="submit">Посчитать</button>
		</div>
	</form>

	{{if .Error}}
		<div class="error"><strong>❌ {{.Error}}</strong></div>
	{{end}}

	{{if .HasResult}}
		<div class="result">
			<p><strong>Исходный срез:</strong> <span class="mono">{{.Numbers}}</span></p>
			<p><strong>Сумма первой половины (горутина 1):</strong> {{.Sum1}}</p>
			<p><strong>Сумма второй половины (горутина 2):</strong> {{.Sum2}}</p>
			<hr>
			<p><strong>Общая сумма:</strong> <span style="font-size: 1.2em; color: #1a7f37;">{{.Total}}</span></p>
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
	fmt.Println("Нажмите Ctrl+C для остановки.")
	log.Fatal(http.ListenAndServe(":8080", nil))
}