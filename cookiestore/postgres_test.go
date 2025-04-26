package cookiestore_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/takumi3488/twocker/cookiestore"

	_ "github.com/lib/pq"
)

const (
	testTableName = "test_cookies"
	dbUser        = "testuser"
	dbPassword    = "testpassword"
	dbName        = "testdb"
)

// setupPostgresContainer sets up a PostgreSQL container for testing.
// It returns the container instance, a database connection, and an error.
func setupPostgresContainer(ctx context.Context) (*postgres.PostgresContainer, *sql.DB, error) {
	pgContainer, err := postgres.Run(ctx,
		"postgres:17-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		postgres.WithInitScripts(), // Make sure init scripts run properly
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithStartupTimeout(5*time.Minute),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	// Get connection string AFTER container is confirmed running
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		// Terminate container if we cannot get connection string
		terminateCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		log.Println("Error getting connection string, terminating container...")
		if err := pgContainer.Terminate(terminateCtx); err != nil {
			log.Printf("Failed to terminate container: %v", err)
		}
		return nil, nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		// Terminate container if DB open fails
		terminateCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		log.Println("Error opening database connection, terminating container...")
		if err := pgContainer.Terminate(terminateCtx); err != nil {
			log.Printf("Failed to terminate container: %v", err)
		}
		return nil, nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test the connection with a ping to ensure it's really ready
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	for i := 0; i < 5; i++ { // Try multiple times if needed
		err = db.PingContext(pingCtx)
		if err == nil {
			break
		}
		log.Printf("Database ping attempt %d failed: %v, retrying...", i+1, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		terminateCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		log.Println("Failed to ping database after multiple attempts, terminating container...")
		if err := pgContainer.Terminate(terminateCtx); err != nil {
			log.Printf("Failed to terminate container: %v", err)
		}
		return nil, nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Println("PostgreSQL container started and connected successfully.")
	return pgContainer, db, nil
}

func TestPostgresCookieStoreIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	pgContainer, db, err := setupPostgresContainer(ctx)
	require.NoError(t, err, "Setup: Failed to set up PostgreSQL container")

	defer func() {
		log.Println("Tearing down PostgreSQL container...")
		terminateCtx, terminateCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer terminateCancel()

		if db != nil {
			if err := db.Close(); err != nil {
				log.Printf("Teardown: Failed to close database connection: %v", err)
			} else {
				log.Println("Teardown: Database connection closed.")
			}
		}
		// Ensure container is not nil before terminating
		if pgContainer != nil {
			if err := pgContainer.Terminate(terminateCtx); err != nil {
				log.Printf("Teardown: Failed to terminate PostgreSQL container: %v", err)
			} else {
				log.Println("Teardown: PostgreSQL container terminated.")
			}
		}
	}()

	// Create the cookie store instance AFTER successful DB connection and table creation check
	store, err := cookiestore.NewPostgresCookieStore(db, testTableName)
	require.NoError(t, err, "Setup: Failed to create PostgresCookieStore")
	require.NotNil(t, store, "Setup: Cookie store should not be nil")

	t.Run("SetAndGetCookies_Basic", func(t *testing.T) {
		testURL, _ := url.Parse("https://sub.example.com/some/path?query=1")
		cookiesToSet := []*http.Cookie{
			{Name: "session-id", Value: "abc123xyz", Path: "/", Domain: "example.com", HttpOnly: true},
			{Name: "user_preference", Value: "theme=dark&lang=en", Path: "/some", Domain: "sub.example.com"},
		}

		// Set cookies
		store.SetCookies(testURL, cookiesToSet)

		// Get cookies for the exact URL (should match based on host)
		retrievedCookies := store.Cookies(testURL)
		require.NotNil(t, retrievedCookies, "Retrieved cookies should not be nil for host sub.example.com")
		compareCookieSlices(t, cookiesToSet, retrievedCookies)

		// Get cookies for the base domain URL (should also work because host matches)
		baseDomainURL, _ := url.Parse("https://example.com/")
		retrievedBaseCookies := store.Cookies(baseDomainURL)
		require.NotNil(t, retrievedBaseCookies, "Retrieved cookies should not be nil for host example.com")
		compareCookieSlices(t, cookiesToSet, retrievedBaseCookies)
	})

	t.Run("GetCookies_NotFound", func(t *testing.T) {
		testURL, _ := url.Parse("https://another-domain.org")
		retrievedCookies := store.Cookies(testURL)
		require.Nil(t, retrievedCookies, "Cookies for a host not previously set should be nil")
	})

	t.Run("SetCookies_Overwrite", func(t *testing.T) {
		testURL, _ := url.Parse("https://sub.example.com/another?q=2")
		originalCookies := store.Cookies(testURL)
		require.NotNil(t, originalCookies, "Should have cookies before overwrite")

		newCookies := []*http.Cookie{
			{Name: "session-id", Value: "new-session-value-456", Path: "/", Domain: "example.com"}, // Overwrite
			{Name: "tracker-status", Value: "opt-out", Path: "/", Domain: "sub.example.com"},       // New cookie
		}

		store.SetCookies(testURL, newCookies)
		retrievedCookies := store.Cookies(testURL)
		require.NotNil(t, retrievedCookies, "Cookies should not be nil after overwrite")
		require.Len(t, retrievedCookies, len(newCookies), "Should have the new number of cookies after overwrite")
		compareCookieSlices(t, newCookies, retrievedCookies)

		// Verify the old 'user_preference' cookie is gone
		foundOldPref := false
		for _, c := range retrievedCookies {
			if c.Name == "user_preference" {
				foundOldPref = true
				break
			}
		}
		require.False(t, foundOldPref, "Old 'user_preference' cookie should not exist after overwrite")
	})

	t.Run("SetCookies_EmptyURLHostname", func(t *testing.T) {
		invalidURL := &url.URL{Scheme: "http", Path: "/no-host"}
		cookiesToSet := []*http.Cookie{{Name: "should-not-be-set", Value: "value"}}
		store.SetCookies(invalidURL, cookiesToSet)
		validURL, _ := url.Parse("https://sub.example.com")
		retrievedExampleCookies := store.Cookies(validURL)
		require.NotNil(t, retrievedExampleCookies, "Cookies for valid host sub.example.com should still exist")
		require.Len(t, retrievedExampleCookies, 2, "Should still have 2 cookies for sub.example.com after invalid set attempt")
	})

	t.Run("Cookies_EmptyURLHostname", func(t *testing.T) {
		invalidURL := &url.URL{Scheme: "http", Path: "/no-host-retrieve"}
		retrievedCookies := store.Cookies(invalidURL)
		require.Nil(t, retrievedCookies, "Retrieving cookies for URL with empty hostname should return nil")
	})
}

// compareCookieSlices compares two slices of cookies. Order doesn't matter.
// More sophisticated comparison (Path, Domain, Expires etc.) can be added if needed.
func compareCookieSlices(t *testing.T, expected, actual []*http.Cookie) {
	t.Helper() // Marks this function as a test helper
	require.NotNil(t, expected, "compareCookieSlices: expected slice cannot be nil")
	require.NotNil(t, actual, "compareCookieSlices: actual slice cannot be nil")
	require.Len(t, actual, len(expected), "Number of cookies mismatch. Expected %d, got %d. Expected: %v, Actual: %v", len(expected), len(actual), expected, actual)

	// Use maps for easier comparison regardless of order
	expectedMap := make(map[string]*http.Cookie)
	for _, c := range expected {
		expectedMap[c.Name] = c
	}

	actualMap := make(map[string]*http.Cookie)
	for _, c := range actual {
		// Check for duplicate names in actual slice, which might indicate an issue
		if _, exists := actualMap[c.Name]; exists {
			require.Fail(t, "Duplicate cookie name found in actual slice", "Cookie name: %s", c.Name)
		}
		actualMap[c.Name] = c
	}

	for name, expCookie := range expectedMap {
		actCookie, ok := actualMap[name]
		require.True(t, ok, "Expected cookie '%s' not found in actual cookies. Actual map: %v", name, actualMap)

		// Compare relevant fields (add more as needed)
		require.Equal(t, expCookie.Value, actCookie.Value, "Value mismatch for cookie '%s'", name)
		require.Equal(t, expCookie.Path, actCookie.Path, "Path mismatch for cookie '%s'", name)
		require.Equal(t, expCookie.Domain, actCookie.Domain, "Domain mismatch for cookie '%s'", name)
		require.Equal(t, expCookie.HttpOnly, actCookie.HttpOnly, "HttpOnly mismatch for cookie '%s'", name)
	}
}
