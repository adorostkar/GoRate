package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

// Movie information is saved in this struct
type Movie struct {
	name                              string
	path                              string
	gener                             []string
	imdbID, year, vote, rate, runtime int
}

func (m Movie) String() string {
	return fmt.Sprintf("%s, %s, %d, %d, %v", m.name, m.path, m.year, m.rate, m.gener)
}

func help(programName string) string {
	return fmt.Sprintf("USAGE: %s <folder name> <options>\n"+
		"OPTIONS:\n"+
		"    - name\n"+
		"    - vote\n"+
		"    - rate\n"+
		"    - year\n"+
		"    - genre\n"+
		"    - runtime", programName)
}

func extractNameAndYear(basePath string) (string, int, error) {
	extractor := regexp.MustCompile("([\\p{L}\\d'`\\._\\-!\\&, ]+)[_\\- \\.\\(]*(\\d{4})[_\\- \\.\\)]")
	repl := regexp.MustCompile("[\\.\\-_ ]")
	matches := extractor.FindStringSubmatch(basePath)
	if len(matches) == 0 {
		log.Printf("Cannot extract info for file %s\n", basePath)
		return "", 0, fmt.Errorf("Can't find name and year from path format")
	}

	name := repl.ReplaceAllString(matches[1], " ")
	year, err := strconv.Atoi(matches[2])
	if err != nil {
		log.Printf("Could not retrieve the year from its string %s\n", matches[2])
	}
	return name, year, err
}

func populateMovieList(path string) map[string]Movie {
	movies := make(map[string]Movie)
	r := regexp.MustCompile("(?i)(mp4|avi|mkv)")

	err := filepath.Walk(path,
		func(thisPath string, info os.FileInfo, err error) error {
			basePath := filepath.Base(thisPath)
			fullPath, _ := filepath.Abs(thisPath)
			if err != nil {
				log.Printf("Couldn't process %s", basePath)
				return err
			}
			if r.MatchString(filepath.Ext(basePath)) {
				name, year, err := extractNameAndYear(basePath)
				if err != nil {
					log.Println(err)
				}
				movies[name] = Movie{name: name, path: fullPath, year: year}
				log.Printf("%s, %s\n", basePath, filepath.Ext(basePath))
			}
			return nil
		})
	if err != nil {
		log.Println(err)
	}
	return movies
}

func omdbInformer(movies map[string]Movie) {

}

func main() {
	log.SetOutput(ioutil.Discard)
	if len(os.Args) < 2 {
		log.Printf("%s\n", "Not enough input arguments")
		fmt.Printf("%s\n", help(os.Args[0]))
		os.Exit(1)
	}

	moviePath := os.Args[1]
	sortArray := os.Args[2:]
	currentInformer := omdbInformer

	path, err := filepath.Abs(moviePath)
	if err != nil {
		log.Fatal("Could not expand path")
		os.Exit(1)
	}
	log.Printf("Given Path is %s\n", path)
	if len(sortArray) > 0 {
		log.Printf("Sorting %v requested", sortArray)
	} else {
		log.Println("No sorting requested")
	}
	movies := populateMovieList(path)
	currentInformer(movies)
	for _, m := range movies {
		fmt.Println(m)
	}
}
