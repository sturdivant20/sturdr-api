package satellite

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strings"
)

type Service interface {
	createSatellite(ctx context.Context, sv Satellite) error
	readSatellite(ctx context.Context, week uint16, tow float32, prn uint8, do_query bool) ([]Satellite, error)
	updateSatellite(ctx context.Context, sv Satellite, id int64) error
	deleteSatellite(ctx context.Context, id int64) error
}

type SatelliteService struct {
	db                     *sql.DB
	CreateStmt             *sql.Stmt
	ReadLatestStmt         *sql.Stmt
	ReadQueryStmt          *sql.Stmt
	ReadLatestSpecificStmt *sql.Stmt
	ReadQuerySpecificStmt  *sql.Stmt
	UpdateStmt             *sql.Stmt
	DeleteStmt             *sql.Stmt
}

// Initialize/create local satellite database file
func NewSatelliteService(db *sql.DB, sql_fname string) Service {
	// --- read sql commands from file ---
	data, err := os.ReadFile(sql_fname)
	if err != nil {
		log.Fatalf("Error reading '%s': %s", sql_fname, err.Error())
	}
	cmd := strings.Split(string(data), ";")

	// --- create table ---
	statement := bindStatement(db, strings.TrimSpace(cmd[0]), "make_satellite_table")
	statement.Exec()
	statement = bindStatement(db, cmd[1], "create_satellite_index")
	statement.Exec()
	log.Printf("Created 'satellite' database table ...")

	// --- bind statements ---
	return &SatelliteService{
		db:                     db,
		CreateStmt:             bindStatement(db, strings.TrimSpace(cmd[2]), "create_satellite"),
		ReadLatestStmt:         bindStatement(db, strings.TrimSpace(cmd[3]), "read_latest_satellite"),
		ReadQueryStmt:          bindStatement(db, strings.TrimSpace(cmd[4]), "read_queried_satellite"),
		ReadLatestSpecificStmt: bindStatement(db, strings.TrimSpace(cmd[5]), "read_latest_specific_satellite"),
		ReadQuerySpecificStmt:  bindStatement(db, strings.TrimSpace(cmd[6]), "read_queried_specific_satellite"),
		UpdateStmt:             bindStatement(db, strings.TrimSpace(cmd[7]), "update_satellite"),
		DeleteStmt:             bindStatement(db, strings.TrimSpace(cmd[8]), "delete_satellite")}
}

// Add a Satellite to the table
func (s *SatelliteService) createSatellite(ctx context.Context, sv Satellite) error {
	_, err := s.CreateStmt.ExecContext(ctx, sv.Args()...)
	return err
}

// Read Satellite from the table
func (s *SatelliteService) readSatellite(
	ctx context.Context, week uint16, tow float32, prn uint8, do_query bool) ([]Satellite, error) {

	var rows *sql.Rows
	var err error
	if do_query {
		// query specific time range
		if prn != 255 {
			// valid prn
			rows, err = s.ReadQuerySpecificStmt.QueryContext(ctx, week, tow, prn)
		} else {
			// invalid/all prn
			rows, err = s.ReadQueryStmt.QueryContext(ctx, week, tow)
		}
	} else {
		// query latest
		if prn != 255 {
			// valid prn
			rows, err = s.ReadLatestSpecificStmt.QueryContext(ctx, prn)
		} else {
			// invalid prn
			rows, err = s.ReadLatestStmt.QueryContext(ctx)
		}
	}

	// add satellites to list
	var items []Satellite
	for rows.Next() {
		var sv Satellite
		if err := rows.Scan(sv.Args()...); err != nil {
			return nil, err
		}
		items = append(items, sv)
	}

	return items, err
}

// Update a Satellite from the table
func (s *SatelliteService) updateSatellite(ctx context.Context, sv Satellite, id int64) error {
	_, err := s.UpdateStmt.ExecContext(ctx, append(sv.Args(), id)...)
	return err
}

// Delete a Satellite from the table
func (s *SatelliteService) deleteSatellite(ctx context.Context, id int64) error {
	_, err := s.DeleteStmt.ExecContext(ctx, id)
	return err
}

// bindStatement
func bindStatement(db *sql.DB, cmd string, name string) *sql.Stmt {
	create_stmt, err := db.Prepare(cmd)
	if err != nil {
		log.Fatalf("Error preparing '%s' statement: %s", name, err.Error())
	}
	return create_stmt
}
