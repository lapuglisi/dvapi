package api_db

import (
	"database/sql"
	"fmt"
	_ "github.com/duckdb/duckdb-go/v2"
	api_model "github.com/lapuglisi/dvapi/model"
	"strings"
	"time"
)

/*
* We will be using DuckDB as the database provider
 */

type DuckDatabase struct {
	db *sql.DB
}

type dbDevice struct {
	ID        int
	Name      string
	Brand     string
	State     string
	CreatedOn time.Time
}

// Function NewDatabse
// Returns a new pointer handle to a DuckDatabase instance
func NewDatabase() *DuckDatabase {
	return &DuckDatabase{}
}

func (ddb *DuckDatabase) Setup() (err error) {
	ddb.db, err = sql.Open("duckdb", "dvapi.db?access_mode=READ_WRITE")
	if err != nil {
		return err
	}

	if err = ddb.db.Ping(); err != nil {
		ddb.db.Close()
		return err
	}

	return nil
}

func (ddb *DuckDatabase) CreateDevice(device *api_model.Device) (err error) {
	stmt, err := ddb.db.Prepare("INSERT INTO devices (name, brand, state, created_on) VALUES ($1, $2, $3, NOW())")
	if err != nil {
		return err
	}

	result, err := stmt.Exec(device.Name, device.Brand, device.State)
	if err != nil {
		return err
	}

	if lastID, err := result.LastInsertId(); err == nil {
		device.ID = int(lastID)
	} else {
		fmt.Printf("could not get LastInsertId: %s\n", err.Error())
	}

	return err
}

func (ddb *DuckDatabase) UpdateDevice(device api_model.Device) (err error) {
	// Load the device first for fine-grained error messages
	current, err := ddb.loadDevice(device.ID)
	if err != nil {
		return err
	}

	// Apply some logic here
	if current.State == api_model.DeviceStateInUse {
		return fmt.Errorf("cannot update a device in 'in-use' state")
	}

	// Now check for input parameters
	if len(device.Name) == 0 {
		device.Name = current.Name
	}

	if len(device.Brand) == 0 {
		device.Brand = current.Brand
	}

	if len(device.State) == 0 {
		device.State = current.State
	}

	stmt, err := ddb.db.Prepare("UPDATE devices SET name = $2, brand = $3, state = $4 WHERE id = $1")
	if err != nil {
		return err
	}

	/*result*/
	_, err = stmt.Exec(device.ID, device.Name, device.Brand, device.State)
	if err != nil {
		return err
	}

	return nil
}

func (ddb *DuckDatabase) DeleteDevice(device api_model.Device) (err error) {
	// Load the device first for fine-grained error messages
	current, err := ddb.loadDevice(device.ID)
	if err != nil {
		return fmt.Errorf("could not determine the device to be deleted")
	}

	// Apply some logic here
	if current.State == api_model.DeviceStateInUse {
		return fmt.Errorf("cannot delete a device in 'in-use' state")
	}

	stmt, err := ddb.db.Prepare("DELETE FROM devices WHERE id = $1")
	if err != nil {
		return err
	}

	/*result*/
	_, err = stmt.Exec(device.ID)
	if err != nil {
		return err
	}

	return nil
}

func (ddb *DuckDatabase) Fetch(id int) (devices api_model.Devices, err error) {
	sql := fmt.Sprintf("SELECT id, name, brand, state, created_on FROM devices WHERE id = %d", id)
	var result dbDevice = dbDevice{}

	rows := ddb.db.QueryRow(sql)
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	err = rows.Scan(&result.ID, &result.Name, &result.Brand, &result.State, &result.CreatedOn)
	if err != nil {
		return nil, err
	}

	devices = append(devices, api_model.Device{
		ID:        result.ID,
		Name:      result.Name,
		Brand:     result.Brand,
		State:     result.State,
		CreatedOn: result.CreatedOn,
	})

	return devices, nil
}

