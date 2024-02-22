package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
)

var db *sql.DB

type Album struct {
    ID     string  `json:"id"`
    Title  string  `json:"title"`
    Artist string  `json:"artist"`
    Price  float32 `json:"price"`
}

func main() {
	   // Open database connection.
		setupMysql()
    router := gin.Default()
    router.GET("/albums", getAlbums)
    router.GET("/albums/:id", getAlbumByID)
    router.POST("/albums", postAlbums)

    router.Run("localhost:8080")
}

func setupMysql() {
    // Capture connection properties.
    cfg := mysql.Config{
        User:   "root",
        Passwd: "",
        Net:    "tcp",
        Addr:   "127.0.0.1:3306",
        DBName: "recordings",
				AllowNativePasswords: true,
    }
    // Get a database handle.
    var err error
    db, err = sql.Open("mysql", cfg.FormatDSN())
    if err != nil {
        log.Fatal(err)
    }

    pingErr := db.Ping()
    if pingErr != nil {
        log.Fatal(pingErr)
    }
    fmt.Println("Connected!")
}

// getAlbums responds with the list of all albums as JSON.
func getAlbums(c *gin.Context) {
		allalb, err := allAlbums()
		if err != nil {
				log.Fatal(err)
		}
		// fmt.Printf("All album: %v\n", allalb)
    c.IndentedJSON(http.StatusOK, allalb)
}

// postAlbums adds an album from JSON received in the request body.
func postAlbums(c *gin.Context) {
    var newAlbum Album

    // Call BindJSON to bind the received JSON to
    // newAlbum.
    if err := c.BindJSON(&newAlbum); err != nil {
        return
    }

		albID, err := addAlbum(newAlbum)
		if err != nil {
				log.Fatal(err)
		}
    c.IndentedJSON(http.StatusCreated, albID)
}

// getAlbumByID locates the album whose ID value matches the id
// parameter sent by the client, then returns that album as a response.
func getAlbumByID(c *gin.Context) {
    id := c.Param("id")

		i, err := strconv.ParseInt(id, 10, 64)
    if err != nil {
        // ... handle error
        panic(err)
    }


		alb, err := albumByID(i)
		if err != nil {
				c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
				// log.Fatal(err)
		}else{
			c.IndentedJSON(http.StatusOK, alb)
		}
}

func allAlbums() ([]Album, error) {
    // An albums slice to hold data from returned rows.
    var albums []Album

    rows, err := db.Query("SELECT * FROM album")
    if err != nil {
        return nil, fmt.Errorf("albumsByArtistQuery %q", err)
    }

		// Defer closing rows so that any resources it holds will be released when the function exits.
    defer rows.Close()
    // Loop through rows, using Scan to assign column data to struct fields.
		// Scan takes a list of pointers to Go values, where the column values will be written. Here, you pass pointers to fields in the alb variable, created using the & operator. Scan writes through the pointers to update the struct fields.
    for rows.Next() {
        var alb Album

        if err := rows.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price); err != nil {
            return nil, fmt.Errorf("albumsByArtistScan %q", err)
        }
        albums = append(albums, alb)
    }
		// error from the overall query, using rows.Err. if the query itself fails, checking for error here is only way to find out results are incomplete
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("albumsByArtistRows %q",err)
    }
    return albums, nil
}

// albumsByArtist queries for albums that have the specified artist name.
func albumsByArtist(name string) ([]Album, error) {
    // An albums slice to hold data from returned rows.
    var albums []Album

    rows, err := db.Query("SELECT * FROM album WHERE artist = ?", name)
    if err != nil {
        return nil, fmt.Errorf("albumsByArtistQuery %q: %v", name, err)
    }

    defer rows.Close()
    for rows.Next() {
        var alb Album

        if err := rows.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price); err != nil {
            return nil, fmt.Errorf("albumsByArtistScan %q: %v", name, err)
        }
        albums = append(albums, alb)
    }
		// error from the overall query, using rows.Err. if the query itself fails, checking for error here is only way to find out results are incomplete
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("albumsByArtistRows %q: %v", name, err)
    }
    return albums, nil
}

// albumByID queries for the album with the specified ID.
func albumByID(id int64) (Album, error) {
    // An album to hold data from the returned row.
    var alb Album

    row := db.QueryRow("SELECT * FROM album WHERE id = ?", id)
    if err := row.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price); err != nil {
        if err == sql.ErrNoRows {
            return alb, fmt.Errorf("albumsById %d: no such album", id)
        }
        return alb, fmt.Errorf("albumsById %d: %v", id, err)
    }
    return alb, nil
}

// addAlbum adds the specified album to the database,
// returning the album ID of the new entry
func addAlbum(alb Album) (int64, error) {
    result, err := db.Exec("INSERT INTO album (title, artist, price) VALUES (?, ?, ?)", alb.Title, alb.Artist, alb.Price)
    if err != nil {
        return 0, fmt.Errorf("addAlbum: %v", err)
    }
    id, err := result.LastInsertId()
    if err != nil {
        return 0, fmt.Errorf("addAlbum: %v", err)
    }
    return id, nil
}

