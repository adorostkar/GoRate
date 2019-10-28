package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

const goUsage = `GoRate is a tool to fetch information of movies listed under a folder

USAGE: gr <folder paths>
`

// InformerFunc is the function type that ''in some way'' retrieves the movie details
type InformerFunc func(title string, year int, path string, config Config, ch chan Movie)

// Movie information is saved in this struct
type Movie struct {
	Title       string
	Path        string
	Genre       []string
	ImdbID      string
	Runtime     string
	Year        int
	Vote        int
	Rate        float64
	Plot        string
	PosterImage string
	Actors      string
}

// Config file to pass around
type Config struct {
	NameParserExpression   []string
	TitleCleanupExpression string
	ExtensionExpression    string
	APIKey                 string
	MovieAPI               string
	humanReadable          bool
}

func extractNameAndYear(basePath string, config Config) (string, int, error) {
	repl, err := regexp.Compile(config.TitleCleanupExpression)
	if err != nil {
		log.WithFields(
			log.Fields{
				"topic": "regex",
				"regex": config.TitleCleanupExpression,
			}).Error("Not a valid regular expression")
	}
	for _, rs := range config.NameParserExpression {
		extractor, err := regexp.Compile(rs)
		if err != nil {
			log.WithFields(
				log.Fields{
					"topic": "regex",
					"regex": rs,
				}).Error("Not a valid regular expression")
		}
		matches := extractor.FindStringSubmatch(basePath)
		if len(matches) == 0 {
			log.WithFields(
				log.Fields{
					"topic":  "regex",
					"regex":  rs,
					"string": basePath,
				}).Error("Cannot extract info")
			continue
		}

		title := repl.ReplaceAllString(matches[1], " ")
		year, err := strconv.Atoi(matches[2])
		if err != nil {
			log.WithFields(
				log.Fields{
					"topic":  "regex",
					"string": matches[2],
				}).Error("Could not retrieve the year")
		}
		return title, year, err
	}
	return "", 0, fmt.Errorf("Could not extract info with any of the provided regular expressions for the movie %s", basePath)
}

func populateMovieList(path string, config Config) []Movie {
	var movies []Movie
	r := regexp.MustCompile("(?i)" + config.ExtensionExpression)

	err := filepath.Walk(path,
		func(thisPath string, info os.FileInfo, err error) error {
			basePath := filepath.Base(thisPath)
			fullPath, _ := filepath.Abs(thisPath)
			if err != nil {
				log.WithFields(
					log.Fields{
						"topic": "file",
						"path":  basePath,
					}).Errorf("Couldn't process")
				return err
			}
			if r.MatchString(filepath.Ext(basePath)) {
				title, year, err := extractNameAndYear(basePath, config)
				if err != nil {
					log.WithFields(
						log.Fields{
							"topic": "file",
							"path":  basePath,
						}).Error(err)
				}
				movies = append(movies, Movie{Title: title, Path: fullPath, Year: year})
				log.WithFields(
					log.Fields{
						"topic": "file",
						"path":  basePath,
					}).Trace(filepath.Ext(basePath))
			}
			return nil
		})
	if err != nil {
		log.Error(err)
	}
	return movies
}

func emptyInformer(title string, year int, path string, config Config, ch chan Movie) {
	ch <- Movie{Title: title, Year: year, Path: path}
}

func omdbInformer(title string, year int, path string, config Config, ch chan Movie) {
	// when finished give done
	apiHTTP := fmt.Sprintf("http://www.omdbapi.com/?apikey=%s&t=%s&y=%d", config.APIKey, title, year)
	apiHTTP = strings.ReplaceAll(apiHTTP, " ", "%20")
	log.WithFields(
		log.Fields{
			"topic": "Movie",
			"Url":   apiHTTP,
		}).Trace("")
	resp, err := http.Get(apiHTTP)
	if err != nil {
		log.WithFields(
			log.Fields{
				"topic": "Movie",
				"name":  title,
				"Url":   apiHTTP,
			}).Error("Could not fetch")
		return
	}
	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)
	var raw map[string]interface{}
	json.Unmarshal(data, &raw)
	if raw["Response"].(string) == "False" {
		log.WithFields(
			log.Fields{
				"topic": "Movie",
				"name":  title,
			}).Trace("Not found")
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
	plot, _ := raw["Plot"].(string)
	poster, _ := raw["Poster"].(string)
	actors, _ := raw["Actors"].(string)

	ch <- Movie{title, path, genre, imdbID, runtime, year, vote, rate, plot, poster, actors}
}

