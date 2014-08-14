package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/ChimeraCoder/anaconda"
	_ "github.com/mattn/go-sqlite3"
)

func checkErr(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
		os.Exit(-1)
	}
}

func getStatusURL(tweetName string, tweetId int64) string {
	tweetId_s := strconv.FormatInt(tweetId, 10)
	url := "https://twitter.com/" + tweetName + "/status/" + tweetId_s
	return url
}

// TODO: create function to write data to logfile.
func writeLog() {
}

func yO(apikey string) bool {
	b := strings.NewReader("api_token={{" + apikey + "}}")
	_, err := http.Post("http://api.justyo.co/yoall/", "api_token={{"+apikey+"}}", b)
	checkErr(err, "http.Post() fatal")
	return true
}

func getLastTweetId(db *sql.DB) int64 {
	var tweetId int64

	stmt := "select max(tweetID) from transitdb;"
	rows, err := db.Query(stmt)
	checkErr(err, "db.Query() failed: "+stmt)
	defer db.Close()

	for rows.Next() {
		rows.Scan(&tweetId)
	}

	return tweetId
}

func insertRec(db *sql.DB, tweetLog map[string][]string) bool {

	lastTweet := strconv.FormatInt(getLastTweetId(db), 10)
	yoFlag := false

	for k, v := range tweetLog {
		if k > lastTweet {
			fmt.Printf("Inserting into db : %s|%s|%s|%s|%s\n", k, v[0], v[1], v[2], v[3])
			_, err := db.Exec("INSERT INTO transitdb (tweetId, timestamp, transitLine, url, yod) VALUES (?, ?, ?, ?, ?);", k, v[0], v[1], v[2], v[3])
			checkErr(err, "db.Exec() fatal!")
			yoFlag = true
		}
	}

	if yoFlag {
		return true
	} else {
		return false
	}
}

type ApiKeys struct {
	ConsumerKey       string
	ConsumerSecret    string
	AccessToken       string
	AccessTokenSecret string
	Yo_Apikey         string
	Yo_ApiUser        string
}

func main() {

	apikeys := ApiKeys{
		os.Getenv("TWITTER_CONSUMERKEY"),
		os.Getenv("TWITTER_CONSUMERSECRET"),
		os.Getenv("TWITTER_ACCESSTOKEN"),
		os.Getenv("TWITTER_ACCESSTOKENSECRET"),
		os.Getenv("YO_APIUSER"),
		os.Getenv("YO_APIKEY"),
	}

	anaconda.SetConsumerKey(apikeys.ConsumerKey)
	anaconda.SetConsumerSecret(apikeys.ConsumerSecret)
	api := anaconda.NewTwitterApi(apikeys.AccessToken, apikeys.AccessTokenSecret)

	username := "NJTRANSIT_ME"

	v := url.Values{}
	v.Set("count", "20")
	v.Set("screen_name", username)

	tweets, err := api.GetUserTimeline(v)
	checkErr(err, "api.GetUserTimeline() failed check connection or credentials")

	// Open the sqlite db
	db, err := sql.Open("sqlite3", "./db/transit.db")
	checkErr(err, "sql.Open() failed!")
	defer db.Close()

	tweetLog := make(map[string][]string) // Create a map to store our returned results
	for _, tweet := range tweets {
		url := getStatusURL(username, tweet.Id)
		tweetId := strconv.FormatInt(tweet.Id, 10)
		tweetLog[tweetId] = []string{
			tweet.CreatedAt,
			username,
			url,
			"1",
		}
	}

	if insertRec(db, tweetLog) {
		yO(apikeys.Yo_Apikey)
	}
	// fmt.Println(getLastTweetId(db))
}
