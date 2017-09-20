package main

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/goline/lapi"
	"github.com/goline/utils"
	_ "github.com/lib/pq"
	"github.com/mattes/migrate"
	"github.com/mattes/migrate/database"
	"github.com/mattes/migrate/database/postgres"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app = kingpin.New("migrate", "A command-line migration tool")

	makeCmd          = app.Command("make", "Make empty migration files")
	makeName         = makeCmd.Arg("name", "File's name to be created").Required().String()
	makeMigrationDir = makeCmd.Flag("dir", "Migration directory").Required().String()

	upCmd          = app.Command("up", "Run upgrade migration")
	upMigrationDir = upCmd.Flag("dir", "Migration directory").Required().String()
	upConfigFile   = upCmd.Flag("config", "Path to configuration file").Required().String()

	downCmd          = app.Command("down", "Run downgrade migration")
	downMigrationDir = downCmd.Flag("dir", "Migration directory").Required().String()
	downConfigFile   = downCmd.Flag("config", "Path to configuration file").Required().String()

	migrationDir = ""
	configFile   = ""
	conf         = new(config)
)

func main() {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case makeCmd.FullCommand():
		migrationDir = *makeMigrationDir
		makeMigrationFile()
	case upCmd.FullCommand():
		configFile = *upConfigFile
		migrationDir = *upMigrationDir
		preloadConfig()
		migrateUp()
	case downCmd.FullCommand():
		configFile = *downConfigFile
		migrationDir = *downMigrationDir
		preloadConfig()
		migrateDown()
	default:
		panic("Unknown action")
	}
}

func makeMigrationFile() {
	dir := fmt.Sprintf("%s", migrationDir)
	name := strings.Replace(*makeName, " ", "_", -1)
	now := time.Now().Unix()

	lapi.Must(
		createNewFile(fmt.Sprintf("%s/%d_%s.up.sql", dir, now, name)),
		createNewFile(fmt.Sprintf("%s/%d_%s.down.sql", dir, now, name)),
	)
}

func createNewFile(path string) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0755)
	defer f.Close()
	if err != nil {
		return err
	}
	fmt.Println(fmt.Sprintf("Created file: %s", path))
	return nil
}

func migrateUp() {
	fmt.Println("Upgrade migrations ...")
	driver, err := getDatabaseDriver()
	PanicOnError(err)

	m, err := getMigrateInstance(driver)
	PanicOnError(err)

	PanicOnError(m.Up())
	fmt.Println("Upgrade completed!")
}

func migrateDown() {
	fmt.Println("Downgrade migrations ...")
	driver, err := getDatabaseDriver()
	PanicOnError(err)

	m, err := getMigrateInstance(driver)
	PanicOnError(err)

	PanicOnError(m.Down())
	fmt.Println("Downgrade completed!")
}

func preloadConfig() {
	utils.NewIniLoader().Load(configFile, conf)
}

func getDatabaseDriver() (database.Driver, error) {
	db, err := sql.Open("postgres", conf.DbAddress)
	PanicOnError(err)
	return postgres.WithInstance(db, &postgres.Config{})
}

func getMigrateInstance(driver database.Driver) (*migrate.Migrate, error) {
	u, err := url.Parse(conf.DbAddress)
	if err != nil {
		return nil, err
	}
	name := strings.Replace(u.Path, "/", "", 1)
	return migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file:///%s", migrationDir),
		name, driver,
	)
}

func PanicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
