package handlers

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"server/internal/crypto"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	limitPopularTodayUsers         = 100
	numIncreasedRemovePopularUsers = 75
)

var popularTodayUsers = make(map[string]*PopularUser)
var popularAlwaysUsers = make(map[string]*PopularUser)

type PopularUser struct {
	UserName     string
	IndexPopular int
	LastFindTime time.Time
}

type FindUser struct {
	Login string `json:"login"`
	Name  string `json:"name"`
}

func SearchUserRequest(c echo.Context) error {
	device := c.FormValue("device")
	loginForSearch := c.FormValue("loginforsearch")
	keyForServerDataBase := c.FormValue("forserver")
	if device == "" || loginForSearch == "" || keyForServerDataBase == "" {
		CountInvalidRequests += 1
		return c.String(http.StatusBadRequest, "Did not receive all server data")
	}

	fileData, err := os.ReadFile(fmt.Sprintf(idsDir, device[:220]))
	if err != nil {
		usersRequestsLog.Printf("Попытка отправки запроса от незарегистрированного устройства LoginForSearch %s, Device %s", loginForSearch, device)
		return c.String(http.StatusBadRequest, "This device is not registered")
	}
	key, err := crypto.StringDecrypt(string(fileData), keyForServerDataBase+secretServerSalt)
	if err != nil {
		usersRequestsLog.Printf("Попытка отправки запроса от незарегистрированного устройства LoginForSearch %s, Device %s", loginForSearch, device)
		encryptResp, _ := crypto.StringEncrypt([]byte("Unknown error"), key)
		return c.String(http.StatusInternalServerError, encryptResp)
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

	var findUsers []FindUser
	sortUsersInSearch(&findUsers, loginForSearch)

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

func sortUsersInSearch(findLogins *[]FindUser, loginForSearch string) {
	searchString := strings.ToLower(loginForSearch)

	limitSearch := 3 + len(loginForSearch)
	for userLogin := range popularTodayUsers {
		userLoginLower := strings.ToLower(userLogin)
		if strings.Contains(userLoginLower, searchString) && userLoginLower != searchString {
			findUser := FindUser{
				Login: userLogin,
				Name:  popularTodayUsers[userLogin].UserName,
			}
			*findLogins = append(*findLogins, findUser)
		} else {
			var timeAfterRemove time.Duration
			if len(popularTodayUsers) > numIncreasedRemovePopularUsers {
				timeAfterRemove = time.Hour
			} else {
				timeAfterRemove = 5 * time.Hour
			}

			if popularTodayUsers[userLogin].IndexPopular == 1 && time.Since(popularTodayUsers[userLogin].LastFindTime) > timeAfterRemove {
				delete(popularTodayUsers, userLogin)
			}

			limitSearch--
			if limitSearch == 0 {
				break
			}
		}
	}

	limitSearch = 3 + len(loginForSearch)
	for userLogin := range popularAlwaysUsers {
		userLoginLower := strings.ToLower(userLogin)
		if strings.Contains(userLoginLower, searchString) && userLoginLower != searchString {
			findUser := FindUser{
				Login: userLogin,
				Name:  popularAlwaysUsers[userLogin].UserName,
			}
			*findLogins = append(*findLogins, findUser)
		} else {
			var timeAfterRemove time.Duration
			if len(popularAlwaysUsers) > numIncreasedRemovePopularUsers {
				timeAfterRemove = time.Hour
			} else {
				timeAfterRemove = 5 * time.Hour
			}

			if popularAlwaysUsers[userLogin].IndexPopular == 1 && time.Since(popularAlwaysUsers[userLogin].LastFindTime) > timeAfterRemove {
				delete(popularAlwaysUsers, userLogin)
			}

			limitSearch--
			if limitSearch == 0 {
				break
			}
		}
	}

	if userName, err := os.ReadFile(fmt.Sprintf(pathToUserName, loginForSearch)); !os.IsNotExist(err) {
		findUser := FindUser{
			Login: loginForSearch,
			Name:  string(userName),
		}
		*findLogins = append(*findLogins, findUser)

		if len(popularTodayUsers) < limitPopularTodayUsers {
			popularTodayUser, ok := popularTodayUsers[loginForSearch]
			if !ok {
				popularTodayUser = &PopularUser{}
				popularTodayUsers[loginForSearch] = popularTodayUser

				popularTodayUsers[loginForSearch].IndexPopular = 1
				popularTodayUsers[loginForSearch].LastFindTime = time.Now()
				popularTodayUsers[loginForSearch].UserName = string(userName)
			} else {
				if time.Since(popularTodayUser.LastFindTime) > 20*time.Minute {
					popularTodayUsers[loginForSearch].IndexPopular++
					popularTodayUsers[loginForSearch].LastFindTime = time.Now()
				}
			}
		}
	}
}

func SavePopularUsers() {
	var mostPopularToday *PopularUser
	haveUserToChange := false
	for userLogin, user := range popularTodayUsers {
		if _, ok := popularAlwaysUsers[userLogin]; ok {
			popularAlwaysUsers[userLogin].LastFindTime = popularTodayUsers[userLogin].LastFindTime

			if popularAlwaysUsers[userLogin].IndexPopular > mostPopularToday.IndexPopular {
				mostPopularToday = user
				haveUserToChange = true
			}
		}
	}

	for userLogin, user := range popularAlwaysUsers {
		sinceTime := time.Since(user.LastFindTime)
		if sinceTime > time.Hour*48 {
			user.IndexPopular--
		}

		if sinceTime < time.Hour*24 {
			user.IndexPopular++
		}

		if haveUserToChange {
			popularAlwaysUsers[userLogin] = mostPopularToday
			haveUserToChange = false
		}
	}

	file, err := os.Create("popularAlwaysUsers")
	if err != nil {
		errServerRoutineLog.Print("Не удалось записать популярных пользователей")
		return
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	if err = encoder.Encode(popularTodayUsers); err != nil {
		errServerRoutineLog.Print("Не удалось закодировать новых популярных пользователей")
		return
	}

	ServerRoutineLog.Print("Успешное обновление популярных пользователей")
}

func LoadPopularUsers() {
	file, err := os.Open("popularAlwaysUsers")
	if err != nil {
		errServerRoutineLog.Print("Не удалось открыть файл с популярными пользователями")
		return
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	if err = decoder.Decode(&popularAlwaysUsers); err != nil {
		errServerRoutineLog.Print("Не удалось декодировать популярных пользователей")
		return
	}

	ServerRoutineLog.Print("Успешное чтение популярных пользователей")
}
