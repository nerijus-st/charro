package spotify

import (
	"fmt"
	"time"

	"github.com/nerijus-st/charro/lastfm"

	spotifyWrapper "github.com/zmb3/spotify"
)

//Tracks stores spotify tracks duuuh.
type Tracks struct {
	SearchResults []spotifyWrapper.SearchResult
}

//TrackIDs holds ids of tracks
type TrackIDs []spotifyWrapper.ID

//Form stores values from spotify input form.
type Form struct {
	Limit     string
	TimeRange string
}

//GetTopTracks gets top tracks for user at given time range and songs limit. Returns tracksIDs
func GetTopTracks(client *spotifyWrapper.Client, timeRange string, limit int) (*TrackIDs, error) {

	opt := spotifyWrapper.Options{
		Limit:     &limit,
		Timerange: &timeRange,
	}

	trackIDs := TrackIDs{}

	results, err := client.CurrentUsersTopTracksOpt(&opt)
	if err != nil {
		return nil, err
	}

	for _, track := range results.Tracks {
		trackIDs = append(trackIDs, track.ID)
	}

	return &trackIDs, nil

}

//GetTracksBasedOnLastFM searches for spotify tracks based on given last.fm tracks
func GetTracksBasedOnLastFM(client *spotifyWrapper.Client, tracks *lastfm.Tracks) (*TrackIDs, error) {

	limit := int(1)
	var opt = spotifyWrapper.Options{
		Limit: &limit,
	}

	trackIDs := TrackIDs{}

	for _, trk := range tracks.Toptracks.Tracks {
		query := "track:" + trk.Name + " artist:" + trk.Artist.Name
		results, err := client.SearchOpt(query, spotifyWrapper.SearchTypeTrack, &opt)
		if err != nil {
			return nil, err
		}

		for _, track := range results.Tracks.Tracks {
			trackIDs = append(trackIDs, track.ID)
		}
	}

	return &trackIDs, nil

}

//GeneratePlaylist creates playlist for a user
func GeneratePlaylist(client *spotifyWrapper.Client, trackIDs *TrackIDs, period, limit *string) (*spotifyWrapper.FullPlaylist, error) {

	user, err := client.CurrentUser()
	if err != nil {
		return nil, err
	}

	timeNow := time.Now()
	playlistName := fmt.Sprintf("Charro %s Top%s (%s)", *period, *limit, timeNow.Format("Jan 2 2006"))
	playlistDesc := fmt.Sprintf("Charro Top%s generated playlist at %s based on %s term.", *limit, timeNow.Format("2006.01.02 15:04:05"), *period)

	playlist, err := client.CreatePlaylistForUser(user.ID, playlistName, playlistDesc, false)
	if err != nil {
		return nil, err
	}

	_, err = client.AddTracksToPlaylist(playlist.ID, *trackIDs...)
	if err != nil {
		return nil, err
	}

	return playlist, nil
}
