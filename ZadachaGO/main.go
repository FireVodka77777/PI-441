package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// detectOS определяет операционную систему
func detectOS() string {
	switch runtime.GOOS {
	case "windows":
		return "Windows"
	case "linux":
		return "Linux"
	case "darwin":
		return "macOS"
	case "freebsd":
		return "FreeBSD"
	default:
		return runtime.GOOS
	}
}

// searchFile ищет файлы по имени начиная с root
func searchFile(root, fileName string, exact bool) ([]string, error) {
	var results []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть директорию %q: %w", root, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%q не является директорией", root)
	}

	var walk func(dir string)
	walk = func(dir string) {
		defer wg.Done()

		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}

		for _, entry := range entries {
			fullPath := filepath.Join(dir, entry.Name())

			if entry.IsDir() {
				wg.Add(1)
				go walk(fullPath)
				continue
			}

			name := entry.Name()
			var match bool
			if exact {
				match = strings.EqualFold(name, fileName)
			} else {
				match = strings.Contains(strings.ToLower(name), strings.ToLower(fileName))
			}

			if match {
				mu.Lock()
				results = append(results, fullPath)
				mu.Unlock()
			}
		}
	}

	wg.Add(1)
	walk(root)
	wg.Wait()

	return results, nil
}

func main() {
	pathFlag := flag.String("path", "", "Директория для поиска (по умолчанию корень ОС)")
	nameFlag := flag.String("name", "", "Имя файла для поиска (обязательно)")
	exactFlag := flag.Bool("exact", false, "Точное совпадение имени")
	osFlag := flag.Bool("os", false, "Только показать определённую ОС")
	timeoutFlag := flag.Duration("timeout", 60*time.Second, "Таймаут поиска")

	flag.Parse()

	// ЭТАП 2: определение ОС
	currentOS := detectOS()
	fmt.Printf("Операционная система: %s (%s/%s)\n", currentOS, runtime.GOOS, runtime.GOARCH)

	if *osFlag {
		return
	}

	if *nameFlag == "" {
		fmt.Println("\nИспользование:")
		flag.PrintDefaults()
		fmt.Println("\nПример:")
		fmt.Println(`  finder -name "config.txt" -path "C:\Users"`)
		os.Exit(1)
	}

	root := *pathFlag
	if root == "" {
		switch runtime.GOOS {
		case "windows":
			root = "C:\\"
		default:
			root = "/"
		}
	}

	mode := "частичное совпадение"
	if *exactFlag {
		mode = "точное совпадение"
	}

	fmt.Printf("Поиск файла: %q\n", *nameFlag)
	fmt.Printf("Директория: %s\n", root)
	fmt.Printf("Режим: %s\n", mode)
	fmt.Println(strings.Repeat("-", 50))

	// ЭТАП 1: поиск файла с таймаутом
	done := make(chan struct{})
	var results []string
	var searchErr error

	go func() {
		results, searchErr = searchFile(root, *nameFlag, *exactFlag)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(*timeoutFlag):
		fmt.Println("Таймаут поиска истёк")
		os.Exit(2)
	}

	if searchErr != nil {
		fmt.Printf("Ошибка: %v\n", searchErr)
		os.Exit(1)
	}

	if len(results) == 0 {
		fmt.Println("Файлы не найдены.")
		return
	}

	fmt.Printf("Найдено файлов: %d\n\n", len(results))
	for i, r := range results {
		fmt.Printf("%3d. %s\n", i+1, r)
	}
}