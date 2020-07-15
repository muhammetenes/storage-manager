package awss3

import (
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogin(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	//c.SetPath("/login")
	//c.SetParamNames("email")
	//c.SetParamValues("jon@labstack.com")
	//if assert.NoError(t, h.getUser(c)) {
	//	assert.Equal(t, http.StatusOK, rec.Code)
	//	assert.Equal(t, "", rec.Body.String())
	//}
}
