package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

// PORT is the TCP port number the server will listen to
var PORT = ":2349"

var (
	nFiles = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "GDRIVE",
			Name:      "number_of_files",
			Help:      "This is the number of files",
		})

	nDirs = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "GDRIVE",
			Name:      "number_of_directories",
			Help:      "This is the number of directories",
		})
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func main() {
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, drive.DriveMetadataReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	http.Handle("/metrics", promhttp.Handler())
	prometheus.MustRegister(nFiles)
	prometheus.MustRegister(nDirs)

	srv, err := drive.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	go func() {
		for {
			var files float64 = 0
			var directories float64 = 0

			r, err := srv.Files.List().PageSize(10).Fields("nextPageToken, files(id,name,md5Checksum,mimeType,size,createdTime,parents)").Do()
			if err != nil {
				log.Fatalf("Unable to retrieve files: %v", err)
			}

			if len(r.Files) == 0 {
				log.Println("No files found.")
				files = 0
				directories = 0
			} else {
				for _, i := range r.Files {
					if i.MimeType == "application/vnd.google-apps.folder" {
						// log.Println("DIRECTORY")
						directories = directories + 1
					} else {
						files = files + 1
					}
				}
			}

			// Set Prometheus metrics
			nFiles.Set(files)
			nDirs.Set(directories)
			fmt.Println("Files:", files, "Folders:", directories)

			time.Sleep(600 * time.Second)
		}
	}()

	log.Println("Listening to port", PORT)
	log.Println(http.ListenAndServe(PORT, nil))
}
