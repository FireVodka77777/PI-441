package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type PageData struct {
	InputNumbers string
	HasResult    bool
	Results      []int
	Logs         []string
	Cancelled    bool
	Error        string
}

// Безопасный логгер для сбора сообщений из разных горутин
type SafeLogger struct {
	mu   sync.Mutex
	logs []string
}

func (l *SafeLogger) Log(msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logs = append(l.logs, msg)
}

func (l *SafeLogger) GetLogs() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.logs
}

// --- ЛОГИКА ЗАДАЧИ 5 ---

// process — имитация "тяжёлой" работы, которая умеет прерываться по контексту
func process(ctx context.Context, num int) (int, error) {
	select {
	case <-time.After(500 * time.Millisecond): // Имитация тяжелой работы (0.5 сек)
		return num * 5, nil
	case <-ctx.Done(): // Если пришла отмена, прерываем работу
		return 0, ctx.Err()
	}
}

// worker — воркер, который слушает контекст и обрабатывает задачи
func worker(
	id int,
	ctx context.Context,
	cancel context.CancelFunc,
	jobs <-chan int,
	results chan<- int,
	logger *SafeLogger,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			logger.Log(fmt.Sprintf("[Воркер %d] Получил сигнал отмены, выхожу.", id))
			return
		case num, ok := <-jobs:
			if !ok {
				logger.Log(fmt.Sprintf("[Воркер %d] Канал задач закрыт, выхожу.", id))
				return
			}

			logger.Log(fmt.Sprintf("[Воркер %d] Взял число: %d", id, num))
			
			res, err := process(ctx, num)
			if err != nil {
				logger.Log(fmt.Sprintf("[Воркер %d] Работа над %d прервана.", id, num))
				return
			}

			logger.Log(fmt.Sprintf("[Воркер %d] Результат для %d = %d", id, num, res))

			// Если результат превышает 50 — отменяем всю работу
			if res > 50 {
				logger.Log(fmt.Sprintf("⚠️ [Воркер %d] Результат %d > 50! ОТМЕНА ВСЕЙ РАБОТЫ!", id, res))
				cancel()
				return
			}

			// Отправляем результат (с проверкой на отмену)
			select {
			case results <- res:
			case <-ctx.Done():
				return
			}
		}
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		InputNumbers: "3, 7, 2, 8, 1, 9, 4, 11, 5, 6", // Число 11 * 5 = 55 вызовет отмену
	}

	if r.Method == http.MethodPost {
		input := r.FormValue("numbers")
		data.InputNumbers = input

		// Парсим числа
		var nums []int
		parts := strings.Split(input, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			num, err := strconv.Atoi(p)
			if err != nil {
				data.Error = "Ошибка: вводите только числа."
				break
			}
			nums = append(nums, num)
		}

		if data.Error == "" && len(nums) > 0 {
			data.HasResult = true
			logger := &SafeLogger{}

			// Создаем контекст с возможностью отмены
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel() // Обязательно вызываем в конце

			jobs := make(chan int)
			results := make(chan int, len(nums))

			// Запускаем РОВНО 2 воркера (не более 2 одновременно)
			var wg sync.WaitGroup
			const numWorkers = 2
			for i := 1; i <= numWorkers; i++ {
				wg.Add(1)
				go worker(i, ctx, cancel, jobs, results, logger, &wg)
			}

			// Producer: кладет задачи в канал, пока не отменят
			go func() {
				defer close(jobs)
				for _, n := range nums {
					select {
					case jobs <- n:
					case <-ctx.Done():
						logger.Log("[Producer] Отмена! Перестаю генерировать задачи.")
						return
					}
				}
			}()

			// Ждем завершения воркеров и закрываем канал результатов
			go func() {
				wg.Wait()
				close(results)
			}()

			// Main: собираем результаты
			for res := range results {
				data.Results = append(data.Results, res)
			}

			data.Logs = logger.GetLogs()
			if ctx.Err() == context.Canceled {
				data.Cancelled = true
			}
		}
	}

	tmpl := `
<!DOCTYPE html>
<html lang="ru">
<head>
	<meta charset="UTF-8">
	<title>Задача 5: Ограничение и Context</title>
	<style>
		body { font-family: system-ui, sans-serif; max-width: 800px; margin: 3rem auto; padding: 0 1rem; background: #f4f4f9; color: #333; }
		h1 { color: #1a1a1a; }
		.task { background: #fff; padding: 1rem; border-left: 5px solid #007acc; border-radius: 4px; margin-bottom: 1.5rem; }
		.form-group { display: flex; gap: 0.5rem; margin-bottom: 1rem; }
		input[type="text"] { flex: 1; padding: 10px; font-size: 16px; border: 1px solid #ccc; border-radius: 4px; font-family: monospace; }
		button { padding: 10px 20px; font-size: 16px; background: #007acc; color: white; border: none; border-radius: 4px; cursor: pointer; }
		button:hover { background: #005a9e; }
		.grid { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; margin-top: 1rem; }
		.panel { background: #fff; padding: 1rem; border-radius: 4px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
		.panel h3 { margin-top: 0; border-bottom: 1px solid #eee; padding-bottom: 0.5rem; }
		.terminal { background: #1e1e1e; color: #d4d4d4; padding: 1rem; border-radius: 4px; font-family: monospace; font-size: 13px; line-height: 1.5; max-height: 300px; overflow-y: auto; }
		.terminal-line { margin: 0; }
		.cancel-line { color: #ff6b6b; font-weight: bold; }
		.warn { background: #ffeef0; color: #b31d28; padding: 1rem; border-radius: 4px; font-weight: bold; text-align: center; margin-bottom: 1rem; }
		.success { background: #e6ffed; color: #1a7f37; padding: 1rem; border-radius: 4px; text-align: center; margin-bottom: 1rem; }
		.res-item { display: inline-block; background: #e6f4ff; border: 1px solid #007acc; padding: 5px 10px; border-radius: 4px; margin: 4px; font-family: monospace; }
	</style>
</head>
<body>
	<h1>Задача 5: Параллелизм (2) + Context Cancel</h1>
	<div class="task">
		<p><strong>Условие:</strong> Обработка (число × 5) занимает 0.5 сек. Работают ровно 2 воркера. Если результат превышает 50, вся работа отменяется через context.</p>
	</div>

	<form method="POST">
		<div class="form-group">
			<input type="text" name="numbers" value="{{.InputNumbers}}" placeholder="Числа через запятую">
			<button type="submit">Запустить обработку</button>
		</div>
	</form>

	{{if .Error}}
		<div class="warn">❌ {{.Error}}</div>
	{{end}}

	{{if .HasResult}}
		<div class="grid">
			<div class="panel">
				<h3>🖥️ Лог событий (горутины)</h3>
				<div class="terminal">
					{{range .Logs}}
						<p class="terminal-line {{if strings_contains . "ОТМЕНА"}}cancel-line{{end}}">{{.}}</p>
					{{end}}
				</div>
			</div>
			<div class="panel">
				<h3>📦 Собранные результаты</h3>
				{{if .Cancelled}}
					<div class="warn" style="font-size: 14px; margin-bottom: 10px;">
						⛔ Работа была прервана из-за превышения лимита (> 50)
					</div>
				{{else}}
					<div class="success" style="font-size: 14px; margin-bottom: 10px;">
						✅ Все задачи выполнены успешно
					</div>
				{{end}}
				
				{{if .Results}}
					{{range .Results}}
						<span class="res-item">{{.}}</span>
					{{end}}
				{{else}}
					<p style="color: #666;">Результатов нет (отмена произошла слишком быстро)</p>
				{{end}}
			</div>
		</div>
	{{end}}
</body>
</html>
`
	// Добавляем кастомную функцию для шаблона (проверка на подстроку)
	funcMap := template.FuncMap{
		"strings_contains": func(s, substr string) bool {
			return strings.Contains(s, substr)
		},
	}

	t, _ := template.New("webpage").Funcs(funcMap).Parse(tmpl)
	t.Execute(w, data)
}

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}