package search

import (
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"server/internal/db"
	"strings"
	"time"

	"server/internal/models"
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

func SortUsersInSearch(findLogins *[]models.FindUser, loginForSearch string) {
	searchString := strings.ToLower(loginForSearch)

	limitSearch := 3 + len(loginForSearch)
	for userLogin := range popularTodayUsers {
		userLoginLower := strings.ToLower(userLogin)
		if strings.Contains(userLoginLower, searchString) && userLoginLower != searchString {
			findUser := models.FindUser{
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
			findUser := models.FindUser{
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

	if userName, err := os.ReadFile(fmt.Sprintf(db.PathToUserName, loginForSearch)); !os.IsNotExist(err) {
		findUser := models.FindUser{
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

func SavePopularUsers() error {
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
		return errors.New("Не удалось записать популярных пользователей")
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	if err = encoder.Encode(popularTodayUsers); err != nil {
		return errors.New("Не удалось закодировать новых популярных пользователей")
	}

	return nil
}

func LoadPopularUsers() error {
	file, err := os.Open("popularAlwaysUsers")
	if err != nil {
		return errors.New("Не удалось открыть файл с популярными пользователями")
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	if err = decoder.Decode(&popularAlwaysUsers); err != nil {
		return errors.New("Не удалось декодировать популярных пользователей")
	}

	return nil
}
