package misc

import (
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

var (
	host, user, password, sslmode string
	port                          int
)

type ConfigDBUtils struct {
	Host     string
	User     string
	Port     int
	Password string
	Sslmode  string
}

func loadConfig(config ConfigDBUtils) {
	host = config.Host
	user = config.User
	port = config.Port
	password = config.Password // Reading secrets from
	sslmode = config.Sslmode
}

var DefaultConfigDBUtils = ConfigDBUtils{Host: "localhost", User: "ubuntu", Port: 5432, Password: "ubuntu", Sslmode: "disable"}

/*
ReplaceDB : Rename the OLD DB and create a new one.
Since we are not journaling, this should be idemponent
*/

func checkAndValidateConfig(configDBUtilsList ...interface{}) ConfigDBUtils {
	if len(configDBUtilsList) != 1 {
		return DefaultConfigDBUtils
	}
	switch configDBUtilsList[0].(type) {
	case ConfigDBUtils:
		return configDBUtilsList[0].(ConfigDBUtils)
	default:
		return DefaultConfigDBUtils
	}
}

func ReplaceDB(dbName, targetName string, configDBUtilsList ...interface{}) {
	config := checkAndValidateConfig(configDBUtilsList...)
	loadConfig(config)
	connInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=%s",
		host, port, user, password, sslmode)
	db, err := sql.Open("postgres", connInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	//Killing sessions on the db
	sqlStatement := fmt.Sprintf("SELECT pid, pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '%s' AND pid <> pg_backend_pid();", dbName)
	rows, err := db.Query(sqlStatement)
	if err != nil {
		panic(err)
	}
	rows.Close()

	renameDBStatement := fmt.Sprintf("ALTER DATABASE \"%s\" RENAME TO \"%s\"",
		dbName, targetName)
	pkgLogger.Debug(renameDBStatement)
	_, err = db.Exec(renameDBStatement)

	// If execution of ALTER returns error, pacicking
	if err != nil {
		panic(err)
	}

	createDBStatement := fmt.Sprintf("CREATE DATABASE \"%s\"", dbName)
	_, err = db.Exec(createDBStatement)
	if err != nil {
		panic(err)
	}
}

func QuoteLiteral(literal string) string {
	return pq.QuoteLiteral(literal)
}
