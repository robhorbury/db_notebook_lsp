package analysis

import "strings"

func GetTempFileName(fileName string) string {
	newFileName := strings.ReplaceAll(fileName, ":", "_")
	newFileName = strings.ReplaceAll(newFileName, "/", "_")
	newFileName = strings.ReplaceAll(newFileName, "\\", "_")

	return newFileName
}

func GetTempPath() string {
	return "./.customLsp/.tempFiles/"
}
