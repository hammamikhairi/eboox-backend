package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
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

func MapToSlice[T any, Q comparable](origin *map[Q]T) []T {
	var out []T
	for _, val := range *origin {
		out = append(out, val)
	}
	return out
}
