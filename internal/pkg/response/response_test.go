package response_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mzhryns/titik-nol-backend/internal/pkg/response"
	"github.com/stretchr/testify/assert"
)

func TestSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test", nil)

	data := map[string]string{"key": "value"}
	response.Success(c, http.StatusOK, "Success", data)

	assert.Equal(t, http.StatusOK, w.Code)

	var res response.Response
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "Success", res.Message)
	
	// Check data
	dataMap := res.Data.(map[string]interface{})
	assert.Equal(t, "value", dataMap["key"])
}

func TestError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/test-error", nil)

	response.Error(c, http.StatusBadRequest, "Bad Request", "Detail error", []response.FieldFailure{
		{Field: "email", Message: "invalid"},
	})

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var res response.Response
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Bad Request", res.Message)
	assert.NotNil(t, res.Error)
	assert.Equal(t, 400, res.Error.Status)
	assert.Equal(t, "Detail error", res.Error.Detail)
	assert.Equal(t, "/test-error", res.Error.Instance)
	assert.Len(t, res.Error.Errors, 1)
	assert.Equal(t, "email", res.Error.Errors[0].Field)
}
