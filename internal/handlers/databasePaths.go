package handlers

import "regexp"

const (
	devicesDir                     = "devices/%s"
	idsDir                         = "ids/%s"
	usersDir                       = "users/%s"
	pathToUserName                 = "users/%s/name"
	pathToUserPassword             = "users/%s/password"
	pathToUserAvatar               = "users/%s/avatar"
	pathToUserMessagesDir          = "users/%s/messages"
	pathToUserBlockMessagesArchive = "users/%s/messages/%v"
	pathToUserNewMessagesFile      = "users/%s/messages/newMessages"
	pathToUserUnreceivedFilesDir   = "users/%s/files/"
	pathToServerRountineLog        = "log/serverRoutine.log"
	pathToRequestsLog              = "log/requests.log"
	pathToEchoErrLog               = "log/errEcho.log"
)

const (
	secretServerSalt = "al2esad;famfopa14-,410-qu82304dfoaspddasdsdo934idsadasd342141das"
	valuesAccessFile = 0600
	valueAccessDir   = 0700
)

var regexpSymbols *regexp.Regexp = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func isValidStr(str string, isKey bool) bool {
	if !isKey && len(str) > 20 && len(str) < 6 {
		return false
	}

	return regexpSymbols.MatchString(str)
}
