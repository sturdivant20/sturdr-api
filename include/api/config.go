package api

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pelletier/go-toml/v2"
)

// Full settings
type Config struct {
	Server    ServerConfig   `toml:"server"`
	Database  DatabaseConfig `toml:"database"`
	Sql       SqlSettings    `toml:"sql"`
	Endpoints EndpointConfig `toml:"endpoints"`
}

// Server settings
type ServerConfig struct {
	Host string `toml:"host"`
	Port int    `toml:"port"`
}

// Database settings
type DatabaseConfig struct {
	DbFile  string `toml:"db_file"`
	MaxSize int    `toml:"max_size"`
	Clear   bool   `toml:"clear"`
}

// SQL settings
type SqlSettings struct {
	NavigationCmds string `toml:"navigation_cmds"`
	SatelliteCmds  string `toml:"satellite_cmds"`
	TelemetryCmds  string `toml:"telemetry_cmds"`
}

// Endpoint settings
type EndpointConfig struct {
	Gui        string `toml:"gui"`
	Navigation string `toml:"navigation"`
	Satellite  string `toml:"satellite"`
	Telemetry  string `toml:"telemetry"`
	Create     string `toml:"create"`
	Read       string `toml:"read"`
	Update     string `toml:"update"`
	Delete     string `toml:"delete"`
}

// ParseSettings
func parseSettings(filename string) (Config, error) {
	var cfg Config

	// read file
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Printf("Error reading TOML file: %s", err.Error())
		return Config{}, err
	}

	// parse toml data
	err = toml.Unmarshal(data, &cfg)
	if err != nil {
		log.Printf("Error unmarshaling TOML data: %s", err.Error())
		return Config{}, err
	}

	// print config
	log.Printf("\n[server]\n host = %s\n port = %d\n"+
		"\n[database]\n db_file = %s\n max_size = %d\n clear = %t\n"+
		"\n[sql]\n navigation_cmds = %s\n satellite_cmds = %s\n telemetry_cmds = %s\n"+
		"\n[endpoints]\n gui = %s\n navigation = %s\n satellite = %s\n telemetry = %s\n "+
		"create = %s\n read = %s\n update = %s\n delete = %s\n\n",
		cfg.Server.Host,
		cfg.Server.Port,
		cfg.Database.DbFile,
		cfg.Database.MaxSize,
		cfg.Database.Clear,
		cfg.Sql.NavigationCmds,
		cfg.Sql.SatelliteCmds,
		cfg.Sql.TelemetryCmds,
		cfg.Endpoints.Gui,
		cfg.Endpoints.Navigation,
		cfg.Endpoints.Satellite,
		cfg.Endpoints.Telemetry,
		cfg.Endpoints.Create,
		cfg.Endpoints.Read,
		cfg.Endpoints.Update,
		cfg.Endpoints.Delete)

	return cfg, nil
}

// InitDatabase
func initDatabase(db_file string, clear bool) (*sql.DB, error) {
	// --- initialize database file ---
	_, err := os.Stat(db_file)
	if !os.IsNotExist(err) || clear {
		log.Printf("Removing old '%s' database file ...", db_file)
		os.Remove(db_file)
		file, err := os.Create(db_file)
		if err != nil {
			log.Printf("Error creating '%s' database file: %s", db_file, err.Error())
			return nil, err
		}
		file.Close()
	}
	log.Printf("Initialized '%s' database file ...", db_file)

	// --- open database ---
	db, err := sql.Open("sqlite3", "file:"+db_file+"?_foreign_keys=on") // open with sqlite
	if err != nil {
		log.Printf("Error opening '%s' database file: %s", db_file, err.Error())
		return nil, err
	}
	log.Printf("Opened '%s' database file ...", db_file)

	return db, nil
}
