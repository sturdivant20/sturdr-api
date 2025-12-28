package telemetry

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"math"
	"os"
	"strings"

	"github.com/sturdivant20/sturdr-api/include/navigation"
	"github.com/sturdivant20/sturdr-api/include/satellite"
)

type Service interface {
	createTelemetry(ctx context.Context, data Telemetry) error
	readTelemetry(ctx context.Context, week uint16, tow float32, do_query bool) ([]Telemetry, error)
	updateTelemetry(ctx context.Context, data Telemetry, sequence int64) error
	deleteTelemetry(ctx context.Context, sequence int64) error
}

type TelemetryService struct {
	db                *sql.DB
	CreateNavStmt     *sql.Stmt
	CreateSatStmt     *sql.Stmt
	ReadLatestNavStmt *sql.Stmt
	ReadQueryNavStmt  *sql.Stmt
	ReadLatestSatStmt *sql.Stmt
	ReadQuerySatStmt  *sql.Stmt
	UpdateNavStmt     *sql.Stmt
	UpdateSatStmt     *sql.Stmt
	DeleteNavStmt     *sql.Stmt
	DeleteSatStmt     *sql.Stmt
}

func NewTelemetryService(db *sql.DB, sql_fname string) Service {
	// --- read sql commands from file ---
	data, err := os.ReadFile(sql_fname)
	if err != nil {
		log.Fatalf("Error reading '%s': %s", sql_fname, err.Error())
	}
	cmd := strings.Split(string(data), ";")

	// --- bind statements ---
	return &TelemetryService{
		db:                db,
		CreateNavStmt:     bindStatement(db, strings.TrimSpace(cmd[0]), "create_telemetry_nav"),
		CreateSatStmt:     bindStatement(db, strings.TrimSpace(cmd[1]), "create_telemetry_sv"),
		ReadLatestNavStmt: bindStatement(db, strings.TrimSpace(cmd[2]), "read_latest_telemetry_nav"),
		ReadQueryNavStmt:  bindStatement(db, strings.TrimSpace(cmd[3]), "read_latest_telemetry_sv"),
		ReadLatestSatStmt: bindStatement(db, strings.TrimSpace(cmd[4]), "read_queried_telemetry_nav"),
		ReadQuerySatStmt:  bindStatement(db, strings.TrimSpace(cmd[5]), "read_queried_telemetry_sv"),
		UpdateNavStmt:     bindStatement(db, strings.TrimSpace(cmd[6]), "update_telemetry_nav"),
		UpdateSatStmt:     bindStatement(db, strings.TrimSpace(cmd[7]), "update_telemetry_sv"),
		DeleteNavStmt:     bindStatement(db, strings.TrimSpace(cmd[8]), "delete_telemetry_nav"),
		DeleteSatStmt:     bindStatement(db, strings.TrimSpace(cmd[9]), "delete_telemetry_sv")}
}

// Add a telemetry to the table
func (s *TelemetryService) createTelemetry(ctx context.Context, data Telemetry) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Create navigation post
	stmt := tx.StmtContext(ctx, s.CreateNavStmt)
	if _, err = stmt.ExecContext(ctx, data.Navigation.Args()...); err != nil {
		return err
	}

	// 2. Create satellite posts
	stmt = tx.StmtContext(ctx, s.CreateSatStmt)
	for _, sv := range data.Satellites {
		sv.Sequence = data.Navigation.Sequence // ensure the same sequence number
		if _, err = stmt.ExecContext(ctx, sv.Args()...); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Read telemetry (navigation and satellites) from the table
func (s *TelemetryService) readTelemetry(
	ctx context.Context, week uint16, tow float32, do_query bool) ([]Telemetry, error) {

	var n_rows, s_rows *sql.Rows
	var n_err, s_err error
	var data []Telemetry

	if do_query {
		// query
		n_rows, n_err = s.ReadQueryNavStmt.QueryContext(ctx, week, tow)
		s_rows, s_err = s.ReadQuerySatStmt.QueryContext(ctx, week, tow)
		if n_err != nil || s_err != nil {
			return []Telemetry{}, errors.Join(n_err, s_err)
		}
		// there are multiple navigation points
		for n_rows.Next() {
			var n navigation.Navigation
			if err := s_rows.Scan(n.Args()...); err != nil {
				return []Telemetry{}, err
			}
			data = append(data, Telemetry{Navigation: n})
		}
		// there are multiple satellites
		i := 0
		for s_rows.Next() {
			var sv satellite.Satellite
			if err := s_rows.Scan(sv.Args()...); err != nil {
				return []Telemetry{}, err
			}
			if math.Abs(float64(sv.ToW-data[i].Navigation.ToW)) > 1e-8 {
				i++
			}
			data[i].Satellites = append(data[i].Satellites, sv)
		}
	} else {
		// latest
		data = append(data, Telemetry{})
		n_rows, n_err = s.ReadLatestNavStmt.QueryContext(ctx)
		s_rows, s_err = s.ReadLatestSatStmt.QueryContext(ctx)
		if n_err != nil || s_err != nil {
			return []Telemetry{}, errors.Join(n_err, s_err)
		}
		// there is only 1 navigation point
		if n_rows.Next() {
			if err := n_rows.Scan(data[0].Navigation.Args()...); err != nil {
				return []Telemetry{}, err
			}
		}
		// there are multiple satellites
		for s_rows.Next() {
			var sv satellite.Satellite
			if err := s_rows.Scan(sv.Args()...); err != nil {
				return []Telemetry{}, err
			}
			data[0].Satellites = append(data[0].Satellites, sv)
		}
	}

	return data, nil
}

// Update a telemetry from the table
func (s *TelemetryService) updateTelemetry(ctx context.Context, data Telemetry, sequence int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Update navigation post
	stmt := tx.StmtContext(ctx, s.UpdateNavStmt)
	if _, err = stmt.ExecContext(ctx, append(data.Navigation.Args(), sequence)...); err != nil {
		return err
	}

	// 2. Update satellite posts
	stmt = tx.StmtContext(ctx, s.UpdateSatStmt)
	for _, sv := range data.Satellites {
		if _, err = stmt.ExecContext(ctx, append(sv.Args(), sequence)...); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Delete a telemetry from the table
func (s *TelemetryService) deleteTelemetry(ctx context.Context, sequence int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Delete navigation post
	stmt := tx.StmtContext(ctx, s.DeleteNavStmt)
	if _, err = stmt.ExecContext(ctx, sequence); err != nil {
		return err
	}

	// 2. Delete satellite posts
	stmt = tx.StmtContext(ctx, s.DeleteSatStmt)
	if _, err = stmt.ExecContext(ctx, sequence); err != nil {
		return err
	}

	return tx.Commit()
}

// bindStatement
func bindStatement(db *sql.DB, cmd string, name string) *sql.Stmt {
	create_stmt, err := db.Prepare(cmd)
	if err != nil {
		log.Fatalf("Error preparing '%s' statement: %s", name, err.Error())
	}
	return create_stmt
}
