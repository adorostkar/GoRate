package main

import (
	"encoding/json"
	"fmt"
	"html/template"
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

USAGE: gr <folder paths>
`

// Informer is the function type that ''in some way'' retrieves the movie details
type Informer func(title string, year int, path string, ch chan Movie)

// Movie information is saved in this struct
type Movie struct {
	Title   string
	Path    string
	Genre   []string
	ImdbID  string
	Runtime string
	Year    int
	Vote    int
	Rate    float64
	Plot    string
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

		title := repl.ReplaceAllString(matches[1], " ")
		year, err := strconv.Atoi(matches[2])
		if err != nil {
			log.Printf("Could not retrieve the year from its string %s\n", matches[2])
		}
		return title, year, err
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
				title, year, err := extractNameAndYear(basePath, config)
				if err != nil {
					log.Println(err)
				}
				movies = append(movies, Movie{Title: title, Path: fullPath, Year: year})
				log.Printf("%s, %s\n", basePath, filepath.Ext(basePath))
			}
			return nil
		})
	if err != nil {
		log.Println(err)
	}
	return movies
}

func omdbInformer(title string, year int, path string, ch chan Movie) {
	// when finished give done
	apiHTTP := fmt.Sprintf("http://www.omdbapi.com/?apikey=%s&t=%s&y=%d", "c3658413", title, year)
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
		log.Printf("Movie %s not found", title)
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

	ch <- Movie{title, path, genre, imdbID, runtime, year, vote, rate, plot}
}

func getMovieInformation(movies []Movie, config Config) {
	ch := make(chan Movie)
	for _, m := range movies {
		go config.informer(m.Title, m.Year, m.Path, ch)
	}

	lenM := len(movies)
	for i := 0; i < lenM; i++ {
		m := <-ch
		movies[i] = m
	}
	close(ch)
}

// var report = template.Must(template.New("report").Parse(`Movies:
// {{range .}}-------------------------------------------------
// Title  : {{.Title}}
// Genre  : {{.Genre}}
// Runtime: {{.Runtime}}
// Year   : {{.Year}}
// Vote   : {{.Vote}}
// Rate   : {{.Rate}}
// URL    : www.imdb.com/title/{{.ImdbID}}
// {{end}}
// -------------------------------------------
// Total count: {{. | len}}`))

var report = template.Must(template.New("report").Parse(`
<!DOCTYPE html>
<html lang="en-US">
<head>
    <meta charset="UTF-8">
    <meta name="MoviesList" content="width=device-width, initial-scale=1">
    <title>Movies List</title>
	<script src="https://kryogenix.org/code/browser/sorttable/sorttable.js"></script>
</head>
	<body>
		<h1>{{. | len}} Movies found</h1>
		<table class="sortable">
			<tr style='text-align: left'>
			<th>#</th>
			<th>Title</th>
			<th>Genre</th>
			<th>Year</th>
			<th>Runtime</th>
			<th>Vote</th>
			<th>Rate</th>
			</tr>
			{{range .}}
			<tr>
			<td></td>
			<td title="{{.Plot}}"><a href="http://www.imdb.com/title/{{.ImdbID}}">{{.Title}}</a></td>
			<td>{{.Genre}}</td>
			<td>{{.Year}}</td>
			<td>{{.Runtime}}</td>
			<td>{{.Vote}}</td>
			<td>{{.Rate}}</td>
			</tr>
			{{end}}
		</table>
	</body>
</html>
`))

func main() {
	log.SetOutput(ioutil.Discard)
	if len(os.Args) < 2 {
		log.Printf("%s\n", "Not enough input arguments")
		fmt.Printf("%s\n", goUsage)
		os.Exit(1)
	}

	moviePath := os.Args[1]

	path, err := filepath.Abs(moviePath)
	if err != nil {
		log.Fatal("Could not expand path")
		os.Exit(1)
	}
	log.Printf("Given Path is %s\n", path)

	config := Config{
		nameParserExpression: []string{`([\p{L}\d'\._\-!\&, ]+)[_\- \.\(]*(\d{4})[_\- \.\)]`},
		replacerExpression:   `[\.\-_ ]`,
		extensionExpression:  `(?i)(mp4|avi|mkv)`,
		informer:             omdbInformer,
	}

	movies := populateMovieList(path, config)
	getMovieInformation(movies, config)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := report.Execute(w, movies); err != nil {
			log.Fatal(err)
		}
	})
	http.ListenAndServe(":8080", nil)

	// if err := report.Execute(os.Stdout, movies); err != nil {
	// 	log.Fatal(err)
	// }
}