func getMovieInformation(movies []Movie, config Config, informer InformerFunc) {
	ch := make(chan Movie)
	for _, m := range movies {
		go informer(m.Title, m.Year, m.Path, config, ch)
	}

	lenM := len(movies)
	for i := 0; i < lenM; i++ {
		m := <-ch
		movies[i] = m
	}
	close(ch)
}

func apiSelectFromName(name string) InformerFunc {
	switch name {
	case "omdb":
		return omdbInformer
	case "rt":
		return emptyInformer
	}
	return emptyInformer
}

func loadConfig() Config {
	var config Config
	configFile, err := os.Open("assets/userConfig.json")
	if err != nil {
		log.WithFields(
			log.Fields{
				"topic": "config",
			}).Info(err)
		configFile, err = os.Open("assets/config.json")
		if err != nil {
			log.WithFields(
				log.Fields{
					"topic": "config",
				}).Fatal("No configuration file found.", err)
		}
	}
	defer configFile.Close()

	byteValue, _ := ioutil.ReadAll(configFile)

	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		log.Fatal(err)
	}
	return config
}

func main() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.WarnLevel)

	config := loadConfig()
	movieInformer := apiSelectFromName(config.MovieAPI)

	mainLayout := template.Must(template.ParseFiles("assets/mainLayout.html"))
	settingsLayout := template.Must(template.ParseFiles("assets/settingsLayout.html"))

	fs := http.FileServer(http.Dir("assets/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	// TODO: add functionality to be able to reset config
	http.HandleFunc("/settings", func(w http.ResponseWriter, r *http.Request) {
		postVals := struct {
			Success int
			Config
		}{0, config}
		if r.Method != http.MethodPost {
			postVals.Success = 2
			if err := settingsLayout.Execute(w, postVals); err != nil {
				log.WithFields(
					log.Fields{
						"topic": "settingsLayout",
					}).Fatal(err)
			}
			return
		}

		config = Config{
			ExtensionExpression:    r.FormValue("extension"),
			TitleCleanupExpression: r.FormValue("separator"),
			NameParserExpression:   strings.Split(r.FormValue("nameparser"), "\n"),
			APIKey:                 r.FormValue("apiKey"),
			MovieAPI:               r.FormValue("informer"),
		}

		data, _ := json.MarshalIndent(config, "", "  ")
		err := ioutil.WriteFile("assets/userConfig.json", data, 0644)
		if err != nil {
			log.WithFields(
				log.Fields{
					"topic": "config",
				}).Info("Can't save updated configurations", err)
			postVals.Success = 1
		}
		settingsLayout.Execute(w, postVals)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[1:]
		var msg string
		movies := populateMovieList(path, config)
		getMovieInformation(movies, config, movieInformer)
		sort.Slice(movies, func(i, j int) bool {
			return movies[i].Title < movies[j].Title
		})
		if len(movies) == 0 {
			msg = `Please specify the path to the movie folder. e.g. localhost:8080/E:/Film`
		}
		if err := mainLayout.Execute(w, struct {
			Msg    string
			Movies []Movie
		}{msg, movies}); err != nil {
			log.WithFields(
				log.Fields{
					"topic": "mainLayout",
				}).Fatal(err)
		}
	})
	log.WithFields(
		log.Fields{
			"topic": "main",
		}).Fatal(http.ListenAndServe(":8080", nil))
}
