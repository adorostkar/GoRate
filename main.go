package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Informer is the function type that ''in some way'' retrieves the movie details
type Informer func(name string, year int, path string, ch chan Movie)

// Movie information is saved in this struct
type Movie struct {
	name            string
	path            string
	genre           []string
	imdbID, runtime string
	year, vote      int
	rate            float64
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

func omdbInformer(name string, year int, path string, ch chan Movie) {
	// when finished give done
	apiHTTP := fmt.Sprintf("http://www.omdbapi.com/?apikey=%s&t=%s&y=%d", "c3658413", name, year)
	apiHTTP = strings.ReplaceAll(apiHTTP, " ", "%20")
	log.Println(apiHTTP)
	resp, err := http.Get(apiHTTP)
	if err != nil {
		log.Printf("Error getting movie details for %s", apiHTTP)
		return
	}

	data, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(data))
	var raw map[string]interface{}
	json.Unmarshal(data, &raw)
	if raw["Response"].(string) == "False" {
		log.Printf("Movie %s not found", name)
		return
	}
	var (
		rate  float64
		vote  int
		genre []string
	)
	if ss, ok := raw["imdbRating"].(string); ok {
		rate, _ = strconv.ParseFloat(ss, 2)
	}
	runtime, _ := raw["Runtime"].(string)
	if ss, ok := raw["imdbVotes"].(string); ok {
		ss = strings.ReplaceAll(ss, ",", "")
		vote, _ = strconv.Atoi(ss)
	}
	imdbID, _ := raw["imdbID"].(string)
	if ss, ok := raw["Genre"].(string); ok {
		genre = strings.Split(ss, ",")
	}

	movie := Movie{name, path, genre, imdbID, runtime, year, vote, rate}
	ch <- movie
}

func getMovieInformation(movies map[string]Movie, movieInformer Informer) map[string]Movie {
	ch := make(chan Movie)
	for _, m := range movies {
		go movieInformer(m.name, m.year, m.path, ch)
	}

	lenM := len(movies)
	filledMovies := make(map[string]Movie)
	for i := 0; i < lenM; i++ {
		m := <-ch
		filledMovies[m.name] = m
	}
	close(ch)
	return filledMovies
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
	movies = getMovieInformation(movies, currentInformer)
	for _, m := range movies {
		fmt.Println(m)
	}
}
