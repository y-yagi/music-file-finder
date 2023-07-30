package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const cmd = "music-file-finder"

var (
	flags    *flag.FlagSet
	location string
	debug    bool
)

func main() {
	setFlags()
	os.Exit(run(os.Args, os.Stdout, os.Stderr))
}

func setFlags() {
	flags = flag.NewFlagSet(cmd, flag.ExitOnError)
	flags.StringVar(&location, "l", ".", "Search location")
	flags.BoolVar(&debug, "d", false, "Enable debug mode")
	flags.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stdout, "Usage: %s [OPTIONS]\n\n", cmd)
	fmt.Fprintln(os.Stdout, "OPTIONS:")
	flags.PrintDefaults()
}

func run(args []string, outStream, errStream io.Writer) int {
	flags.Parse(args[1:])
	if _, err := os.Stat(location); err != nil {
		fmt.Fprintf(outStream, "Location is invalid value. %v\n", err)
		return 1
	}

	search(location, outStream, errStream)
	return 0
}

func search(location string, outStream, errStream io.Writer) {
	var wg sync.WaitGroup

	entries, err := os.ReadDir(location)
	if err != nil {
		fmt.Fprintf(outStream, "%v\n", err)
		return
	}

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			fmt.Fprintf(outStream, "%v\n", err)
		}

		fullPath := filepath.Join(location, info.Name())
		if info.IsDir() {
			wg.Add(1)
			go func() {
				search(fullPath, outStream, errStream)
				wg.Done()
			}()
		} else if isMusicFile(fullPath, errStream) {
			fmt.Fprintf(outStream, "%s\n", fullPath)
		}
	}
	wg.Wait()
}

func isMusicFile(path string, errSteam io.Writer) bool {
	file, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(errSteam, "file open failed %s, %v\n", path, err)
		return false
	}
	defer file.Close()

	buf := make([]byte, 512)
	if _, err := file.Read(buf); err != nil {
		fmt.Fprintf(errSteam, "file read failed %s, %v\n", path, err)
		return false
	}

	contentType := http.DetectContentType(buf)
	if debug && !strings.HasPrefix(contentType, "audio") {
		fmt.Fprintf(errSteam, "Not audio file: '%s' is '%s'\n", path, contentType)
	}
	return strings.HasPrefix(contentType, "audio")
}
