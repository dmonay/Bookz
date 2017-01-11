// ./bookz -dir=/Users/Dima/Desktop/test/
package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sync"
)

// add a status indicator
// add tests
// add benchmarks
type flagsStr struct {
	dir     string
	verbose bool
}

var flags flagsStr

// List of formats supported by calibre here:
// https://manual.calibre-ebook.com/faq.html#id12
func main() {
	empty := ""
	dir := flag.String("dir", empty, "directory containing input files")
	ver := flag.Bool("v", false, "Set to true if you want verbose output")

	flag.Parse()
	if *dir == empty {
		log.Fatal("Must provide directory with input files")
	}
	flags = flagsStr{*dir, *ver}

	// check if dir exists
	dirExists, err := exists(*dir)
	if !dirExists || err != nil {
		log.Fatal("Provided input dir does not exist")
	}

	// create output dir "mobi"
	outputDir := path.Join(*dir, "mobi")
	err = os.Mkdir(outputDir, 0700)
	if err != nil && !os.IsExist(err) {
		log.Fatalf("Error creating output dir: %s ", err)
	}

	var wg sync.WaitGroup

	// for each file in directory, spin up a goroutine to convert it
	err = filepath.Walk(*dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && isAllowed(info.Name()) {
			wg.Add(1)
			fullPath := filepath.Join(*dir, info.Name())
			go convertBook(fullPath, &wg)
		}

		return nil
	})
	if err != nil {
		log.Fatalf("Error listing files in %s: %s", *dir, err)
	}

	wg.Wait()
}

func isAllowed(filename string) bool {
	var extension = filepath.Ext(filename)
	switch extension {
	case ".fb2":
		return true
	case ".txt":
		return true
	case ".pdf":
		return true
	case ".epub":
		return true
	default:
		return false
	}
}

func convertBook(filename string, wg *sync.WaitGroup) {
	defer wg.Done()

	baseName := filepath.Base(filename)
	output := filepath.Join(flags.dir, "mobi", trimExtension(baseName)+".mobi")
	convert := "/Applications/calibre.app/Contents/MacOS/ebook-convert"

	cmd := exec.Command(convert, filename, output)
	if flags.verbose {
		cmd.Stdout = os.Stdout
	}
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Printf("Error converting %s: %s", baseName, err)
	}
}

// Could use strings.LastIndex to find the last period (.) but this may be more fragile
// in that there will be edge cases (e.g. no extension) or if Go were to be run on a
// theoretical OS that uses an extension delimiter other than the period.
func trimExtension(filename string) string {
	var extension = filepath.Ext(filename)
	return filename[0 : len(filename)-len(extension)]
}

// exists returns whether the given file or directory exists or not
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}
