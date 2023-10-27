package posts

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

type PostInfo struct {
	PostDate time.Time
	URL      string
}

type Post struct {
	ID      int64
	Info    PostInfo
	Author  string
	Preview string
	Content string
}

type PostsSlice []Post

func (p *Post) String() string {
	return fmt.Sprintf("*%s:* %s", p.Author, p.Preview)
}

// NewDB создает Posts для хранения постов в базе данных SQLite.
func NewDB(dbFile string) (*sql.DB, error) {
	sqlDB, err := sql.Open("sqlite3", dbFile)
	println("sqldb:", sqlDB)

	if err != nil {
		return nil, err
	}

	schemaSQL := `
  CREATE TABLE IF NOT EXISTS posts (
      id INT NOT NULL PRIMARY KEY,
      time TIMESTAMP,
      preview VARCHAR(256),
      content TEXT,
      url TEXT,
      author VARCHAR(256)
  );
  `

	if _, err = sqlDB.Exec(schemaSQL); err != nil {
		return nil, err
	}

	return sqlDB, nil
}

func StorePost(db *sql.DB, post Post) {
	sqlAddPost := `
	INSERT OR REPLACE INTO posts (
		id,
    time,
    preview,
    content,
    url,
    author
	) values(?, ?, ?, ?, ?, ?)
	`

	stmt, err := db.Prepare(sqlAddPost)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		post.ID,
		post.Info.PostDate,
		post.Preview,
		post.Content,
		post.Info.URL,
		post.Author,
	)
	if err != nil {
		panic(err)
	}
}

func FindPost(db *sql.DB, ID int64) bool {
	sqlFindPost := `SELECT author FROM posts WHERE id = ?`

	log.Printf("searching %d", ID)
	var Author string
	log.Println(db)

	stmt, err := db.Prepare(sqlFindPost)
	if err != nil {
		log.Println("start error")
		log.Println(err)
		log.Println("eof error")
		return false
	}
	defer stmt.Close()

	err = stmt.QueryRow(ID).Scan(&Author)
	return err == nil
}
