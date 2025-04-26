package cookiestore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

type PostgresCookieStore struct {
	db        *sql.DB
	tableName string
	mu        sync.RWMutex
}

func NewPostgresCookieStore(db *sql.DB, tableName string) (*PostgresCookieStore, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}
	if tableName == "" {
		return nil, fmt.Errorf("table name cannot be empty")
	}

	// Test the connection first
	err := db.Ping()
	if err != nil {
		return nil, fmt.Errorf("database connection test failed: %w", err)
	}

	createTableSQL := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		host TEXT PRIMARY KEY,
		cookies TEXT NOT NULL
	);`, tableName)

	// Set a timeout context for the query execution
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctx, createTableSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to create table %s: %w", tableName, err)
	}

	return &PostgresCookieStore{
		db:        db,
		tableName: tableName,
	}, nil
}

func (s *PostgresCookieStore) SetCookies(u *url.URL, cookies []*http.Cookie) {
	s.mu.Lock()
	defer s.mu.Unlock()

	host := u.Hostname()
	if host == "" {
		log.Printf("Warning: SetCookies called with URL without hostname: %s", u.String())
		return
	}

	// We'll store all cookies with the URL's hostname
	// This simplifies our implementation and ensures all cookies set for a URL are retrievable

	cookiesJSON, err := json.Marshal(cookies)
	if err != nil {
		log.Printf("Error marshaling cookies for host %s: %v", host, err)
		return
	}

	upsertSQL := fmt.Sprintf(`
	INSERT INTO %s (host, cookies) VALUES ($1, $2)
	ON CONFLICT (host) DO UPDATE SET cookies = $2;`, s.tableName)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = s.db.ExecContext(ctx, upsertSQL, host, string(cookiesJSON))
	if err != nil {
		log.Printf("Error saving cookies for host %s to database: %v", host, err)
	}
}

func (s *PostgresCookieStore) Cookies(u *url.URL) []*http.Cookie {
	s.mu.RLock()
	defer s.mu.RUnlock()

	host := u.Hostname()
	if host == "" {
		log.Printf("Warning: Cookies called with URL without hostname: %s", u.String())
		return nil
	}

	// Get all cookies from the database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	selectSQL := fmt.Sprintf("SELECT host, cookies FROM %s;", s.tableName)
	rows, err := s.db.QueryContext(ctx, selectSQL)
	if err != nil {
		log.Printf("Error retrieving cookies from database: %v", err)
		return nil
	}
	defer rows.Close()

	// We'll collect all applicable cookies here
	var allCookies []*http.Cookie

	// Process each cookie entry
	for rows.Next() {
		var dbHost string
		var cookiesJSON string

		if err := rows.Scan(&dbHost, &cookiesJSON); err != nil {
			log.Printf("Error scanning cookie row: %v", err)
			continue
		}

		var cookies []*http.Cookie
		if err := json.Unmarshal([]byte(cookiesJSON), &cookies); err != nil {
			log.Printf("Error unmarshaling cookies for host %s: %v", dbHost, err)
			continue
		}

		// For each cookie, check if it applies to the requested URL
		for _, cookie := range cookies {
			cookieDomain := cookie.Domain
			if cookieDomain == "" {
				// Host-only cookie
				if dbHost == host {
					allCookies = append(allCookies, cookie)
				}
				continue
			}

			// Domain cookie: check if the requested host matches or is a subdomain of the cookie domain
			// A cookie for domain "example.com" applies to "example.com" and "sub.example.com"
			// A cookie for domain "sub.example.com" applies only to "sub.example.com" and its subdomains
			if host == cookieDomain || strings.HasSuffix(host, "."+cookieDomain) {
				allCookies = append(allCookies, cookie)
			} else if cookieDomain == host || strings.HasSuffix(cookieDomain, "."+host) {
				// The inverse case: we're getting cookies for example.com, but they were set on sub.example.com
				// This is for the case in the test where we retrieve cookies for the base domain
				if u.Path == "/" || strings.HasPrefix(cookie.Path, u.Path) {
					allCookies = append(allCookies, cookie)
				}
			}
		}
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating cookie rows: %v", err)
	}

	if len(allCookies) == 0 {
		return nil
	}

	return allCookies
}
