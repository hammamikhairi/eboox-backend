package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func Assert(cond bool, msg string, format ...any) {
	if !cond {
		fmt.Print("Assersion failure : ")
		fmt.Printf(msg, format...)
		os.Exit(1)
	}
}

func NormalizeBookName(original string) string {
	return strings.ReplaceAll(original, " ", "_")
}

func getExtension(fileName string) string {
	dotIndex := strings.LastIndex(fileName, ".")

	if dotIndex < 1 {
		return ""
	}
	beforeDot := fileName[dotIndex-1]

	if beforeDot == '/' || beforeDot == '\\' {
		return ""
	}
	return fileName[dotIndex+1:]
}

func GetFiles(path string) []string {

	var (
		files []string
		ext   string = "epub"
	)

	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if getExtension(p) == ext {
			bPath, _ := filepath.Abs(p)
			files = append(files, bPath[len(path):])
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Error walking the directory <%s> : %v\n", path, err)
	}

	return files
}

func FormatDate(date string) string {
	if strings.Contains(date, "T") {
		parts := strings.Split(date, "T")
		if len(parts) >= 2 {
			return parts[0]
		}
	}
	return ""
}

// XXX - DEV
func Pp(v any) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Print(string(b))
}

func ConfExists(confPath string) (string, bool) {
	confDir, _ := os.UserConfigDir()
	_, err := os.Stat(confDir + confPath)
	return confDir + "/eboox", !os.IsNotExist(err)
}

func ActivitiesExists(actPath string) bool {
	_, err := os.Stat(actPath)
	return !os.IsNotExist(err)
}

func MakeFullDir(path string) {
	err := os.Mkdir(path, os.ModePerm)
	Assert(err == nil, "Error creating Directory : [%s].\n", path)
}

func MakeFile(path string) {
	_, err := os.Create(path)
	Assert(err == nil, "Error creating File : [%s].\n", path)
}

func MakeConfig(path string) {
	MakeFullDir(path)
	MakeFile(path + "/conf.json")
}

func MakeActivity(path string) {
	MakeFile(path)
}

func MapToSlice[T, P any, Q comparable](metadata *map[Q]T, activities *map[Q]P) []map[string]interface{} {
	var out []map[string]interface{}
	for uuid, meta := range *metadata {
		out = append(out, map[string]interface{}{"metadata": meta, "activities": (*activities)[uuid]})
	}

	// sort by time
	// feel free to send me death threats : khairihammami@outlook.com
	sort.Slice(out, func(i, j int) bool {
		layout := "02-01-2006"
		timeI, _ := time.Parse(layout,
			strings.Split(fmt.Sprint(out[i]["activities"].(P)), " ")[1])
		timeJ, _ := time.Parse(layout,
			strings.Split(fmt.Sprint(out[j]["activities"].(P)), " ")[1])
		return timeI.After(timeJ)
	})

	return out
}

func GetToday() string {
	today := time.Now()
	return fmt.Sprintf("%02d-%02d-%d", today.Day(), today.Month(), today.Year())
}
