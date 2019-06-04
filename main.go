package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ChimeraCoder/anaconda"
)

var (
	consumerKey       = flag.String("consumer_key", "", "Issued from Twitter. ")
	consumerSecret    = flag.String("consumer_secret", "", "Issued from Twitter. ")
	accessToken       = flag.String("access_token", "", "Issued from Twitter. ")
	accessTokenSecret = flag.String("access_token_secret", "", "Issued from Twitter. ")
	user              = flag.String("user", "none", "The account of user. ")
	friendsPath       = flag.String("friends_path", "/tmp", "The path of the lists of user's friends. ")
	accountsPath      = flag.String("accounts_path", "/tmp", "The path of the lists of accounts. ")
	tweetsPath        = flag.String("tweets_path", "/tmp", "The path of the lists of tweets. ")
	// hashTag           = flag.String("hash_tag", "", "The name of hash tag.")
)

func getTwitterAPI() *anaconda.TwitterApi {
	anaconda.SetConsumerKey(*consumerKey)
	anaconda.SetConsumerSecret(*consumerSecret)
	return anaconda.NewTwitterApi(*accessToken, *accessTokenSecret)
}

func getUserTimeline(api *anaconda.TwitterApi, scname string, count int) []anaconda.Tweet {

	start := time.Now()

	v := url.Values{}
	v.Set("screen_name", scname)
	v.Set("count", strconv.Itoa(count))
	v.Set("exclude_replies", "true")
	v.Set("include_rts", "false")

	tweets, _ := api.GetUserTimeline(v)

	end := time.Now()

	d := end.Sub(start).Nanoseconds()
	wait := 1*1000*1000*1000 - d + (1 * 1000 * 1000)

	// fmt.Println("getUserTimeline")
	// fmt.Printf("%v ns\n", wait)

	time.Sleep(time.Duration(wait))

	return tweets
}

func writeFriends(api *anaconda.TwitterApi, scname string, path string) error {

	file, err := os.Create(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	file.Write(([]byte)("screen_name,name,url\n"))

	start := time.Now()
	ncur := "-1"
	for {
		v := url.Values{}
		v.Set("screen_name", scname)
		v.Set("count", "60")
		v.Set("cursor", ncur)

		cursor, _ := api.GetFriendsList(v)
		for _, user := range cursor.Users {
			ts := getUserTimeline(api, user.ScreenName, 20)
			lt6m := isLessThan6Mionth(ts, start)
			isPR := (strings.Index(user.Name, "公式") > -1) || (strings.Index(user.ScreenName, "official") > -1)
			url := fmt.Sprintf("https://twitter.com/%s", user.ScreenName)
			fmt.Printf("user=%s, protected=%v, last6mon=%v, isPR=%v \n", user.ScreenName, user.Protected, lt6m, isPR)

			if lt6m && !isPR {
				file.Write(([]byte)(fmt.Sprintf("%s,\"%s\",%s\n", user.ScreenName, user.Name, url)))
			}
		}
		ncur = cursor.Next_cursor_str
		fmt.Printf("next: %s\n", ncur)
		if cursor.Next_cursor == 0 {
			break
		}
	}

	end := time.Now()
	fmt.Println("writeFriends")
	fmt.Printf("%f秒\n", (end.Sub(start)).Seconds())
	return nil
}

func isLessThan6Mionth(tweets []anaconda.Tweet, now time.Time) bool {
	lt6m := false
	if len(tweets) > 0 {
		if t, err := tweets[0].CreatedAtTime(); err == nil {
			if now.Sub(t).Hours() < 24*30*6 {
				lt6m = true
			} else {
				lt6m = false
			}
		}
	}
	return lt6m
}

const (
	ScreenName = 0
	Name       = 1
	URL        = 2
	Sex        = 3
	IsEngineer = 4
)

func writeTweetsByAccountList(api *anaconda.TwitterApi, inPath string, outPath string) error {
	out, err := os.Create(outPath)
	if err != nil {
		return nil
	}
	defer out.Close()

	header := []string{"tweet", "screen_name", "sex", "is_engineer"}

	in, err := os.Open(inPath)
	if err != nil {
		return nil
	}
	defer in.Close()

	reader := csv.NewReader(in)
	reader.LazyQuotes = true

	writer := csv.NewWriter(out)
	writer.Write(header)

	line := -1
	for {
		cols, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		line++
		if line == 0 { // ヘッダー行
			continue
		}

		twts := getUserTimeline(api, cols[ScreenName], 200)
		for _, t := range twts {
			writer.Write([]string{convNewline(t.Text, " "), cols[ScreenName], cols[Sex], cols[IsEngineer]})
		}
	}
	return nil
}
func convNewline(str, nlcode string) string {
	return strings.NewReplacer(
		"\r\n", nlcode,
		"\r", nlcode,
		"\n", nlcode,
	).Replace(str)
}

func main() {
	flag.Parse()

	api := getTwitterAPI()

	// if err := writeFriends(api, *user, *friendsPath); err != nil {
	// 	fmt.Errorf("[writeFriends] error occured: %v", err)
	// }

	if err := writeTweetsByAccountList(api, *accountsPath, *tweetsPath); err != nil {
		fmt.Errorf("[writeTweetsByAccountList] error occured: %v", err)
	}
}
