package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type pqDB struct {
	db *sql.DB
}

type pqDbInfo struct {
	host string
	port uint16
	user string
	pass string
	name string
}

type execSQLData struct {
	res sql.Result
	err error
}

func (p *pqDB) execSQL(ctx context.Context, s string) (sql.Result, error) {
	c := make(chan execSQLData)

	go func(c chan execSQLData, s string) {
		res, err := p.db.Exec(s)
		c <- execSQLData{res, err}
	}(c, s)

	select {
	case <-ctx.Done():
		return nil, ErrContextCancelled
	case r := <-c:
		return r.res, r.err
	}
}

func (p *pqDB) initDB(ctx context.Context, dbName string) error {
	statement := fmt.Sprintf("SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = '%s');", dbName)

	row := p.db.QueryRow(statement)
	var exists bool
	err := row.Scan(&exists)
	if err != nil {
		return err
	}

	if exists == false {
		statement = fmt.Sprintf("CREATE DATABASE %s;", dbName)
		_, err = p.execSQL(ctx, statement)
	}

	return err
}

func (p *pqDB) initVisitsTable(ctx context.Context) error {
	statement := `CREATE TABLE IF NOT EXISTS
	visits(
		id SERIAL PRIMARY KEY NOT NULL,
		userId TEXT,
		name TEXT
	)`
	_, err := p.execSQL(ctx, statement)
	return err
}

func (p *pqDB) enableTrigrams(ctx context.Context) error {
	statement := "CREATE EXTENSION pg_trgm;"
	_, err := p.execSQL(ctx, statement)
	return err
}

func (p *pqDB) Close(ctx context.Context) {
	p.db.Close()
}

func (p *pqDB) ping(ctx context.Context, timeout, interval time.Duration) error {

	i := time.NewTicker(interval)
	o := time.NewTimer(timeout)

	var err error
	if err = p.db.PingContext(ctx); err == nil {
		return nil
	}

	for {
		select {
		case <-i.C:
			if err = p.db.PingContext(ctx); err == nil {
				return nil
			}
		case <-o.C:
			return err
		}
	}
}

// NewPqDB creates a new pqDB object and makes sure it is connected,
// otherwise returns an error
func NewPqDB(ctx context.Context, info *pqDbInfo, pingTimeout, pingInterval time.Duration) (*pqDB, error) {
	p := &pqDB{}

	dbinfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		info.host, info.port, info.user, info.pass, info.name)

	var err error
	p.db, err = sql.Open("postgres", dbinfo)
	if err != nil {
		return nil, err
	}

	if err := p.ping(ctx, pingTimeout, pingInterval); err != nil {
		return nil, err
	}

	if err := p.initDB(ctx, info.name); err != nil {
		return nil, err
	}

	if err := p.initVisitsTable(ctx); err != nil {
		return nil, err
	}

	// ignore errors for now
	p.enableTrigrams(ctx)

	return p, nil
}

type writeHistoryData struct {
	lastInsertID int
	err          error
}

// WriteHistory writes a new record to the database
func (p *pqDB) WriteHistory(ctx context.Context, e uLoc) (int, error) {

	c := make(chan writeHistoryData)

	go func(e uLoc, c chan writeHistoryData) {
		var r writeHistoryData
		r.err = p.db.QueryRow("INSERT INTO visits(userId,name) VALUES($1,$2) returning id;", e.UserID, e.Name).Scan(&r.lastInsertID)
		c <- r
		fmt.Println("last inserted id =", r.lastInsertID)
	}(e, c)

	select {
	case <-ctx.Done():
		return -1, ErrContextCancelled
	case r := <-c:
		return r.lastInsertID, r.err
	}
}

// GetHistoryByVisitID takes a visitId and returns the record found
// GetHistoryByVisitID returns an empty slice when no records are found
func (p *pqDB) GetHistoryByVisitID(ctx context.Context, visit string) ([]uLocVisit, error) {

	visitID, err := internalID(visit)
	if err != nil {
		return nil, err
	}

	rows, err := p.db.Query("SELECT * FROM visits WHERE id=$1", visitID)
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		return []uLocVisit{}, nil
	}

	var vid int
	var userID string
	var name string

	select {
	case <-ctx.Done():
		return nil, ErrContextCancelled
	default:
		err = rows.Scan(&vid, &userID, &name)
		if err != nil {
			return nil, err
		}
		// fmt.Println("uid | userId | name")
		// fmt.Printf("%3v | %8v | %6v\n", vid, userID, name)
	}
	return []uLocVisit{uLocVisit{uLoc{userID, name}, externalID(vid)}}, nil
}

// GetHistoryByUserID takes a context.Context because it may need to cancel searching through many records
// GetHistoryByUserID returns an empty slice when no records are found
func (p *pqDB) GetHistoryByUserID(ctx context.Context, userID, searchString string) ([]uLocVisit, error) {

	rows, err := p.db.Query("SELECT * FROM visits WHERE userId=$1 AND name % ANY(STRING_TO_ARRAY($2,' ')) ORDER BY id DESC LIMIT 5", userID, searchString)
	if err != nil {
		return nil, err
	}

	r := []uLocVisit{}

	for rows.Next() {
		select {
		case <-ctx.Done():
			return nil, ErrContextCancelled
		default:
			var vid int
			var userID string
			var name string

			err = rows.Scan(&vid, &userID, &name)
			if err != nil {
				return nil, err
			}
			// fmt.Println("vid | userId | name")
			// fmt.Printf("%3v | %8v | %6v\n", vid, userID, name)
			r = append(r, uLocVisit{uLoc{userID, name}, externalID(vid)})
		}
	}
	return r, nil
}
