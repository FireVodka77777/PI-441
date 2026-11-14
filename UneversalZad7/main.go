package main

import "fmt"

type Record struct {
	Name string
	Age  int
}

type Validator interface {
	IsValid(r Record) bool
}

type NameValidator struct{}

func (v NameValidator) IsValid(r Record) bool {
	return r.Name != ""
}

type AgeValidator struct{}
func (v AgeValidator) IsValid(r Record) bool {
	return r.Age >= 18 && r.Age <= 99
}

func validateAll(records []Record, v Validator) []Record {
	result := make([]Record, 0)

	for _, r := range records {
		if v.IsValid(r) {
			result = append(result, r)
		}
	}

	return result
}


func main() {

	records := []Record{
		{Name: "Анна", Age: 25},   
		{Name: "", Age: 30},       
		{Name: "Иван", Age: 17},   
		{Name: "Пётр", Age: 150}, 
		{Name: "Мария", Age: 40},  
		{Name: "", Age: 200},      
	}

	fmt.Println("Всего записей:", len(records))
	fmt.Println("---------------------------------")

	byName := validateAll(records, NameValidator{})
	fmt.Println("Валидные по имени:")
	for _, r := range byName {
		fmt.Printf("  %s, %d\n", r.Name, r.Age)
	}

	byAge := validateAll(records, AgeValidator{})
	fmt.Println("Валидные по возрасту:")
	for _, r := range byAge {
		fmt.Printf("  %s, %d\n", r.Name, r.Age)
	}
}