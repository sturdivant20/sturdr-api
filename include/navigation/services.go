package navigation

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strings"
)

type Service interface {
	createNavigation(ctx context.Context, n Navigation) error
	readNavigation(ctx context.Context, week uint16, tow float32, do_query bool) ([]Navigation, error)
	updateNavigation(ctx context.Context, n Navigation, id int64) error
	deleteNavigation(ctx context.Context, id int64) error
}

type NavigationService struct {
	db             *sql.DB
	CreateStmt     *sql.Stmt
	ReadLatestStmt *sql.Stmt
	ReadQueryStmt  *sql.Stmt
	UpdateStmt     *sql.Stmt
	DeleteStmt     *sql.Stmt
}

// Initialize/create local navigation database file
func NewNavigationService(db *sql.DB, sql_fname string) Service {
	// --- read sql commands from file ---
	data, err := os.ReadFile(sql_fname)
	if err != nil {
		log.Fatalf("Error reading '%s': %s", sql_fname, err.Error())
	}
	cmd := strings.Split(string(data), ";")

	// --- create table ---
	statement := bindStatement(db, strings.TrimSpace(cmd[0]), "make_navigation_table")
	statement.Exec()
	statement = bindStatement(db, cmd[1], "create_navigation_index")
	statement.Exec()
	log.Printf("Created 'navigation' database table ...")

	// --- bind statements ---
	return &NavigationService{
		db:             db,
		CreateStmt:     bindStatement(db, strings.TrimSpace(cmd[2]), "create_navigation"),
		ReadLatestStmt: bindStatement(db, strings.TrimSpace(cmd[3]), "read_latest_navigation"),
		ReadQueryStmt:  bindStatement(db, strings.TrimSpace(cmd[4]), "read_queried_navigation"),
		UpdateStmt:     bindStatement(db, strings.TrimSpace(cmd[5]), "update_navigation"),
		DeleteStmt:     bindStatement(db, strings.TrimSpace(cmd[6]), "delete_navigation")}
}

// Add a navigation to the table
func (s *NavigationService) createNavigation(ctx context.Context, n Navigation) error {
	_, err := s.CreateStmt.ExecContext(ctx, n.Args()...)
	return err
}

// Read navigation from the table
func (s *NavigationService) readNavigation(
	ctx context.Context, week uint16, tow float32, do_query bool) ([]Navigation, error) {

	var rows *sql.Rows
	var err error
	if do_query {
		// fetch queried rows
		rows, err = s.ReadQueryStmt.QueryContext(ctx, week, tow)
	} else {
		// fetch latest row
		rows, err = s.ReadLatestStmt.QueryContext(ctx)
	}

	// add navigations to list
	var items []Navigation
	for rows.Next() {
		var n Navigation
		if err := rows.Scan(n.Args()...); err != nil {
			return nil, err
		}
		items = append(items, n)
	}

	return items, err
}

// Update a navigation from the table
func (s *NavigationService) updateNavigation(ctx context.Context, n Navigation, id int64) error {
	_, err := s.UpdateStmt.ExecContext(ctx, append(n.Args(), id)...)
	return err
}

// Delete a navigation from the table
func (s *NavigationService) deleteNavigation(ctx context.Context, id int64) error {
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
