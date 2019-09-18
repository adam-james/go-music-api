package main

// TODO add routes
// work through routes one at a time, end to end

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// MAIN
func main() {
	db := initDB()
	seedDB(db)

	router := gin.Default()
	router.GET("/albums", handleListAlbums(db))
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	router.Run()
}

// INIT

func initDB() *gorm.DB {
	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		log.Fatal(err)
	}
	db.AutoMigrate(&Album{})
	db.AutoMigrate(&Track{})
	return db
}

func seedDB(db *gorm.DB) {
	var album Album
	db.First(&album)
	if !db.NewRecord(&album) {
		return
	}

	// TODO Seed data here
	album = createAlbum(db)("Blonde", 2016)
	createAlbum(db)("Blonde on Blonde", 1966)
	createAlbum(db)("Harvest Moon", 1992)

	tracks := []Track{
		Track{AlbumID: album.ID, Title: "Nikes", TrackNumber: 1},
		Track{AlbumID: album.ID, Title: "Ivy", TrackNumber: 2},
		Track{AlbumID: album.ID, Title: "Pink + White", TrackNumber: 3},
		Track{AlbumID: album.ID, Title: "Be Yourself", TrackNumber: 4},
		Track{AlbumID: album.ID, Title: "Solo", TrackNumber: 5},
		Track{AlbumID: album.ID, Title: "Skyline To", TrackNumber: 6},
		Track{AlbumID: album.ID, Title: "Self Control", TrackNumber: 7},
		Track{AlbumID: album.ID, Title: "Good Guy", TrackNumber: 8},
		Track{AlbumID: album.ID, Title: "Solo (Reprise)", TrackNumber: 9},
	}

	for _, t := range tracks {
		db.Create(&t)
	}
}

// ALBUM Context

// Album database model
type Album struct {
	gorm.Model
	Title string `gorm:"not null"`
	Year  uint   `gorm:"not null"`
}

// AlbumVM view model for rendering Album.
type AlbumVM struct {
	Title string `json:"title"`
	Year  uint   `json:"year"`
}

// TODO validations would be nice

func handleListAlbums(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		albums := listAlbums(db)()
		c.JSON(200, gin.H{
			"albums": renderAlbums(albums),
		})
	}
}

func listAlbums(db *gorm.DB) func() []Album {
	return func() []Album {
		var albums []Album
		db.Find(&albums)
		return albums
	}
}

func renderAlbums(albums []Album) []AlbumVM {
	vms := make([]AlbumVM, len(albums))
	for i, a := range albums {
		vms[i] = renderAlbum(a)
	}
	return vms
}

func renderAlbum(album Album) AlbumVM {
	return AlbumVM{
		Title: album.Title,
		Year:  album.Year,
	}
}

func createAlbum(db *gorm.DB) func(title string, year uint) Album {
	return func(title string, year uint) Album {
		album := Album{Title: title, Year: year}
		db.Create(&album)
		return album
	}
}

func getAlbum(db *gorm.DB) func(id uint) (Album, error) {
	return func(id uint) (Album, error) {
		var album Album
		db.Where("id = ?", id).First(&album)
		if db.NewRecord(&album) {
			return album, fmt.Errorf("Cannot find album with id %d", id)
		}
		return album, nil
	}
}

func updateAlbum(db *gorm.DB) func(id uint, attrs *Album) (Album, error) {
	return func(id uint, attrs *Album) (Album, error) {
		album, err := getAlbum(db)(id)
		db.Model(&album).Updates(Album{Title: attrs.Title, Year: attrs.Year})
		return album, err
	}
}

func deleteAlbum(db *gorm.DB) func(id uint) (Album, error) {
	return func(id uint) (Album, error) {
		album, err := getAlbum(db)(id)
		db.Delete(&album)
		return album, err
	}
}

// TRACK

// Track DB model
type Track struct {
	gorm.Model
	// TODO not sure this does anything
	Title       string `gorm:"not null"`
	AlbumID     uint   `gorm:"not null"`
	TrackNumber uint   `gorm:"not null"`
}

func createTrack(db *gorm.DB) func(track Track) (Track, error) {
	return func(track Track) (Track, error) {
		db.Create(&track)
		if db.NewRecord(track) {
			return track, fmt.Errorf("Could not created track with title %s", track.Title)
		}
		return track, nil
	}
}

func listAlbumTracks(db *gorm.DB) func(id uint) []Track {
	return func(id uint) []Track {
		var tracks []Track
		db.Where("album_id = ?", id).Find(&tracks)
		return tracks
	}
}
