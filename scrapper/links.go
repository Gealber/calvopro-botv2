package scrapper

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly"
	log "gopkg.in/inconshreveable/log15.v2"
)

const BASE_URL = "https://xnxx.com"

//Video ...
type Video struct {
	Title       string
	URL         string
	PageURL     string
	ImageURL    string
	Description string
	Size        int64
	Format      string
	Duration    string
}

//InfoVideos basic info of the videos for that query
func InfoVideos(query string) []*Video {
	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(len(USER_AGENTS))
	// Instantiate default collector
	c := colly.NewCollector(
		// Cache responses to prevent multiple download of pages
		// even if the collector is restarted
		colly.CacheDir(XNXX_CACHE),
		colly.UserAgent(USER_AGENTS[n]),
	)

	videos := make([]*Video, 0)
	videosChan := make(chan *Video)
	done := make(chan struct{})

	// On every a element which has href attribute call callback
	c.OnHTML("div.thumb-block", func(e *colly.HTMLElement) {
		video := &Video{}
		e.ForEach("div.thumb-inside > div.thumb", func(_ int, el *colly.HTMLElement) {
			video.PageURL = BASE_URL + el.ChildAttr("a", "href")
			video.ImageURL = el.ChildAttr("a > img:nth-child(1)", "data-src")
		})
		e.ForEach("div.thumb-under", func(_ int, el *colly.HTMLElement) {
			title := el.ChildAttr("p > a", "title")
			video.Title = cleanTitle(title)
			//setting duration
			md := strings.Split(el.ChildText("p.metadata"), " ")
			if len(md) > 2 {
				data := strings.Split(md[1], "\n")
				if len(data) > 1 {
					video.Duration = data[1]
				} else {
					video.Duration = "Uknown"
				}
			} else {
				video.Duration = "Uknown"
			}
		})
		//filter videos and only add those with size
		//less than 52428800 bytes (50 mgb)
		go func() {
			GetVideoURL(video)
			videosChan <- video
		}()
	})

	go func() {
        //timeout
        for {
            select {
            case <-time.After(7*time.Second):
                done <- struct{}{}
                return
            case v := <- videosChan:
			    if allowedSize(v) {
			    	videos = append(videos, v)
			    }
			    if len(videos) == 10 {
                    done <- struct{}{}
			    	return
			    }
            }
        }
	}()

	query = strings.Join(strings.Split(query, " "), "+")
	url := fmt.Sprintf("%s/search/%s", BASE_URL, query)
	c.Visit(url)

	<-done
	return videos
}

//verifySize ...
func allowedSize(video *Video) bool {
	url := video.URL
	//need to parse url and validate it
	client := http.Client{
		Timeout: 2 * time.Second,
	}
	response, err := client.Head(url)
	if err != nil {
		//no time for wait
		//ignoring error
		return false
	}
	defer response.Body.Close()

	contentLength := response.ContentLength
	if contentLength < 52428800 {
		video.Size = contentLength
		return true
	}
	return false
}

//randVideosSegment ...
func randVideosSegment(videos []*Video) []*Video {
	if len(videos) <= 10 {
		return videos
	}
	rand.Seed(time.Now().UnixNano())
	max := len(videos) - 10
	n := rand.Intn(max)
	return videos[n : n+10]
}

//GetVideoURL retrieve video url and add it to the struct
func GetVideoURL(video *Video) {
	text := retrieveContent(video.PageURL)
	patternURL := regexp.MustCompile(`setVideoUrlHigh\(\'(.+?)\'\)`)
	matches := patternURL.FindStringSubmatch(text)
	if len(matches) > 1 {
		video.URL = matches[1]
		return
	}
	optPatterURL := regexp.MustCompile(`setVideoUrlLow\(\'(.+?)\'\)`)
	optMatches := optPatterURL.FindStringSubmatch(text)
	if len(optMatches) > 1 {
		video.URL = optMatches[1]
	}
}

//retrieveContent of a webpage
func retrieveContent(url string) string {
	//retrieving first page of results
	response, err := http.Get(url)
	if err != nil {
		log.Crit(fmt.Sprintf("Unable to retrieve url: %s"), "err", err)
	}
	defer response.Body.Close()

	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Crit("Unable to read body response", "err", err)
	}
	text := string(b)
	return text
}

//PrepateQuery ...
func PrepareQuery(rawQuery string) string {
	query := strings.Map(func(r rune) rune {
		if r == '.' {
			return ' '
		}
		if r == '(' || r == ')' {
			return -1
		}
		return r
	}, rawQuery)
	query = strings.Join(strings.Split(query, " "), "+")
	return query
}

func cleanTitle(title string) string {
	if len(title) > 70 {
		title = title[:70]
	}
	title = strings.ReplaceAll(title, "http", "NOP")
	title = strings.Map(func(r rune) rune {
		if r == '.' {
			return -1
		}
		return r
	}, title)

	return title
}

//Serialize the videos to be passed to redis
func Serialize(videos []*Video) []byte {
	var result bytes.Buffer
	if len(videos) == 0 {
		return []byte{}
	}
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(videos)
	if err != nil {
		log.Crit("Error serializing videos", "err", err)
	}

	return result.Bytes()
}

//Deserialize the bytes into a slice of videos
func Deserialize(data []byte) []*Video {
	var videos []*Video
	if len(data) == 0 {
		return videos
	}
	decoder := gob.NewDecoder(bytes.NewReader(data))

	err := decoder.Decode(&videos)
	if err != nil {
		log.Crit("Error deserializing videos", "err", err)
	}

	return videos
}
