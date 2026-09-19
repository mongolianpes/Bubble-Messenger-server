package db

const (
	IdsDir                         = "ids/%s"
	UsersDir                       = "users/%s"
	PathToUserName                 = "users/%s/name"
	PathToUserPassword             = "users/%s/password"
	PathToUserAvatar               = "users/%s/avatar"
	PathToUserMessagesDir          = "users/%s/messages"
	PathToUserBlockMessagesArchive = "users/%s/messages/%v"
	PathToUserNewMessagesFile      = "users/%s/messages/newMessages"
	PathToUserUnreceivedFilesDir   = "users/%s/files/"
	PathToServerRountineLog        = "log/serverRoutine.log"
	PathToRequestsLog              = "log/requests.log"
	PathToEchoErrLog               = "log/errEcho.log"
)

const (
	SecretServerSalt = "al2esad;famfopa14-,410-qu82304dfoaspddasdsdo934idsadasd342141das"
	ValuesAccessFile = 0600
	ValueAccessDir   = 0700
)
