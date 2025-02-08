package _skeleton

var migrations = []string{
	`CREATE TABLE IF NOT EXISTS Migration (
		MigrationID INTEGER PRIMARY KEY,
		Md5Sum TEXT,
		UnixTimestamp INTEGER
	) STRICT;
    `,
	`CREATE TABLE IF NOT EXISTS User (
		UserID INTEGER PRIMARY KEY,
		Username TEXT NOT NULL
	) STRICT`,
	`CREATE TABLE IF NOT EXISTS Conversation (
		ConversationID INTEGER PRIMARY KEY,
		FirstUserID INTEGER REFERENCES User (UserID) NOT NULL,
		SecondUserID INTEGER REFERENCES User (UserID) NOT NULL
	) STRICT`,
	`CREATE TABLE IF NOT EXISTS Message(
		MessageID INTEGER PRIMARY KEY,
		ConversationID INTEGER REFERENCES Conversation (ConversationID) NOT NULL,
		Message TEXT NOT NULL
	) STRICT`,
}
