package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	dvapi_db "github.com/lapuglisi/dvapi/database"
	dvapi_http "github.com/lapuglisi/dvapi/http"
	dvapi_model "github.com/lapuglisi/dvapi/model"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// For testing purposes, we will be using a separate database
// This can be done by copying ./dvapi.db.dist to ./dvadpi.test.db
const (
	AppTestDistDBFile string = "./dvapi.db.dist"
	AppTestDBFilePath string = "./dvapi.test.db"
)

// For this test, one array of devices will be used in all the tests
var devices dvapi_model.Devices
var apiServer dvapi_http.ApiHttpServer

func init() {
	fmt.Printf("--- Creating/Resetting file '%s'.\n", AppTestDBFilePath)
	if _, err := os.Stat(AppTestDBFilePath); err == nil {
		os.Remove(AppTestDBFilePath)
		// Remove any ${AppTestDBFilePath}.wal if needed
		os.Remove(fmt.Sprintf("%s.wal", AppTestDBFilePath))
	}

	// Copy ./dvapi.db.dist -> ./dvapi.test.db
	// Opens the source file
	fmt.Printf("--- Copying '%s' -> '%s'...\n", AppTestDistDBFile, AppTestDBFilePath)
	source, err := os.Open(AppTestDistDBFile)
	if err != nil {
		panic(err)
	}
	defer source.Close()

	// Opens the target file
	target, err := os.OpenFile(AppTestDBFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer target.Close()

	_, err = io.Copy(target, source)
	if err != nil {
		panic(err)
	}

	fmt.Printf("--- File '%s' created successfully.\n", AppTestDBFilePath)

	// Setup the api Server
	var db *dvapi_db.DuckDatabase = dvapi_db.NewDatabase()
	db.Setup(AppTestDBFilePath)
	apiServer.Setup("", 0, db)
}

// We will be creating two devices
// One with 'state' = 'available'
// Another with 'state' = 'in-use'
func TestDeviceCreate(t *testing.T) {
	devices = make(dvapi_model.Devices, 0)
	devices = append(devices,
		dvapi_model.Device{
			ID:    0,
			Name:  "TestDeviceOne",
			Brand: "BrandOne",
			State: "available",
			// CreatedOn: to be filled
		},
		dvapi_model.Device{
			ID:    0,
			Name:  "TestDeviceTwo",
			Brand: "BrandOne",
			State: "in-use",
			// CreatedOn: to be filled
		})

	// Iterate through our devices and try to create them
	for index, device := range devices {
		jsonBytes, err := device.ToJsonBytes()
		if err != nil {
			t.Fatal(err)
			break
		}

		req, err := http.NewRequest("POST", "/devices", bytes.NewBuffer(jsonBytes))
		if err != nil {
			t.Fatal(err)
			break
		}

		rr := httptest.NewRecorder()
		hh := http.HandlerFunc(apiServer.HandleDevicesCreate)
		hh.ServeHTTP(rr, req)

		// Check the return status
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("unexpected http status: got %d want %d\n", status, http.StatusOK)
			break
		}

		// Check if our returned json means success
		ar := dvapi_http.HttpApiResponse{}
		if err = json.Unmarshal(rr.Body.Bytes(), &ar); err != nil {
			t.Errorf("unexpected reponse from API: '%s'\n", rr.Body.String())
			break
		}

		//
		if ar.Status != "success" {
			t.Errorf("unexpected API status: got '%s' want 'success'\n", ar.Status)
			break
		} else {
			// Unmarshall the returned into struct
			json.Unmarshal([]byte(ar.Reason), &device)

			// Update our 'devices' array
			devices[index] = device
		}
	}

	// By this time, all devices have probably been created...
}

// TestDeviceUpdate:
// - Updating device1 ['state': 'available'] should succeed
// - Updating device2 ['state': "in-use'] should fail
func TestDeviceUpdate(t *testing.T) {
	//
	for _, device := range devices {
		jsonBytes, err := device.ToJsonBytes()
		if err != nil {
			t.Fatal(err)
			break
		}

		req, err := http.NewRequest("PATCH", "/devices", bytes.NewBuffer(jsonBytes))
		if err != nil {
			t.Fatal(err)
			break
		}

		// Setup changes in the device
		device.Name = "SomeRandomName"

		rr := httptest.NewRecorder()
		hh := http.HandlerFunc(apiServer.HandleDevicesUpdate)
		hh.ServeHTTP(rr, req)

		// Is the device state 'in-use'? Must return an error
		// Check the return status
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("unexpected http status: got %d want %d\n", status, http.StatusOK)
			break
		}

		// Check if our returned json means success
		ar := dvapi_http.HttpApiResponse{}
		if err = json.Unmarshal(rr.Body.Bytes(), &ar); err != nil {
			t.Errorf("unexpected reponse from API: '%s'\n", rr.Body.String())
			break
		}

		// Is the device in 'in-use' state? If so, it is fine to get status = 'error'
		// Would be nice to have some regex to test the 'Reason' description
		if device.State == "in-use" {
			if ar.Status != "error" {
				t.Errorf("unexpected API status for 'in-use' device: got '%s' want 'error'\n", ar.Status)
				break
			}
		} else {
			if ar.Status != "success" {
				t.Errorf("body: [%s]", rr.Body.String())
				t.Errorf("unexpected API status for 'not in-use' device: got '%s' want 'success'\n", ar.Status)
				break
			}
		}
	}
}

