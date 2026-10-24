package main

import (
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type PageData struct {
	Gen1Input, Gen2Input, Gen3Input string
	HasResult                       bool
	Gen1, Gen2, Gen3, Merged        []int
	Error                           string
}

// --- ЛОГИКА ЗАДАЧИ 4 (FAN-IN) ---

// generator — горутина, которая пишет свой набор чисел в канал
// и закрывает его по завершении.
func generator(nums []int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			// Небольшая случайная задержка, чтобы показать,
			// что числа из разных каналов реально перемешиваются
			time.Sleep(time.Duration(rand.Intn(50)+10) * time.Millisecond)
			out <- n
		}
	}()
	return out
}

// merge — функция Fan-in. Принимает любое количество каналов,
// сливает их в один выходной канал.
func merge(channels ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup

	// Для каждого входного канала запускаем горутину-перекладыватель
	for _, ch := range channels {
		wg.Add(1)
		go func(c <-chan int) {
			defer wg.Done()
			for v := range c {
				out <- v
			}
		}(ch)
	}

	// Отдельная горутина закрывает out, когда все входные каналы опустеют
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// Вспомогательная функция для парсинга строки в срез чисел
func parseNums(input string) ([]int, error) {
	var nums []int
	parts := strings.Split(input, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		num, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("'%s' не является числом", p)
		}
		nums = append(nums, num)
	}
	return nums, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Gen1Input: "1, 2, 3, 4",
		Gen2Input: "10, 20, 30",
		Gen3Input: "100, 200, 300, 400, 500",
	}

	if r.Method == http.MethodPost {
		data.Gen1Input = r.FormValue("gen1")
		data.Gen2Input = r.FormValue("gen2")
		data.Gen3Input = r.FormValue("gen3")

		var err1, err2, err3 error
		data.Gen1, err1 = parseNums(data.Gen1Input)
		data.Gen2, err2 = parseNums(data.Gen2Input)
		data.Gen3, err3 = parseNums(data.Gen3Input)

		if err1 != nil || err2 != nil || err3 != nil {
			data.Error = "Ошибка ввода. Убедитесь, что везде введены числа через запятую."
		} else if len(data.Gen1)+len(data.Gen2)+len(data.Gen3) == 0 {
			data.Error = "Введите хотя бы одно число."
		} else {
			data.HasResult = true

			// Запускаем 3 независимых генератора
			ch1 := generator(data.Gen1)
			ch2 := generator(data.Gen2)
			ch3 := generator(data.Gen3)

			// Сливаем их в один канал (Fan-in)
			mergedCh := merge(ch1, ch2, ch3)

			// Main собирает всё в один срез
			for v := range mergedCh {
				data.Merged = append(data.Merged, v)
			}
		}
	}

	tmpl := `
<!DOCTYPE html>
<html lang="ru">
<head>
	<meta charset="UTF-8">
	<title>Задача 4: Fan-in</title>
	<style>
		body { font-family: system-ui, sans-serif; max-width: 800px; margin: 3rem auto; padding: 0 1rem; background: #f4f4f9; color: #333; }
		h1 { color: #1a1a1a; }
		.task { background: #fff; padding: 1rem; border-left: 5px solid #007acc; border-radius: 4px; margin-bottom: 1.5rem; }
		.grid { display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 10px; margin-bottom: 1rem; }
		.input-box { display: flex; flex-direction: column; gap: 5px; }
		input[type="text"] { padding: 8px; font-size: 14px; border: 1px solid #ccc; border-radius: 4px; font-family: monospace; }
		button { padding: 10px 20px; font-size: 16px; background: #007acc; color: white; border: none; border-radius: 4px; cursor: pointer; width: 100%; }
		button:hover { background: #005a9e; }
		.result { background: #e6ffed; padding: 1rem; border-left: 5px solid #34d058; border-radius: 4px; margin-top: 1rem; }
		.error { background: #ffeef0; padding: 1rem; border-left: 5px solid #d73a49; border-radius: 4px; color: #b31d28; margin-bottom: 1rem; }
		.mono { font-family: monospace; }
		.badge { background: #007acc; color: white; padding: 2px 8px; border-radius: 12px; font-size: 12px; margin-left: 5px; vertical-align: middle; }
		.merged-list { display: flex; flex-wrap: wrap; gap: 8px; margin-top: 10px; }
		.merged-item { background: #fff; border: 1px solid #f9c513; padding: 5px 10px; border-radius: 4px; font-family: monospace; color: #735c0f; }
		.channels { display: flex; justify-content: space-between; margin-bottom: 10px; color: #555; font-size: 14px; }
	</style>
</head>
<body>
	<h1>Задача 4: Fan-in (слияние каналов) <span class="badge">Merge</span></h1>
	<div class="task">
		<p><strong>Условие:</strong> Три независимых генератора выдают числа в свои каналы. Функция <code>merge</code> сливает их в один общий канал, а Main собирает все числа в один срез.</p>
	</div>

	<form method="POST">
		<div class="grid">
			<div class="input-box">
				<label>Генератор 1</label>
				<input type="text" name="gen1" value="{{.Gen1Input}}">
			</div>
			<div class="input-box">
				<label>Генератор 2</label>
				<input type="text" name="gen2" value="{{.Gen2Input}}">
			</div>
			<div class="input-box">
				<label>Генератор 3</label>
				<input type="text" name="gen3" value="{{.Gen3Input}}">
			</div>
		</div>
		<button type="submit">Слить в один канал</button>
	</form>

	{{if .Error}}
		<div class="error"><strong>❌ {{.Error}}</strong></div>
	{{end}}

	{{if .HasResult}}
		<div class="result">
			<p><strong>Входные данные:</strong></p>
			<div class="channels">
				<span>Канал 1: <span class="mono">{{.Gen1}}</span></span>
				<span>Канал 2: <span class="mono">{{.Gen2}}</span></span>
				<span>Канал 3: <span class="mono">{{.Gen3}}</span></span>
			</div>
			<hr>
			<p><strong>Объединенный канал (результат Fan-in):</strong> <br>
			<small><em>(порядок случайный, так как генераторы работают параллельно)</em></small></p>
			<div class="merged-list">
				{{range .Merged}}
					<span class="merged-item">{{.}}</span>
				{{end}}
			</div>
		</div>
	{{end}}
</body>
</html>
`
	t, _ := template.New("webpage").Parse(tmpl)
	t.Execute(w, data)
}

func main() {
	rand.Seed(time.Now().UnixNano()) // Для случайных задержек в генераторах
	http.HandleFunc("/", handler)
	fmt.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}