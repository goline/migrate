package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mattes/migrate"
	_ "github.com/mattes/migrate/database/postgres"
	_ "github.com/mattes/migrate/source/file"
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

	Must(
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
	m, err := getMigrateInstance()
	PanicOnError(err)

	PanicOnError(m.Up())
	fmt.Println("Upgrade completed!")
}

func migrateDown() {
	fmt.Println("Downgrade migrations ...")
	m, err := getMigrateInstance()
	PanicOnError(err)

	PanicOnError(m.Down())
	fmt.Println("Downgrade completed!")
}

func preloadConfig() {
	utils.NewIniLoader().Load(configFile, conf)
}

func getMigrateInstance() (*migrate.Migrate, error) {
	return migrate.New(
		fmt.Sprintf("file://%s", migrationDir),
		conf.DbAddress,
	)
}

func PanicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func Must(errors ...error) {
	for _, err := range errors {
		PanicOnError(err)
	}
}