// /
// / FetchAll retrieves all devices in the database
// / Consider retrieving a JSON object directly
func (ddb *DuckDatabase) FetchAll() (devices api_model.Devices, err error) {
	var sql string = "SELECT id, name, brand, state, created_on from devices order by created_on"
	var result dbDevice

	rows, err := ddb.db.Query(sql)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		result = dbDevice{}

		err = rows.Scan(&result.ID, &result.Name, &result.Brand, &result.State, &result.CreatedOn)
		if err != nil {
			break
		}

		devices = append(devices, api_model.Device{
			ID:        result.ID,
			Name:      result.Name,
			Brand:     result.Brand,
			State:     result.State,
			CreatedOn: result.CreatedOn,
		})
	}

	return devices, err // Keep err here
}

func (ddb *DuckDatabase) FetchByBrand(brands []string) (devices api_model.Devices, err error) {
	devices = api_model.Devices{}
	var totalBrands int = len(brands)

	if totalBrands == 0 {
		return nil, fmt.Errorf("no brand defined")
	}

	// I'll be using a poor man's approach
	// This is quite dumb actually, but anyway...
	sql := fmt.Sprintf("SELECT id, name, brand, state, created_on FROM devices WHERE brand IN (?%s)",
		strings.Repeat(", ?", totalBrands-1))

	stmt, err := ddb.db.Prepare(sql)
	if err != nil {
		return nil, err
	}

	// Now prepare the arguments for stmt.Query
	args := make([]any, totalBrands)
	for i, brand := range brands {
		args[i] = brand
	}

	fmt.Printf("stmt is %v\n", stmt)

	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	for rows.Next() {
		// Retrieve current row and append it to 'devices'
		r := api_model.Device{}
		if err = rows.Scan(&r.ID, &r.Name, &r.Brand, &r.State, &r.CreatedOn); err != nil {
			break
		}

		devices = append(devices, r)
	}

	if err != nil {
		return nil, err
	}

	return devices, nil
}

func (ddb *DuckDatabase) FetchByState(states []string) (devices api_model.Devices, err error) {
	devices = api_model.Devices{}
	var totalStates int = len(states)

	if totalStates == 0 {
		return nil, fmt.Errorf("no state defined")
	}

	// I'll be using a poor man's approach (once again)
	// This is quite dumb actually, but anyway...
	sql := fmt.Sprintf("SELECT id, name, brand, state, created_on FROM devices WHERE state IN (?%s)",
		strings.Repeat(", ?", totalStates-1))

	stmt, err := ddb.db.Prepare(sql)
	if err != nil {
		return nil, err
	}

	// Now prepare the arguments for stmt.Query
	args := make([]any, totalStates)
	for i, state := range states {
		args[i] = state
	}

	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	for rows.Next() {
		// Retrieve current row and append it to 'devices'
		r := api_model.Device{}
		if err = rows.Scan(&r.ID, &r.Name, &r.Brand, &r.State, &r.CreatedOn); err != nil {
			break
		}

		devices = append(devices, r)
	}

	if err != nil {
		return nil, err
	}

	return devices, nil
}

func (ddb *DuckDatabase) loadDevice(id int) (device *api_model.Device, err error) {
	var result dbDevice = dbDevice{}
	var rows *sql.Row = nil

	stmt, err := ddb.db.Prepare("SELECT id, name, brand, state, created_on FROM devices WHERE id = $1")

	if err != nil {
		return nil, err
	}

	if rows = stmt.QueryRow(id); rows.Err() != nil {
		return nil, err
	}

	err = rows.Scan(&result.ID, &result.Name, &result.Brand, &result.State, &result.CreatedOn)
	if err != nil {
		return nil, err
	}

	return &api_model.Device{
		ID:        result.ID,
		Name:      result.Name,
		Brand:     result.Brand,
		State:     result.State,
		CreatedOn: result.CreatedOn,
	}, nil
}

func (ddb *DuckDatabase) Release() (err error) {
	// TODO: Check error type before return
	return ddb.db.Close()
}
