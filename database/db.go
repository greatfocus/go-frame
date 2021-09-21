package database

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/greatfocus/gf-sframe/config"
)

// Conn struct
type Conn struct {
	master *db
	slave  *db
}

// db struct
type db struct {
	conn    *sql.DB
	timeout int64
}

// Init database connection for Master and Slave
func (c *Conn) Init(config *config.Config, impl *config.Impl) {
	var master = db{}
	master.connect(config.Database.Master, impl)
	var slave = db{}
	slave.connect(config.Database.Master, impl)
	c.master = &master
	c.slave = &slave
}

// Connect method make a database connection
func (d *db) connect(dbConfig config.DatabaseType, impl *config.Impl) {
	// initialize variables rom config
	log.Println("Preparing Database configuration")
	host := dbConfig.Host
	database := dbConfig.Database
	user := dbConfig.User
	password := dbConfig.Password
	sslmode := "disable"
	cert := dbConfig.Secure.Cert
	key := dbConfig.Secure.Key
	if dbConfig.Secure.SslMode {
		sslmode = "require"
	}
	port, err := strconv.ParseUint(dbConfig.Port, 0, 64)
	if err != nil {
		log.Fatal(fmt.Println(err))
	}
	maxLifetime := time.Duration(dbConfig.MaxLifetime) * time.Minute
	maxIdleConns := int(dbConfig.MaxIdleConns)
	maxOpenConns := int(dbConfig.MaxOpenConns)

	// create database connection
	psqlInfo := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s sslcert=%s sslkey=%s",
		host, port, user, password, database, sslmode, cert, key)
	conn, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(fmt.Println(err))
	}
	conn.SetConnMaxLifetime(maxLifetime)
	conn.SetMaxIdleConns(maxIdleConns)
	conn.SetMaxOpenConns(maxOpenConns)
	log.Println("Initiating Database connection")

	// execute database scripts
	if dbConfig.ExecuteSchema {
		d.executeSchema(conn)
		d.RebuildIndexes(conn, dbConfig.Database)
	}
	d.conn = conn
}

// ExecuteSchema prepare and execute database changes
func (d *db) executeSchema(db *sql.DB) {
	// read the scripts in the folder
	var path = os.Getenv("DATABASE_PATH") + "/"
	path = filepath.Clean(path)
	log.Println("Preparing to execute database schema")
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(fmt.Println(err))
	}

	// loop thru files to create schemas
	for _, f := range files {
		filepath := filepath.Clean(path + f.Name())
		scriptFile, err := os.OpenFile(filepath, os.O_RDONLY, 0600)
		if err != nil {
			log.Fatal(fmt.Println(err))
		}
		// read the config file
		scriptContent, err := ioutil.ReadAll(scriptFile)
		if err != nil {
			log.Fatal(fmt.Println(err))
		}
		sql := string(scriptContent)
		log.Println("Executing schema: ", path+f.Name())
		if _, err := db.Exec(sql); err != nil {
			log.Fatal(fmt.Println(err))
		}
	}

	log.Println("Database scripts successfully executed")
}

// RebuildIndexes within sframe
func (d *db) RebuildIndexes(db *sql.DB, dbname string) {
	log.Println("Rebuild Indexes")

	// Rebuild Indexes
	sqlReindexScript := string("REINDEX DATABASE " + dbname + ";")
	if _, err := db.Exec(sqlReindexScript); err != nil {
		log.Fatal(fmt.Println(err))
	}

	log.Println("Rebuild Indexes successfully executed")
}

// Insert method make a single row query to the slave databases
func (c *Conn) Insert(ctx context.Context, query string, args ...interface{}) *sql.Row {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(c.master.timeout)*time.Second)
	defer cancel()
	return c.master.conn.QueryRowContext(ctx, query, args...)
}

// Query method make a resultset rows query to the slave databases
func (c *Conn) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(c.slave.timeout)*time.Second)
	defer cancel()
	return c.slave.conn.QueryContext(ctx, query, args...)
}

// Select method make a single row query to the slave databases
func (c *Conn) Select(ctx context.Context, query string, args ...interface{}) *sql.Row {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(c.slave.timeout)*time.Second)
	defer cancel()
	return c.slave.conn.QueryRowContext(ctx, query, args...)
}

// Update method executes update database changes to the master databases
func (c *Conn) Update(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(c.master.timeout)*time.Second)
	defer cancel()
	return c.master.conn.ExecContext(ctx, query, args...)
}

// Delete method executes delete database changes to the master databases
func (c *Conn) Delete(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(c.master.timeout)*time.Second)
	defer cancel()
	return c.master.conn.ExecContext(ctx, query, args...)
}
