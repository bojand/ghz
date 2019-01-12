package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func TestInfoAPI(t *testing.T) {

	api := InfoAPI{
		Info: ApplicationInfo{
			Version:   "1.2.3",
			GOVersion: "1.11",
			BuildDate: "January 1, 2019",
			StartTime: time.Now(),
		}}

	t.Run("GetInfo", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/info", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)

		if assert.NoError(t, api.GetApplicationInfo(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)

			info := new(InfoResponse)
			err := json.NewDecoder(rec.Body).Decode(info)

			assert.NoError(t, err)

			assert.Equal(t, "1.2.3", info.Version)
			assert.Equal(t, "1.11", info.RuntimeVersion)
			assert.NotEmpty(t, "1.11", info.BuildDate)
		}
	})
}
