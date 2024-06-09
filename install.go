package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

const (
	replaceID           = "AABBCCDD"
	bootstrapFolderName = "bootstrap"
	reponame            = "https://github.com/anthdm/gothkit.git"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Println()
		fmt.Println("install requires your project name as the first argument")
		fmt.Println()
		fmt.Println("\tgo run gothkit/install.go [your_project_name]")
		fmt.Println()
		os.Exit(1)
	}

	projectName := args[0]

	// check if gothkit folder already exists, if so, delete
	fi, err := os.Stat("gothkit")
	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}

	if fi != nil && fi.IsDir() {
		fmt.Println("-- deleting gothkit folder cause its already present")
		if err := os.RemoveAll("gothkit"); err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("-- cloning", reponame)
	clone := exec.Command("git", "clone", reponame)
	if err := clone.Run(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("-- rename bootstrap to", projectName)
	if err := os.Rename(path.Join("gothkit", bootstrapFolderName), projectName); err != nil {
		log.Fatal(err)
	}

	err = filepath.Walk(path.Join(projectName), func(fullPath string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		b, err := os.ReadFile(fullPath)
		if err != nil {
			return err
		}

		contentStr := string(b)
		if strings.Contains(contentStr, replaceID) {
			replacedContent := strings.ReplaceAll(contentStr, replaceID, projectName)
			file, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = file.WriteString(replacedContent)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("-- project (%s) successfully installed!\n", projectName)
}
