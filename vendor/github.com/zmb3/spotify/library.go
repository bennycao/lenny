package spotify

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// UserHasTracks checks if one or more tracks are saved to the current user's
// "Your Music" library.  This call requires authorization.
func (c *Client) UserHasTracks(ids ...ID) ([]bool, error) {
	if l := len(ids); l == 0 || l > 50 {
		return nil, errors.New("spotify: UserHasTracks supports 1 to 50 IDs per call")
	}
	spotifyURL := fmt.Sprintf("%sme/tracks/contains?ids=%s", baseAddress, strings.Join(toStringSlice(ids), ","))
	resp, err := c.http.Get(spotifyURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, decodeError(resp.Body)
	}
	var result []bool
	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}

// AddTracksToLibrary saves one or more tracks to the current user's
// "Your Music" library.  This call requires authorization (the
// ScopeUserLibraryModify scope).
// A track can only be saved once; duplicate IDs are ignored.
func (c *Client) AddTracksToLibrary(ids ...ID) error {
	return c.modifyLibraryTracks(true, ids...)
}

// RemoveTracksFromLibrary removes one or more tracks from the current user's
// "Your Music" library.  This call requires authorization (the ScopeUserModifyLibrary
// scope).  Trying to remove a track when you do not have the user's authorization
// results in a `spotify.Error` with the status code set to http.StatusUnauthorized.
func (c *Client) RemoveTracksFromLibrary(ids ...ID) error {
	return c.modifyLibraryTracks(false, ids...)
}

func (c *Client) modifyLibraryTracks(add bool, ids ...ID) error {
	if l := len(ids); l == 0 || l > 50 {
		return errors.New("spotify: this call supports 1 to 50 IDs per call")
	}
	spotifyURL := fmt.Sprintf("%sme/tracks?ids=%s", baseAddress, strings.Join(toStringSlice(ids), ","))
	method := "DELETE"
	if add {
		method = "PUT"
	}
	req, err := http.NewRequest(method, spotifyURL, nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return decodeError(resp.Body)
	}
	return nil
}