// TestDeviceDelete test for device deletion
func TestDeviceDelete(t *testing.T) {
	for _, device := range devices {
		jsonBytes, err := device.ToJsonBytes()
		if err != nil {
			t.Fatal(err)
			break
		}

		req, err := http.NewRequest("DELETE", "/devices", bytes.NewBuffer(jsonBytes))
		if err != nil {
			t.Fatal(err)
			break
		}

		rr := httptest.NewRecorder()
		hh := http.HandlerFunc(apiServer.HandleDevicesDelete)
		hh.ServeHTTP(rr, req)

		// Check the return status
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("unexpected http status: got %d want %d\n", status, http.StatusOK)
			break
		}

		// Check if our returned json means success
		ar := dvapi_http.HttpApiResponse{}
		if err = json.Unmarshal(rr.Body.Bytes(), &ar); err != nil {
			t.Errorf("unexpected reponse from API: '%s'\n", rr.Body.String())
			break
		}

		// Is the device in 'in-use' state? If so, it is fine to get status = 'error'
		// Would be nice to have some regex to test the 'Reason' description
		if device.State == "in-use" {
			if ar.Status != "error" {
				t.Errorf("unexpected API status for 'in-use' device: got '%s' want 'error'\n", ar.Status)
			}
		} else {
			if ar.Status != "success" {
				t.Errorf("unexpected API status for 'not in-use' device: got '%s' want 'success'\n", ar.Status)
				continue
			}

			// Try to fetch the device that has been deleted
			idJson := fmt.Sprintf(`{"id": %d}`, device.ID)
			req, err := http.NewRequest("GET", "/fetch", bytes.NewBuffer([]byte(idJson)))

			if err != nil {
				t.Errorf("Error while trying to fetch the device: %s", err.Error())
				continue
			}

			rr = nil
			rr := httptest.NewRecorder()

			hh := http.HandlerFunc(apiServer.HandleDevicesFetch)
			hh.ServeHTTP(rr, req)
			// Check the return status
			if status := rr.Code; status != http.StatusOK {
				t.Errorf("unexpected http status: got %d want %d\n", status, http.StatusOK)
				continue
			}

			// Check if our returned json means success
			ar := dvapi_http.HttpApiResponse{}
			if err = json.Unmarshal(rr.Body.Bytes(), &ar); err != nil {
				t.Errorf("unexpected reponse from API: '%s'\n", rr.Body.String())
				continue
			}

			if ar.Status != "error" {
				t.Errorf("unexpected status from API: got '%s' want 'error'\n", ar.Status)
				continue
			}
		}
	}
}

func TestFetchByBrand(t *testing.T) {
	brand := "BrandOne"

	req, err := http.NewRequest("GET", "/fetch/brand", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Set our 'brands' value when fetching devices by brand
	req.SetPathValue("brands", brand)

	rr := httptest.NewRecorder()
	hh := http.HandlerFunc(apiServer.HandleDevicesFetchByBrand)
	hh.ServeHTTP(rr, req)

	// Check the return status
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("unexpected http status: got %d want %d\n", status, http.StatusOK)
		return
	}

	// Check if our returned json means success
	ds := dvapi_model.Devices{}
	if err = json.Unmarshal(rr.Body.Bytes(), &ds); err != nil {
		t.Errorf("unexpected reponse from API: '%s'\n", rr.Body.String())
		return
	}

	// At this point of the test, we should have only one device registered
	if len(ds) != 1 {
		t.Errorf("wrong devices count: got %d want 1\n", len(ds))
		return
	}

	if ds[0].Brand != brand {
		t.Errorf("fetch by brands failed: returning device has brand '%s' != '%s'\n",
			ds[0].Brand, brand)
	}
}

func TestFetchByState(t *testing.T) {
	req, err := http.NewRequest("GET", "/fetch/state", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Set our 'brands' value when fetching devices by brand
	req.SetPathValue("states", "in-use,available")

	rr := httptest.NewRecorder()
	hh := http.HandlerFunc(apiServer.HandleDevicesFetchByState)
	hh.ServeHTTP(rr, req)

	// Check the return status
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("unexpected http status: got %d want %d\n", status, http.StatusOK)
		return
	}

	// Check if our returned json means success
	ds := dvapi_model.Devices{}
	if err = json.Unmarshal(rr.Body.Bytes(), &ds); err != nil {
		t.Errorf("unexpected reponse from API: '%s'\n", rr.Body.String())
		return
	}

	// At this point of the test, we should have only one device registered
	if len(ds) != 1 {
		t.Errorf("wrong devices count: got %d want 1\n", len(ds))
		return
	}

	if ds[0].State != "in-use" {
		t.Errorf("fetch by states failed: returning device has state '%s' != 'in-use'\n",
			ds[0].State)
	}
}
