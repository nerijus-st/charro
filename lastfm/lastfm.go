package lastfm

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
)

//Tracks is result from last.fm API
type Tracks struct {
	Toptracks struct {
		Tracks []struct {
			Artist struct {
				Name string `json:"name"`
			} `json:"artist"`
			Name string `json:"name"`
		} `json:"track"`
	} `json:"toptracks"`
}

//Form stores values from lastfm input form.
type Form struct {
	Username string
	Period   string
	Limit    string
}

//Error stores last.fm error message
type Error struct {
	Message string `json:"message"`
}

//GetTopTracks makes a call to last.fm
func GetTopTracks(username, period, limit *string) (*Tracks, error) {

	apiKey := os.Getenv("LASTFM_API_KEY")
	url := "http://ws.audioscrobbler.com/2.0/?method=user.gettoptracks&user=" + *username + "&api_key=" + apiKey + "&format=json&limit=" + *limit + "&period=" + *period

	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusOK {
		tracks := Tracks{}

		if err := json.Unmarshal(body, &tracks); err != nil {
			return nil, err
		}

		return &tracks, err
	}

	e := Error{}
	if err := json.Unmarshal(body, &e); err != nil {
		return nil, err
	}

	return nil, errors.New("Last.fm " + e.Message)
}
