package main

import "fmt"

type Resource interface {
	Name() string
	Close()
}

type File struct {
	filename string
}

func (f File) Name() string {
	return f.filename
}

func (f File) Close() {
	fmt.Println("close file:", f.filename)
}

type Connection struct {
	address string
}

func (c Connection) Name() string {
	return c.address
}

func (c Connection) Close() {
	fmt.Println("close connection:", c.address)
}

func useAll(resources []Resource) {
	defer func() {
		for _, r := range resources {
			func() {
				defer func() {
					if rec := recover(); rec != nil {
						fmt.Println("panic while closing", r.Name(), ":", rec)
					}
				}()
				r.Close()
			}()
		}
	}()

	for _, r := range resources {
		fmt.Println("using", r.Name())
	}
}

func main() {
	resources := []Resource{
		File{filename: "data.txt"},
		Connection{address: "localhost:5432"},
		File{filename: "report.pdf"},
		Connection{address: "api.example.com:443"},
	}

	useAll(resources)
}