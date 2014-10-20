package main

import (
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"time"

	"database/sql"
	_ "github.com/lib/pq"

	"github.com/willemvds/tuksfm"
	"github.com/willemvds/tuksfm/webscraper"
)

const artistsPath = "artists.gob"
const songsPath = "songs.gob"
const playlistPath = "playlist.gob"

func Persist(what interface{}, where io.Writer) error {
	encoder := gob.NewEncoder(where)
	return encoder.Encode(what)
}

type PersistJob struct {
	filename string
	data     interface{}
}

func NewPersistJob(filename string, data interface{}) *PersistJob {
	job := PersistJob{}
	job.filename = filename
	job.data = data
	return &job
}

type PersistWorker chan *PersistJob

func NewPersistWorker() PersistWorker {
	worker := make(PersistWorker, 0)
	return worker
}

func (worker PersistWorker) Start() {
	go func() {
		for {
			job := <-worker
			f, err := os.Create(job.filename)
			if err != nil {
				fmt.Println(err)
				continue
			}
			err = Persist(job.data, f)
			if err != nil {
				fmt.Println(err)
			}
		}
	}()
}

func LoadData(artists *tuksfm.Artists, songs *tuksfm.Songs, playlist *tuksfm.Playlist) error {
	f, err := os.Open(artistsPath)
	if err != nil {
		return err
	}
	decoder := gob.NewDecoder(f)
	decoder.Decode(artists)

	f, err = os.Open(songsPath)
	if err != nil {
		return err
	}
	decoder = gob.NewDecoder(f)
	decoder.Decode(songs)

	f, err = os.Open(playlistPath)
	if err != nil {
		return err
	}
	decoder = gob.NewDecoder(f)
	decoder.Decode(playlist)
	return nil
}

func GetDbConn() (*sql.DB, error) {
	return sql.Open("postgres", "user=tuks dbname=tuksfm sslmode=disable password=webd3v port=5434")
}

func main() {
	artists := make(tuksfm.Artists, 0)
	songs := make(tuksfm.Songs, 0)
	playlist := make(tuksfm.Playlist, 0)
	err := LoadData(&artists, &songs, &playlist)
	if err != nil {
		fmt.Println("Failed to load data, start over...", err)
	}
	fmt.Println(playlist)

	pworker := NewPersistWorker()
	pworker.Start()

	dbconn, dberr := GetDbConn()
	fmt.Println(dbconn)
	fmt.Println(dberr)

	for {
		newArtists := false
		newSongs := false
		newSongPlays := false
		websongs, err := webscraper.GetSongList()
		if err != nil {
			fmt.Println("Error loading songs from tuks website", err)
		} else {
			newstack := make([]*tuksfm.Song, 0)
			for i := range websongs {
				artist := artists.Find(websongs[i].Artist)
				if artist == nil {
					artist = &tuksfm.Artist{Name: websongs[i].Artist}
					artists.Add(artist)
					fmt.Println(artist.SaveToDB(dbconn))
					newArtists = true
				}
				song := songs.Find(websongs[i].Name, artist)
				if song == nil {
					song = &tuksfm.Song{Name: websongs[i].Name, Artist: artist}
					songs.Add(song)
					fmt.Println(song.SaveToDB(dbconn))
					newSongs = true
				}
				if song.Equals(playlist.Last()) {
					break
				}
				newstack = append(newstack, song)
			}
			for i := len(newstack) - 1; i >= 0; i-- {
				playlist.Add(newstack[i])
				fmt.Println(playlist.SaveLastPlayToDB(dbconn))
				newSongPlays = true
				fmt.Println("Adding ", newstack[i])
			}
			if newArtists {
				pworker <- NewPersistJob(artistsPath, artists)
			}
			if newSongs {
				pworker <- NewPersistJob(songsPath, songs)
			}
			if newSongPlays {
				pworker <- NewPersistJob(playlistPath, playlist)
				fmt.Println("<PLAYLIST>")
				for _, song := range playlist {
					fmt.Println(song)
				}
				fmt.Println("</PLAYLIST>")
			}
		}
		time.Sleep(10 * time.Second)
	}
}
