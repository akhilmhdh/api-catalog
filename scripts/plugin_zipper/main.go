package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var version = "0.1-alpha"

func main() {
	if err := os.RemoveAll("./plugins/builtin.zip"); err != nil {
		panic(err)
	}

	fmt.Printf("Setting version number: %s\n", version)
	data := make(map[string]any)
	versionFile, err := ioutil.ReadFile("./plugins/builtin/version.json")
	if err != nil {
		panic(err)
	}
	json.Unmarshal(versionFile, &data)
	data["version"] = version
	jsonString, _ := json.Marshal(data)
	ioutil.WriteFile("./plugins/builtin/version.json", jsonString, os.ModePerm)
	fmt.Println("Updated version number")

	file, err := os.Create("./plugins/builtin.zip")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	w := zip.NewWriter(file)
	defer w.Close()

	walker := func(path string, info os.FileInfo, err error) error {
		fmt.Printf("Crawling: %#v\n", path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		f, err := w.Create(strings.TrimPrefix(path, "plugins"))
		if err != nil {
			return err
		}

		_, err = io.Copy(f, file)
		if err != nil {
			return err
		}

		return nil
	}
	err = filepath.Walk("./plugins/builtin", walker)
	if err != nil {
		panic(err)
	}
}
