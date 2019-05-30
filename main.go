package main

import (
	"flag"
	"fmt"
	"net/url"

	"github.com/ChimeraCoder/anaconda"
)

var (
	consumerKey       = flag.String("consumer_key", "", "Issued from Twitter. ")
	consumerSecret    = flag.String("consumer_secret", "", "Issued from Twitter. ")
	accessToken       = flag.String("access_token", "", "Issued from Twitter. ")
	accessTokenSecret = flag.String("access_token_secret", "", "Issued from Twitter. ")
	user              = flag.String("user", "", "The account of user. ")
	// hashTag           = flag.String("hash_tag", "", "The name of hash tag.")
)

func getTwitterAPI() *anaconda.TwitterApi {
	anaconda.SetConsumerKey(*consumerKey)
	anaconda.SetConsumerSecret(*consumerSecret)
	return anaconda.NewTwitterApi(*accessToken, *accessTokenSecret)
}

func main() {
	flag.Parse()

	api := getTwitterAPI()

	v := url.Values{}
	v.Set("screen_name", *user)
	v.Set("count", "10")
	v.Set("exclude_replies", "true")
	v.Set("include_rts", "false")

	tweets, _ := api.GetUserTimeline(v)
	for _, tweet := range tweets {
		fmt.Println(tweet.Text)
	}

}
