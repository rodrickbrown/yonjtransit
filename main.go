package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"

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

func insertRec(db *sql.DB) bool {
	tx, err := db.Begin()
	checkErr(err, "db.Begin() failed!")
	defer db.Close()

	stmt, err := tx.Prepare("insert into transitdb(timestamp, tweetid, transitline, yod) values()")
	defer stmt.Close()
	return true
}

type ApiKeys struct {
	ConsumerKey       string
	ConsumerSecret    string
	AccessToken       string
	AccessTokenSecret string
}

func main() {

	twtapikeys := ApiKeys{
		os.Getenv("CONSUMERKEY"),
		os.Getenv("CONSUMERSECRET"),
		os.Getenv("ACCESSTOKEN"),
		os.Getenv("ACCESSTOKENSECRET"),
	}

	anaconda.SetConsumerKey(twtapikeys.ConsumerKey)
	anaconda.SetConsumerSecret(twtapikeys.ConsumerSecret)
	api := anaconda.NewTwitterApi(twtapikeys.AccessToken, twtapikeys.AccessTokenSecret)

	username := "NJTRANSIT_ME"

	v := url.Values{}
	v.Set("count", "20")
	v.Set("screen_name", username)

	tweets, err := api.GetUserTimeline(v)
	checkErr(err, "api.GetUserTimeline() failed check connection or credentials")

	db, err := sql.Open("sqlite3", "./db/transit.db")
	checkErr(err, "sql.Open() failed!")
	defer db.Close()

	tweetLog := make(map[string]string) // Create a map to store our returned results
	for _, tweet := range tweets {
		url := getStatusURL(username, tweet.Id)
		// fmt.Println(tweet.CreatedAt, url)
		tweetLog[url] = tweet.CreatedAt
	}

	fmt.Println(getLastTweetId(db))
}
