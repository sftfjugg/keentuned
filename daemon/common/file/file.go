/*
Copyright © 2021 KeenTune

Package file for daemon, this package contains the file. The package implements the processing of files, including the transformation of file types, path acquisition, and directory existence judgment.
*/
package file

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// IsPathExist ...
func IsPathExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		if os.IsNotExist(err) {
			return false
		}
		fmt.Println(err)
		return false
	}
	return true
}

// DecoratePath cut the end of the path separator "/"
func DecoratePath(path string) string {
	if len(path) == 0 {
		return path
	}

	if string(path[len(path)-1]) == "/" {
		return strings.TrimSuffix(path, "/")
	}

	return path
}

// ReadFile2Map ...
func ReadFile2Map(path string) (map[string]interface{}, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read [%v] file:%v", path, err)
	}

	var retMap map[string]interface{}
	err = json.Unmarshal(bytes, &retMap)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal err: %v", err)
	}

	return retMap, nil
}

// Dump2File ...
func Dump2File(path, fileName string, info interface{}) error {
	if !IsPathExist(path) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return fmt.Errorf("make dir err:[%v]", err)
		}
	}

	fullPath := path + "/" + fileName

	resultBytes, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("marshal info %v to bytes err:[%v]", info, err)
	}

	err = ioutil.WriteFile(fullPath, resultBytes, os.ModePerm)
	if err != nil {
		return fmt.Errorf("write to file [%v] err:[%v]", fileName, err)
	}

	return nil
}

//  WalkFilePath walk file path
// return: 0) full paths; 1) file names; 2) err msg
func WalkFilePath(folder, match string, exact ...bool) ([]string, []string, error) {
	var files, fullPaths []string

	filepath.Walk(folder, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		pathSections := strings.Split(path, "/")
		if len(pathSections) == 0 {
			return fmt.Errorf("get path [%v] Sections is null", path)
		}

		var fileName string
		if !fi.IsDir() {
			fileName = pathSections[len(pathSections)-1]
			if len(exact) > 0 && exact[0] && fileName == match {
				files = append(files, fileName)
				fullPaths = append(fullPaths, path)
			}

			if len(exact) == 0 && strings.Contains(path, match) {
				files = append(files, fileName)
				fullPaths = append(fullPaths, path)
			}
		}

		return nil
	})

	return fullPaths, files, nil
}

// Save2CSV save data to csv file
func Save2CSV(path, fileName string, data map[string][]float32) error {
	if !IsPathExist(path) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			fmt.Printf("save to %v cvs file, make dir err:[%v]\n", fileName, err)
			return fmt.Errorf("make dir err:[%v]", err)
		}
	}

	if len(data) == 0 {
		fmt.Printf("save to %v cvs file, data length is 0\n", fileName)
		return nil
	}

	fullName := fmt.Sprintf("%s/%s", path, fileName)

	newFile, err := os.Create(fullName)
	if err != nil {
		fmt.Printf("creat %v cvs file failed\n", fullName)
		return fmt.Errorf("create [%v] failed", fullName)
	}

	defer newFile.Close()

	w := csv.NewWriter(newFile)
	var headers []string
	for name, _ := range data {
		headers = append(headers, name)
	}

	dataPair := make([][]string, len(data[headers[0]]))

	sort.Strings(headers)
	contents := [][]string{
		headers,
	}

	for index := 0; index < len(headers); index++ {
		for row, value := range data[headers[index]] {
			dataPair[row] = append(dataPair[row], fmt.Sprintf("%v", value))
		}
	}

	contents = append(contents, dataPair...)
	w.WriteAll(contents)
	w.Flush()
	return nil
}

// GetWalkPath get match file path
func GetWalkPath(folder, match string) (string, error) {
	var result string
	filepath.Walk(folder, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !fi.IsDir() && strings.Contains(path, match) {
			result = path
		}

		return nil
	})

	return result, nil
}

// GetPlainName get plain file name without any path
func GetPlainName(fileName string) string {
	if !strings.Contains(fileName, "/") || fileName == "" {
		return fileName
	}

	parts := strings.Split(fileName, "/")
	if len(parts) < 1 {
		return ""
	}

	return parts[len(parts)-1]
}

// Copy copy file from src to dst
func Copy(src, dst string) error {
	input, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(dst, input, 0666)
	return err
}


