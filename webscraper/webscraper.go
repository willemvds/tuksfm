package webscraper

import (
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"
)

const songListUrl = "http://tuksfm.co.za/Forms/SongList.aspx"

type song struct {
	Name string
	Artist string
}

func getHtml() ([]byte, error) {
	resp, err := http.Get(songListUrl)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(resp.Body)
}

func parseHtml(html []byte) ([]song, error) {
	songs := make([]song, 0)
	re := regexp.MustCompile(`<td class="text1">(.*)</td><td class="text2">(.*)</td>`)
	matches := re.FindAllSubmatch(html, -1)
	for _, match := range matches {
		if len(match) != 3 {
			return nil, errors.New("Incorrect number of matches, aw well")
		}
		songs = append(songs, song{Name: string(match[1]), Artist: string(match[2])})
	}
	return songs, nil
}

func GetSongList() ([]song, error) {
	html, err := getHtml()
	if err != nil {
		return nil, err
	}
	return parseHtml(html)
}

