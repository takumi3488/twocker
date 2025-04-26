package cookiestore_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

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
