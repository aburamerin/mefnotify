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

// NewDB создает Posts для хранения постов в базе данных postgres.
func NewDB(DSN string) (*sql.DB, error) {
	sqlDB, err := sql.Open("pgx", DSN)
	if err != nil {
		log.Printf("Unable to connect to database: %v\n", err)
		return nil, err
	}

	schemaSQL := `
  CREATE TABLE IF NOT EXISTS posts (
      id INT NOT NULL PRIMARY KEY,
      time TIMESTAMP,
      preview varchar(256),
      content TEXT,
      url varchar(256),
      author varchar(256)
  );
  `

	if _, err = sqlDB.Exec(schemaSQL); err != nil {
		return nil, err
	}

	return sqlDB, nil
}

func StorePost(db *sql.DB, post Post) {
	sqlAddPost := `
	INSERT INTO posts (
		id,
    time,
    preview,
    content,
    url,
    author
	) values
  (
    $1::int,
    $2::timestamp,
    $3::varchar,
    $4::text,
    $5::varchar,
    $6::varchar
  )
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
	sqlFindPost := `SELECT author FROM posts WHERE id = $1::int`

	log.Printf("searching %d", ID)
	var Author string
	log.Println(db)

	stmt, err := db.Prepare(sqlFindPost)
	if err != nil {
		log.Println(err)
		return false
	}
	defer stmt.Close()

	err = stmt.QueryRow(ID).Scan(&Author)
	return err == nil
}
