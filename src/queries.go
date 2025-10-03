package utils

import (
    "database/sql"
    "fmt"
    "log"
    "os"

    _ "github.com/mattn/go-sqlite3"
)

type Page struct {
    ID       int
    Title    string
    Language string
    Content  string
}

func RunQueries() {
    // Get database URL from environment or use default
    dbURL := os.Getenv("DATABASE_URL")
    if dbURL == "" {
        dbURL = "data/oggole.db"
    }

    db, err := sql.Open("sqlite3", dbURL)
    if err != nil {
        log.Fatalf("Failed to open database: %v", err)
    }
    defer db.Close()

    lastID, err := InsertUserQuery(db)
    if err != nil {
        log.Printf("InsertUserQuery error: %v", err)
    } else {
        fmt.Printf("InsertUserQuery: Inserted user with id %d\n", lastID)
    }

    userID, err := GetUserIDQuery(db)
    if err != nil {
        log.Printf("GetUserIDQuery error: %v", err)
    } else {
        fmt.Printf("GetUserIDQuery: User 'johndoe' has id %d\n", userID)
    }

    id, username, email, password, err := GetUserByIDQuery(db)
    if err != nil {
        log.Printf("GetUserByIDQuery error: %v", err)
    } else {
        fmt.Printf("GetUserByIDQuery: id=%d username=%s email=%s password=%s\n", id, username, email, password)
    }

    id, username, email, password, err = GetUserByUsernameQuery(db)
    if err != nil {
        log.Printf("GetUserByUsernameQuery error: %v", err)
    } else {
        fmt.Printf("GetUserByUsernameQuery: id=%d username=%s email=%s password=%s\n", id, username, email, password)
    }

    pages, err := SearchPagesQuery(db)
    if err != nil {
        log.Printf("SearchPagesQuery error: %v", err)
    } else {
        for _, page := range pages {
            fmt.Printf("SearchPagesQuery: id=%d title=%s language=%s content=%s\n", page.ID, page.Title, page.Language, page.Content)
        }
    }
}

func InsertUserQuery(db *sql.DB) (int64, error) {
    query := "INSERT INTO users (username, email, password) VALUES (?, ?, ?)"
    res, err := db.Exec(query, "johndoe", "john@example.com", "5f4dcc3b5aa765d61d8327deb882cf99")
    if err != nil {
        return 0, err
    }
    lastID, err := res.LastInsertId()
    if err != nil {
        return 0, err
    }
    return lastID, nil
}

func GetUserIDQuery(db *sql.DB) (int, error) {
    query := "SELECT id FROM users WHERE username = ?"
    var id int
    err := db.QueryRow(query, "johndoe").Scan(&id)
    if err != nil {
        return 0, err
    }
    return id, nil
}

func GetUserByIDQuery(db *sql.DB) (int, string, string, string, error) {
    query := "SELECT * FROM users WHERE id = ?"
    row := db.QueryRow(query, 1)
    var id int
    var username, email, password string
    err := row.Scan(&id, &username, &email, &password)
    if err != nil {
        return 0, "", "", "", err
    }
    return id, username, email, password, nil
}

func GetUserByUsernameQuery(db *sql.DB) (int, string, string, string, error) {
    query := "SELECT * FROM users WHERE username = ?"
    row := db.QueryRow(query, "johndoe")
    var id int
    var username, email, password string
    err := row.Scan(&id, &username, &email, &password)
    if err != nil {
        return 0, "", "", "", err
    }
    return id, username, email, password, nil
}

func SearchPagesQuery(db *sql.DB) ([]Page, error) {
    query := "SELECT * FROM pages WHERE language = ? AND content LIKE ?"
    rows, err := db.Query(query, "en", "%golang%")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var pages []Page
    for rows.Next() {
        var page Page
        err := rows.Scan(&page.ID, &page.Title, &page.Language, &page.Content)
        if err != nil {
            log.Printf("SearchPagesQuery row scan error: %v", err)
            continue
        }
        pages = append(pages, page)
    }
    if err := rows.Err(); err != nil {
        return nil, err
    }
    return pages, nil
}
