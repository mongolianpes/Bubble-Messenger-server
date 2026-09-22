package handlers

import (
	"encoding/json"
	"net/http"
	"server/internal/crypto"

	"github.com/labstack/echo/v4"
)

func (h *Handler) SearchUser(c echo.Context) error {
	device := c.FormValue("device")
	loginForSearch := c.FormValue("loginforsearch")
	keyForServerDataBase := c.FormValue("forserver")
	if device == "" || loginForSearch == "" || keyForServerDataBase == "" {
		CountInvalidRequests += 1
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	key, err := h.UsersService.GetKey(c.Request().Context(), device)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	loginForSearch, err = crypto.StringDecrypt(loginForSearch, key)
	if err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("Decrypted error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
	}

	if !isValidStr(loginForSearch, false) {
		encryptResp, _ := crypto.StringEncrypt([]byte("Login for search is not valid"), key)
		return c.String(http.StatusBadRequest, encryptResp)
	}

	findUsers, err := h.UsersService.Search(c.Request().Context(), loginForSearch)
	if err != nil {
		encryptResp, _ := crypto.StringEncrypt([]byte("Can not find users"), key)
		return c.String(http.StatusBadRequest, encryptResp)
	}

	b, err := json.Marshal(findUsers)
	if err != nil {
		errRequestsLog.Printf("searchuse: Не удалось преобразовать в json найденных пользователей Device %s", device)
		encryptResp, _ := crypto.StringEncrypt([]byte("Can not marshal find users"), key)
		return c.String(http.StatusBadRequest, encryptResp)
	}

	b, _ = crypto.StringEncryptByte(b, key)

	usersRequestsLog.Printf("Запрос serchuser. LoginForSearch %s. DeviceID: %s", loginForSearch, device)
	return c.Blob(http.StatusOK, "application/octet-stream", b)
}
