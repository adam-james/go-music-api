package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/qor/validations"
)

// MAIN
func main() {
	db := initDB()
	validations.RegisterCallbacks(db)
	seedDB(db)

	router := gin.Default()

	router.GET("/artists", handleListArtists(db))
	router.POST("/artists", handleCreateArtist(db))
	router.GET("/artists/:id", handleGetArtist(db))

	// TODO rest of artist CRUD

	router.GET("/albums", handleListAlbums(db))
	router.POST("/albums", handleCreateAlbums(db))
	router.GET("/albums/:id", handleGetAlbum(db))
	router.PATCH("/albums/:id", handleUpdateAlbum(db))
	router.DELETE("/albums/:id", handleDeleteAlbum(db))

	// TODO CRUD tracks

	router.Run()
}

// INIT

func getPostgresConf() string {
	var dbConf string

	if os.Getenv("APP_ENV") == "production" {
		dbHost := os.Getenv("DB_HOST")
		dbPort := "5432"
		dbUser := os.Getenv("DB_USERNAME")
		dbName := os.Getenv("DB_DATABASE")
		dbPassword := os.Getenv("DB_PASSWORD")

		dbConf = fmt.Sprintf(
			"host=%s port=%s user=%s dbname=%s password=%s",
			dbHost,
			dbPort,
			dbUser,
			dbName,
			dbPassword,
		)
	} else {
		dbConf = "dbname=music_api_dev sslmode=disable"
	}

	return dbConf
}

func initDB() *gorm.DB {
	dbConf := getPostgresConf()

	db, err := gorm.Open("postgres", dbConf)
	if err != nil {
		log.Fatal(err)
	}

	db.AutoMigrate(&Album{})
	db.AutoMigrate(&Track{})
	db.AutoMigrate(&Artist{})

	return db
}

func seedDB(db *gorm.DB) {
	var album Album
	db.First(&album)
	if !db.NewRecord(&album) {
		return
	}

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

	db.Create(&Artist{Name: "Frank Ocean"})
	db.Create(&Artist{Name: "Neil Young"})
	db.Create(&Artist{Name: "Bob Dylan"})
}

// Artist Context

// Artist db model
type Artist struct {
	gorm.Model
	Name string `gorm:"not null" valid:"required"`
}

// ArtistVM is a view model for the artist
type ArtistVM struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// CreateArtistParams are params for creating an Artist
type CreateArtistParams struct {
	Name string `json:"name" binding:"required"`
}

func handleListArtists(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var artists []Artist
		db.Find(&artists)

		c.JSON(http.StatusOK, gin.H{
			"artists": renderArtists(artists),
		})
	}
}

func handleCreateArtist(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var params CreateArtistParams
		err := c.BindJSON(&params)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
				"error": err.Error(),
			})
			return
		}

		artist := Artist{Name: params.Name}
		errs := db.Create(&artist).GetErrors()
		if len(errs) > 0 {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "An internal server error occurred.",
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"artist": renderArtist(artist),
		})
	}
}

func handleGetArtist(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")
		var artist Artist
		db.Where("id = ?", id).First(&artist)
		if db.NewRecord(&artist) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"message": fmt.Sprintf("Cannot find artist with id %s", id),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"artist": renderArtist(artist),
		})
	}
}

func renderArtist(artist Artist) ArtistVM {
	return ArtistVM{
		ID:   artist.ID,
		Name: artist.Name,
	}
}

func renderArtists(artists []Artist) []ArtistVM {
	vms := make([]ArtistVM, len(artists))
	for i, a := range artists {
		vms[i] = renderArtist(a)
	}
	return vms
}

// ALBUM Context

// Album database model
type Album struct {
	gorm.Model
	Title string `gorm:"not null" valid:"required"`
	Year  uint   `gorm:"not null" valid:"required"`
}

// AlbumVM view model for rendering Album.
type AlbumVM struct {
	ID    uint   `json:"id"`
	Title string `json:"title"`
	Year  uint   `json:"year"`
}

// CreateAlbumParams represents params for creating an album.
type CreateAlbumParams struct {
	Title string `json:"title" binding:"required"`
	Year  uint   `json:"year" binding:"required"`
}

// UpdateAlbumParams represents params for updating an album.
type UpdateAlbumParams struct {
	Title string `json:"title"`
	Year  uint   `json:"year"`
}

func handleListAlbums(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		albums := listAlbums(db)()
		c.JSON(200, gin.H{
			"albums": renderAlbums(albums),
		})
	}
}

func handleCreateAlbums(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var params CreateAlbumParams
		err := c.BindJSON(&params)
		if err != nil {
			// TODO can you render this better?
			c.AbortWithStatusJSON(422, gin.H{
				"error": err.Error(),
			})
			return
		}

		album := Album{Title: params.Title, Year: params.Year}
		errors := db.Create(&album).GetErrors()
		if len(errors) > 0 {
			c.AbortWithStatusJSON(500, gin.H{
				"message": "An internal server error occurred.",
			})
			return
		}

		c.JSON(201, gin.H{
			"album": renderAlbum(album),
		})
	}
}

func handleGetAlbum(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")
		album, found := findAlbum(db)(id)
		if !found {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"message": fmt.Sprintf("Cannot find album with id %s", id),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"album": renderAlbum(album),
		})
	}
}

func handleUpdateAlbum(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")
		album, found := findAlbum(db)(id)
		if !found {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"message": fmt.Sprintf("Cannot find album with id %s", id),
			})
			return
		}

		var params UpdateAlbumParams
		err := c.BindJSON(&params)
		if err != nil {
			// TODO can you render this better?
			c.AbortWithStatusJSON(422, gin.H{
				"error": err.Error(),
			})
			return
		}

		// TODO what if this fails
		db.Model(&album).Update(Album{
			Title: params.Title,
			Year:  params.Year,
		})

		c.JSON(http.StatusOK, gin.H{
			"album": renderAlbum(album),
		})
	}
}

func handleDeleteAlbum(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")
		album, found := findAlbum(db)(id)
		if !found {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"message": fmt.Sprintf("Cannot find album with id %s", id),
			})
			return
		}

		db.Delete(&album)
		c.JSON(http.StatusOK, gin.H{
			"album": renderAlbum(album),
		})
	}
}

func findAlbum(db *gorm.DB) func(id string) (Album, bool) {
	return func(id string) (Album, bool) {
		var album Album
		db.Where("id = ?", id).First(&album)
		if db.NewRecord(&album) {
			return album, false
		}
		return album, true
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
		ID:    album.ID,
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
