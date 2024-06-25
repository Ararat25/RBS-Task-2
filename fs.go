package main

import (
	"flag"
	"fmt"
	"os"
	"io/fs"
	"path/filepath"
	"sort"
)

// formatSize преобразует размер в байтах к удобному для чтения виду
func formatSize(size int64) string {
	s := float64(size)

    nameSizes := []string{"b", "Kb", "Mb", "Gb", "Tb"}

    if s < 1024 {
        return fmt.Sprintf("%d B", s)
    }

    i := 0

    for s >= 1024 && i < len(nameSizes)-1 {
        s /= float64(1024)
        i++
    }

    return fmt.Sprintf("%.2f %s", s, nameSizes[i])
}

// determineSize определяет полный размер директории вместе с файлами
func determineSize(f string) (int64, error) {
	var size int64
   
	err := filepath.Walk(f, func(_ string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}

		return nil
	})
	if err != nil {
		return 0, err
	}
	
	return size, nil
}

// getPropertiesFiles возвращает свойства файлов из заданной директории
func getPropertiesFiles(dirName string) (map[string]int64, error) {
	files, err := os.ReadDir(dirName)
    if err != nil {
		return nil, err
    }

	propertiesFiles := map[string]int64{}

    for _, file := range files {
		fPath := fmt.Sprintf("%s/%s", dirName, file.Name())

		fileInfo, err := os.Stat(fPath)
        if err != nil {
            fmt.Println(err)
            continue
        }

		typeFile := "file"
		size := fileInfo.Size()

		if file.IsDir() {
			typeFile = "dir"

			s, err := determineSize(fPath)
			if err != nil {
				fmt.Println(err)
				continue
			}

			size += s
		}

		propertiesFiles[fmt.Sprintf("%s\t%s\t", file.Name(), typeFile)] = size
    }

	return propertiesFiles, nil
}

type Pair struct {
	Key   string
	Value int64
}

func (p Pair) Less(other Pair) bool {
	return p.Value < other.Value
}

func sortFiles(propertiesFiles map[string]int64, sortMethod string) {
	pairs := []Pair{}

	for key, value := range propertiesFiles {
		pairs = append(pairs, Pair{Key: key, Value: value})
	}
	
	if sortMethod == "ASK" {
		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].Value < pairs[j].Value
		})
	}

	if sortMethod == "DESC" {
		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].Value > pairs[j].Value
		})
	}
	
	
	// Выводим отсортированный список ключей и значений
	for _, pair := range pairs {
		fmt.Println(pair.Key, formatSize(pair.Value))
	}
}


func main() {
	rootPtr := flag.String("root", "", "Путь до нужной директории")
	sortPtr := flag.String("sort", "ASK", "Параметр сортировки (возрастание/убывание)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Использование: go run fs.go --root=<путь_до_нужной_директории> --sort=<параметр_сортировки>\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *rootPtr == "" || *sortPtr == "" {
		fmt.Println("Error: missing required flags.")
		flag.Usage()
		return
	}


	if !(*sortPtr == "ASK" || *sortPtr == "DESC") {
		fmt.Println("Error: sort can't have that value.")
		flag.Usage()
		return
	}

	dirName := *rootPtr
	sortMethod := *sortPtr

	propertiesFiles, err := getPropertiesFiles(dirName)
	if err != nil {
		fmt.Println(err)
		return
	}

	sortFiles(propertiesFiles, sortMethod)
}
