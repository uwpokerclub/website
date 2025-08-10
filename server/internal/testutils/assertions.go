package testutils

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SetAuthCookie sets the authentication cookie for testing
func SetAuthCookie(req *http.Request, sessionID uuid.UUID) {
	cookieName := "uwpsc-dev-session-id"
	if os.Getenv("ENVIRONMENT") == "production" {
		cookieName = "uwpsc-session-id"
	}

	cookie := &http.Cookie{
		Name:  cookieName,
		Value: sessionID.String(),
	}
	req.AddCookie(cookie)
}

// TestUnauthenticatedEndpoint tests that an endpoint requires authentication
func TestUnauthenticatedEndpoint(t *testing.T, container *PostgresTestContainer, apiServer *gin.Engine, method, endpoint string, requestBody ...map[string]interface{}) {
	// Reset database for clean state
	ctx := context.Background()
	require.NoError(t, container.ResetDatabase(ctx))

	// Create request without authentication (use empty body if none provided)
	var body map[string]interface{}
	if len(requestBody) > 0 {
		body = requestBody[0]
	}
	req, err := MakeJSONRequest(method, endpoint, body)
	require.NoError(t, err)

	// Execute request
	w := httptest.NewRecorder()
	apiServer.ServeHTTP(w, req)

	// Assert unauthenticated response
	AssertUnauthenticatedResponse(t, w)
}

// TestUnauthorizedEndpoint tests that an endpoint properly checks authorization for different roles
func TestUnauthorizedEndpoint(t *testing.T, container *PostgresTestContainer, apiServer *gin.Engine, method, endpoint string, unauthorizedRoles []string, requestBody ...map[string]interface{}) {
	for _, role := range unauthorizedRoles {
		t.Run("unauthorized_role_"+role, func(t *testing.T) {
			// Reset database for clean state
			ctx := context.Background()
			require.NoError(t, container.ResetDatabase(ctx))
			db := container.GetDB()

			// Create session with unauthorized role
			sessionID, err := CreateTestSession(db, "testuser", role)
			require.NoError(t, err)

			// Create request with authentication but unauthorized role (use empty body if none provided)
			var body map[string]interface{}
			if len(requestBody) > 0 {
				body = requestBody[0]
			}
			req, err := MakeJSONRequest(method, endpoint, body)
			require.NoError(t, err)
			SetAuthCookie(req, sessionID)

			// Execute request
			w := httptest.NewRecorder()
			apiServer.ServeHTTP(w, req)

			// Assert unauthorized response
			AssertUnauthorizedResponse(t, w)
		})
	}
}

// TestInvalidAuthForEndpoint combines both unauthenticated and unauthorized tests
func TestInvalidAuthForEndpoint(t *testing.T, container *PostgresTestContainer, apiServer *gin.Engine, method, endpoint string, unauthorizedRoles []string, requestBody ...map[string]interface{}) {
	t.Run("unauthenticated", func(t *testing.T) {
		TestUnauthenticatedEndpoint(t, container, apiServer, method, endpoint, requestBody...)
	})

	t.Run("unauthorized", func(t *testing.T) {
		TestUnauthorizedEndpoint(t, container, apiServer, method, endpoint, unauthorizedRoles, requestBody...)
	})
}

// AssertErrorResponse tests that error responses have the expected status and message
func AssertErrorResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedMessage string) {
	assert.Equal(t, expectedStatus, w.Code)
	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	assert.Equal(t, expectedMessage, response["message"])
}

// AssertUnauthenticatedResponse tests that unauthenticated requests return 401
func AssertUnauthenticatedResponse(t *testing.T, w *httptest.ResponseRecorder) {
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	assert.Equal(t, "UNAUTHORIZED", response["type"])
	assert.Equal(t, "Authentication required", response["message"])
}

// AssertUnauthorizedResponse tests that unauthorized requests return 403
func AssertUnauthorizedResponse(t *testing.T, w *httptest.ResponseRecorder) {
	assert.Equal(t, http.StatusForbidden, w.Code)
	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	assert.Equal(t, "FORBIDDEN", response["type"])
	assert.Equal(t, "You do not have permission to perform this action.", response["message"])
}

// AssertBadRequestResponse tests that bad requests return 400
func AssertBadRequestResponse(t *testing.T, w *httptest.ResponseRecorder, expectedFieldError string) {
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	assert.Equal(t, "INVALID_REQUEST", response["type"])
	if expectedFieldError != "" {
		assert.Contains(t, response["message"], expectedFieldError)
	}
}

// AssertSuccessResponse tests that successful requests have the expected status and response structure
func AssertSuccessResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedResponse interface{}) {
	assert.Equal(t, expectedStatus, w.Code)

	// Convert expectedResponse to JSON string if it's not already
	var expectedJSON string

	switch v := expectedResponse.(type) {
	case string:
		expectedJSON = v
	case []byte:
		expectedJSON = string(v)
	default:
		expectedBytes, err := json.Marshal(v)
		require.NoError(t, err)
		expectedJSON = string(expectedBytes)
	}

	// Use assert.JSONEq to compare JSON strings
	assert.JSONEq(t, expectedJSON, w.Body.String())
}

// MakeJSONRequest creates an HTTP request with JSON body
func MakeJSONRequest(method, url string, body map[string]interface{}) (*http.Request, error) {
	var requestBody []byte
	if body != nil {
		var err error
		requestBody, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}
