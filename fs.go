package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"text/tabwriter"
	"sync"
)

func main() {
	rootPtr := flag.String("root", "", "Путь до нужной директории")
	sortPtr := flag.String("sort", "ASK", "Параметр сортировки (возрастание/убывание)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Использование: go run fs.go --root=<путь_до_нужной_директории> --sort=<параметр_сортировки>\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *rootPtr == "" || *sortPtr == "" {
		fmt.Println("Ошибка: пропущены нужные флаги.")
		flag.Usage()
		return
	}

	if !(*sortPtr == "ASK" || *sortPtr == "DESC") {
		fmt.Println("Ошибка: флаг сорт не может быть с таким значением.")
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

	pairs := sortFiles(propertiesFiles, sortMethod)

	w := tabwriter.NewWriter(os.Stdout, 0, 1, 3, ' ', tabwriter.TabIndent)

	fmt.Fprintln(w, "Name\tType\tSize")

	for _, pair := range pairs {
		fmt.Fprintf(w, "%s\t%s\t%s\n", pair.Key, pair.Value.fileType, formatSize(pair.Value.size))
	}

	w.Flush()
}

var nameSizes = [5]string{"b", "Kb", "Mb", "Gb", "Tb"}

// formatSize преобразует размер в байтах к удобному для чтения виду
func formatSize(size int64) string {
	s := float64(size)

	if s < 1024 {
		return fmt.Sprintf("%.2f B", s)
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

	err := filepath.Walk(f, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		size += info.Size()

		return nil
	})
	if err != nil {
		return 0, err
	}

	return size, nil
}

func processFile(dirName string, file fs.DirEntry, propertiesFiles map[string]fileProperty, wg *sync.WaitGroup, mu *sync.Mutex) {
	defer wg.Done()

	fPath := fmt.Sprintf("%s/%s", dirName, file.Name())

	fileInfo, err := os.Stat(fPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	typeFile := "file"
	size := fileInfo.Size()

	if file.IsDir() {
		typeFile = "dir"

		var err error
		size, err = determineSize(fPath)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	mu.Lock()
	propertiesFiles[file.Name()] = fileProperty{fileType: typeFile, size: size}
	mu.Unlock()
}

// getPropertiesFiles возвращает свойства файлов из заданной директории
func getPropertiesFiles(dirName string) (map[string]fileProperty, error) {
	files, err := os.ReadDir(dirName)
	if err != nil {
		return nil, err
	}

	propertiesFiles := map[string]fileProperty{}

	var (
		mu sync.Mutex
		wg sync.WaitGroup
	)

	for _, file := range files {
		wg.Add(1)
		go processFile(dirName, file, propertiesFiles, &wg, &mu)
	}

	wg.Wait()

	return propertiesFiles, nil
}

type fileProperty struct {
	fileType string
	size     int64
}

type Pair struct {
	Key   string
	Value fileProperty
}

// sortFiles cортирует мапу с размерами файлов по размеру
func sortFiles(propertiesFiles map[string]fileProperty, sortMethod string) []Pair {
	pairs := []Pair{}

	for key, value := range propertiesFiles {
		pairs = append(pairs, Pair{Key: key, Value: value})
	}

	if sortMethod == "ASK" {
		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].Value.size < pairs[j].Value.size
		})
	}

	if sortMethod == "DESC" {
		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].Value.size > pairs[j].Value.size
		})
	}

	return pairs
}
