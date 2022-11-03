package models

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgx/v4/pgxpool"
	"time"
)

type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

type SnippetModel struct {
	DB *pgxpool.Pool
}

func (m *SnippetModel) Insert(title string, content string, expires int) (int, error) {
	stmt := `INSERT INTO snippets (title, content, created, expires)
VALUES($1, $2, current_timestamp, $3) RETURNING id`
	var lastInsertId int
	now := time.Now()
	err := m.DB.QueryRow(context.Background(), stmt, title, content, now.AddDate(0, 0, expires)).Scan(&lastInsertId)
	if err != nil {
		return 0, err
	}
	return lastInsertId, nil
}

func (m *SnippetModel) Get(id int) (*Snippet, error) {
	stmt := `SELECT id, title, content, created, expires FROM snippets
WHERE expires > current_date AND id = $1`

	row := m.DB.QueryRow(context.Background(), stmt, id)

	s := &Snippet{}

	err := row.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	return s, nil
}

func (m *SnippetModel) Latest() ([]*Snippet, error) {
	stmt := `SELECT id, title, content, created, expires FROM snippets
WHERE expires > current_date ORDER BY id DESC LIMIT 10`

	rows, err := m.DB.Query(context.Background(), stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	snippets := []*Snippet{}

	for rows.Next() {
		s := &Snippet{}

		err = rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
		if err != nil {
			return nil, err
		}

		snippets = append(snippets, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return snippets, nil
}
