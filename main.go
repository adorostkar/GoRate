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

const goUsage = `GoRate is a tool to fetch information of movies listed under a folder

USAGE: gr <folder name> <options>
    OPTIONS:
        - name
        - vote
        - rate
        - year
        - genre
        - runtime
`

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

// Config file to pass around
type Config struct {
	nameParserExpression []string
	replacerExpression   string
	extensionExpression  string
	informer             Informer
}

func extractNameAndYear(basePath string, config Config) (string, int, error) {
	repl, err := regexp.Compile(config.replacerExpression)
	if err != nil {
		log.Printf("String '%s' is not a valid regular expression", config.replacerExpression)
	}
	for _, rs := range config.nameParserExpression {
		extractor, err := regexp.Compile(rs)
		if err != nil {
			log.Printf("String '%s' is not a valid regular expression", rs)
		}
		matches := extractor.FindStringSubmatch(basePath)
		if len(matches) == 0 {
			log.Printf("Cannot extract info for file %s, with regex %s\n", basePath, rs)
			continue
		}

		name := repl.ReplaceAllString(matches[1], " ")
		year, err := strconv.Atoi(matches[2])
		if err != nil {
			log.Printf("Could not retrieve the year from its string %s\n", matches[2])
		}
		return name, year, err
	}
	return "", 0, fmt.Errorf("Could not extract info with any of the provided regular expressions for the movie %s", basePath)
}

func populateMovieList(path string, config Config) []Movie {
	var movies []Movie
	r := regexp.MustCompile(config.extensionExpression)

	err := filepath.Walk(path,
		func(thisPath string, info os.FileInfo, err error) error {
			basePath := filepath.Base(thisPath)
			fullPath, _ := filepath.Abs(thisPath)
			if err != nil {
				log.Printf("Couldn't process %s", basePath)
				return err
			}
			if r.MatchString(filepath.Ext(basePath)) {
				name, year, err := extractNameAndYear(basePath, config)
				if err != nil {
					log.Println(err)
				}
				movies = append(movies, Movie{name: name, path: fullPath, year: year})
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

	ch <- Movie{name, path, genre, imdbID, runtime, year, vote, rate}
}

func getMovieInformation(movies []Movie, config Config) []Movie {
	ch := make(chan Movie)
	for _, m := range movies {
		go config.informer(m.name, m.year, m.path, ch)
	}

	lenM := len(movies)
	for i := 0; i < lenM; i++ {
		m := <-ch
		movies[i] = m
	}
	close(ch)
	return movies
}

func main() {
	log.SetOutput(ioutil.Discard)
	if len(os.Args) < 2 {
		log.Printf("%s\n", "Not enough input arguments")
		fmt.Printf("%s\n", goUsage)
		os.Exit(1)
	}

	moviePath := os.Args[1]
	sortArray := os.Args[2:]

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

	config := Config{
		nameParserExpression: []string{`([\p{L}\d'\._\-!\&, ]+)[_\- \.\(]*(\d{4})[_\- \.\)]`},
		replacerExpression:   `[\.\-_ ]`,
		extensionExpression:  `(?i)(mp4|avi|mkv)`,
		informer:             omdbInformer,
	}

	movies := populateMovieList(path, config)
	movies = getMovieInformation(movies, config)
	for _, m := range movies {
		fmt.Println(m)
	}
}
