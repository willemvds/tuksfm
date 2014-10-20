package tuksfm

import (
	"database/sql"
	"fmt"
	"time"
)

type Artist struct {
	Id   int
	Name string
}

func (a *Artist) SaveToDB(db *sql.DB) (int, error) {
	var id int
	if a.Id == 0 {
		err := db.QueryRow(`INSERT INTO artist(name) VALUES($1) RETURNING id`, a.Name).Scan(&id)
		if err != nil {
			return 0, nil
		}
		a.Id = id
	}
	return id, nil
}

func (a *Artist) Equals(a2 *Artist) bool {
	if a2 == nil {
		return false
	}
	if a.Name == a2.Name {
		return true
	}
	return false
}

type Song struct {
	Id int
	*Artist
	Name  string
	Plays int
}

func (s *Song) SaveToDB(db *sql.DB) (int, error) {
	var id int
	if s.Id == 0 {
		err := db.QueryRow(`INSERT INTO song(artist_id, name) VALUES($1, $2) RETURNING id`, s.Artist.Id, s.Name).Scan(&id)
		if err != nil {
			return 0, nil
		}
		s.Id = id
	}
	return id, nil
}

func (s Song) String() string {
	return fmt.Sprintf("%s by %s", s.Name, s.Artist.Name)
}

func (s *Song) Equals(s2 *Song) bool {
	if s2 == nil {
		return false
	}
	if s.Name == s2.Name && s.Artist.Equals(s2.Artist) {
		return true
	}
	return false
}

type Artists []*Artist

func (a Artists) Find(name string) *Artist {
	for i := range a {
		if a[i].Name == name {
			return a[i]
		}
	}
	return nil
}

func (a *Artists) Add(artist *Artist) {
	*a = append(*a, artist)
}

type Songs []*Song

func (s Songs) Find(name string, artist *Artist) *Song {
	for i := range s {
		if s[i].Name == name && s[i].Artist == artist {
			return s[i]
		}
	}
	return nil
}

func (s *Songs) Add(song *Song) {
	*s = append(*s, song)
}

type SongPlay struct {
	Id int
	*Song
	AddTime time.Time
}

type Playlist []*SongPlay

func (pl *Playlist) Add(song *Song) {
	sp := &SongPlay{Song: song, AddTime: time.Now()}
	*pl = append(*pl, sp)
}

func (pl Playlist) Last() *Song {
	if len(pl) > 0 {
		return pl[len(pl)-1].Song
	}
	return nil
}

func (pl *Playlist) SaveLastPlayToDB(db *sql.DB) (int, error) {
	var id int
	err := db.QueryRow(`INSERT INTO playlist(song_id, time) VALUES($1, $2) RETURNING id`, []*SongPlay(*pl)[len([]*SongPlay(*pl))-1].Song.Id, time.Now()).Scan(&id)
	return id, err
}
