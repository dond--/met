package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/dond--/met/exif"
	"github.com/jessevdk/go-flags"
)

var opts struct {
	IsVerbose  bool `short:"v" long:"verbose" description:"Print logging"`
	RecuSearch bool `short:"r" long:"recsearch" description:"Recursive heuristic search for EXIF source"`
}

func main() {
	args, err := flags.ParseArgs(&opts, os.Args)
	if !flags.WroteHelp(err) {
		ch(err)
	}

	var files []string
	if len(args) < 2 {
		cwd, err := filepath.Abs(".")
		ch(err)
		if opts.IsVerbose {
			fmt.Println("Processing actual dir:", cwd)
		}
		files = expandDir(cwd)
	} else {
		sArgs := args[1:]
		for _, a := range sArgs {
			if isDir(a) {
				abs, err := filepath.Abs(filepath.Dir(a))
				ch(err)
				if abs == "/" { // in root dir proceed with provided argument only
					files = append(files, expandDir(a)...)
				} else { // otherwise process parent folder
					files = append(files, expandDir(abs)...)
				}
			} else {
				abs, err := filepath.Abs(a) // we need fullpath to FILE here!
				ch(err)
				if filterImages(a) {
					files = append(files, abs)
				} else if opts.IsVerbose {
					fmt.Println("❗️ Unsupported file type:", a)
				}
			}
		}
	}
	files = uniq(files)
	//   fmt.Println("list of files:", files)
	for _, f := range files {
		fmt.Println("processing:", f)
		extime, err := exif.ReadExifTime(f)
		ch(err)
		if extime != "" {
			touch(f, extime)
			fmt.Printf("Times → EXIF(%s): %s\n", extime, f)
		} else {
			if opts.RecuSearch {
				suitable := 0
				found := findSimilar(f)
				if len(found) > 0 {
					for _, ff := range found {
						extime, err = exif.ReadExifTime(ff)
						ch(err)
						if extime != "" { // use first non-empty similar file EXIF time
							touch(f, extime)
							fmt.Printf("Times → EXIF(%s): %s (used from %s)\n", extime, f, ff)
							suitable++
							break
						}
					}
				}
				if suitable == 0 {
					fmt.Println("❗️ No suitable source of EXIF time found")
				}
			} else {
				fmt.Println("❗️ EXIF time missing:", f)
			}
		}
	}
}

func ch(e error) {
	if e != nil {
		panic(e)
	}
}

func drillDown(d string, f string) []string {
	pDir, _ := os.Getwd()
	dir, err := os.Open(d)
	//   fmt.Println("drilling down dir",d)
	ch(err)
	content, err := dir.Readdir(0)
	ch(err)
	var foundSimilar []string
	for _, item := range content {
		if item.IsDir() {
			subDir, err := filepath.Abs(item.Name())
			ch(err)
			err = os.Chdir(subDir)
			ch(err)
			foundSimilar = append(foundSimilar, drillDown(subDir, f)...)
		} else {
			if strings.HasPrefix(item.Name(), f) {
				simF, err := filepath.Abs(item.Name())
				ch(err)
				foundSimilar = append(foundSimilar, simF)
				//         fmt.Println("adding item",simF)
			}
		}
	}
	dir.Close()
	err = os.Chdir(pDir)
	ch(err)
	return foundSimilar
}

func expandDir(a string) []string {
	var fl []string
	d, err := os.Open(a)
	ch(err)
	files, err := d.Readdir(0)
	ch(err)
	d.Close()

	for _, f := range files {
		if !f.IsDir() {
			fullpath := a + "/" + f.Name()
			if filterImages(fullpath) {
				fl = append(fl, fullpath)
			} else if opts.IsVerbose {
				fmt.Println("❗️ Unsupported file type:", fullpath)
			}
		}
	}
	return fl
}

func filterImages(name string) bool {
	re, _ := regexp.Compile("^.*\\.(jpg|JPG|jpeg|JPEG|png|PNG|heic|HEIC)")
	return re.MatchString(name)
}

func findSimilar(f string) []string {
	dir := filepath.Dir(f)
	fn := filepath.Base(f)
	re := regexp.MustCompile("[-–.]")
	lookFor := re.Split(fn, -1)[0]
	found := drillDown(dir, lookFor)
	// filter out original file
	i := 0
	for _, ff := range found {
		if ff != f {
			found[i] = ff
			i++
		}
	}
	found = found[:i]
	return found
}

func isDir(a string) bool {
	f, err := os.Open(a)
	ch(err)
	fi, err := f.Stat()
	ch(err)
	id := fi.IsDir()
	f.Close()
	return id
}

func touch(f string, extime string) {
	t, err := exif.GetTime(extime)
	ch(err)
	err = os.Chtimes(f, t, t)
	ch(err)
}

func uniq(files []string) []string {
	dups := make(map[string]int)
	for _, f := range files {
		_, prev := dups[f]
		if prev {
			dups[f] += 1
		} else {
			dups[f] = 1
		}
	}

	var uniqFiles []string
	for k, v := range dups {
		uniqFiles = append(uniqFiles, k)
		if opts.IsVerbose && v > 1 {
			fmt.Println("👉 Duplicate filename:", k)
		}
	}
	sort.Strings(uniqFiles)
	return uniqFiles
}
