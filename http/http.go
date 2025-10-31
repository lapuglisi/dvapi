package api_http

import (
	"encoding/json"
	"fmt"
	api_db "github.com/lapuglisi/dvapi/database"
	api_model "github.com/lapuglisi/dvapi/model"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// The structure that holds the ApiHttpServer implementation
type ApiHttpServer struct {
	listenUri string
	db        *api_db.DuckDatabase
}

// HttpErrorResponse is used to send errors to a http.Request
type HttpApiResponse struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

func init() {
}

func (s *ApiHttpServer) writeApiReponse(w http.ResponseWriter, e HttpApiResponse) error {
	var err error

	if jsonBytes, err := json.Marshal(e); err == nil {
		return s.writeResponseJson(w, jsonBytes)
	}

	return err
}

func (s *ApiHttpServer) writeResponseJson(w http.ResponseWriter, bytes []byte) (err error) {
	w.Header().Set("Content-Type", "application/json")
	total, err := w.Write(bytes)

	log.Printf("writeReponseJson: wrote %d bytes\n", total)

	return err
}

func (s *ApiHttpServer) Setup(host string, port int, db *api_db.DuckDatabase) {
	if len(host) == 0 {
		host = "0.0.0.0"
	}

	s.listenUri = fmt.Sprintf("%s:%d", host, port)

	if s.db = db; s.db == nil {
		log.Fatal("ApiHttpServer: No database handle defined")
	}

	// Setup endpoints here
	http.HandleFunc("/devices", s.handleDevices)
	http.HandleFunc("/fetch", s.handleDevicesFetchAll)
	http.HandleFunc("GET /fetch/id/{id}", s.handleDevicesFetch)
	http.HandleFunc("GET /fetch/brand/{brands}", s.handleDevicesFetchByBrand)
	http.HandleFunc("GET /fetch/state/{states}", s.handleDevicesFetchByState)
}

func (s *ApiHttpServer) Run() error {
	fmt.Println("\033[32minfo\033[0m: dvapi listening on", s.listenUri)
	return http.ListenAndServe(s.listenUri, nil)
}

// I'll be using actual implementations of the handlers instead of defining them when calling HandleFunc

// handleDevicesCreate: triggered when handleDevices receives a POST request
func (s *ApiHttpServer) handleDevicesCreate(w http.ResponseWriter, r *http.Request) {
	var device api_model.Device
	var jsonBytes []byte = make([]byte, 0)
	var err error

	jsonBytes, err = io.ReadAll(r.Body)
	if err != nil {
		s.writeApiReponse(w, HttpApiResponse{
			Status: "error",
			Reason: fmt.Sprintf("could not create the device: %s", err.Error()),
		})

		return
	}
	defer r.Body.Close()

	log.Printf("[handleDevicesCreate] received data '%s'\n", string(jsonBytes))

	if err = device.FromJsonBytes(jsonBytes); err != nil {
		s.writeApiReponse(w, HttpApiResponse{
			Status: "error",
			Reason: fmt.Sprintf("could not create the device: %s", err.Error()),
		})

		return
	}

	// Insert the new device into the database
	if err = s.db.CreateDevice(&device); err != nil {
		s.writeApiReponse(w, HttpApiResponse{
			Status: "error",
			Reason: fmt.Sprintf("could not create the device: %s", err.Error()),
		})

		return
	}

	s.writeApiReponse(w, HttpApiResponse{
		Status: "success",
		Reason: fmt.Sprintf("device created succesfully. id is %d\n", device.ID),
	})
}

// Triggered when handleDevices receives a PUT request
func (s *ApiHttpServer) handleDevicesUpdate(w http.ResponseWriter, r *http.Request) {
	var device api_model.Device
	var jsonBytes []byte = make([]byte, 0)
	var err error

	jsonBytes, err = io.ReadAll(r.Body)
	if err != nil {
		s.writeApiReponse(w, HttpApiResponse{
			Status: "error",
			Reason: fmt.Sprintf("could not update the device: %s", err.Error()),
		})

		return
	}
	defer r.Body.Close()

	log.Printf("[handleDevicesUpdate] received data '%s'\n", string(jsonBytes))

	if err = device.FromJsonBytes(jsonBytes); err != nil {
		s.writeApiReponse(w, HttpApiResponse{
			Status: "error",
			Reason: fmt.Sprintf("invalid JSON data: %s", err.Error()),
		})

		return
	}

	// Uupdate the in the database
	if err = s.db.UpdateDevice(device); err != nil {
		s.writeApiReponse(w, HttpApiResponse{
			Status: "error",
			Reason: fmt.Sprintf("could not update the device: %s", err.Error()),
		})

		return
	}

	s.writeApiReponse(w, HttpApiResponse{
		Status: "success",
		Reason: "device updated succesfully",
	})
}

// Triggered when handleDevices receives a DELETE request
func (s *ApiHttpServer) handleDevicesDelete(w http.ResponseWriter, r *http.Request) {
	var device api_model.Device
	var jsonBytes []byte = make([]byte, 0)
	var err error

	jsonBytes, err = io.ReadAll(r.Body)
	if err != nil {
		s.writeApiReponse(w, HttpApiResponse{
			Status: "error",
			Reason: fmt.Sprintf("could not delete the device: %s", err.Error()),
		})

		return
	}
	defer r.Body.Close()

	log.Printf("[handleDevicesDelete] received data '%s'\n", string(jsonBytes))

	if err = device.FromJsonBytes(jsonBytes); err != nil {
		s.writeApiReponse(w, HttpApiResponse{
			Status: "error",
			Reason: fmt.Sprintf("invalid JSON data: %s", err.Error()),
		})

		return
	}

	// Delte the device from the database
	if err = s.db.DeleteDevice(device); err != nil {
		s.writeApiReponse(w, HttpApiResponse{
			Status: "error",
			Reason: fmt.Sprintf("could not delete the device: %s", err.Error()),
		})

		return
	}

	s.writeApiReponse(w, HttpApiResponse{
		Status: "success",
		Reason: "device deleted succesfully",
	})

}

// When handleDevices receives a GET request
func (s *ApiHttpServer) handleDevicesFetch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var devices api_model.Devices = nil
	var err error = nil
	var jsonBytes []byte = nil

	deviceID, err := strconv.Atoi(r.PathValue("id"))

	if devices, err = s.db.Fetch(deviceID); err != nil {
		s.writeApiReponse(w, HttpApiResponse{
			Status: "error",
			Reason: fmt.Sprintf("could not fetch devices: %s", err.Error()),
		})
		return
	}

	// Check if len(devices) > 0 just in case
	if len(devices) == 0 {
		s.writeApiReponse(w, HttpApiResponse{
			Status: "error",
			Reason: "no devices found",
		})
		return
	}

	if jsonBytes, err = devices.ToJsonBytes(); err != nil {
		// same thing as above
		s.writeApiReponse(w, HttpApiResponse{
			Status: "error",
			Reason: fmt.Sprintf("could not fetch devices: %s", err.Error()),
		})
		return
	}

	s.writeResponseJson(w, jsonBytes)
}

func (s *ApiHttpServer) handleDevicesFetchAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var devices api_model.Devices
	var err error
	if devices, err = s.db.FetchAll(); err != nil {
		s.writeApiReponse(w, HttpApiResponse{
			Status: "error",
			Reason: err.Error(),
		})

		return
	}

	jsonBytes, err := devices.ToJsonBytes()
	if err != nil {
		s.writeApiReponse(w, HttpApiResponse{
			Status: "error",
			Reason: err.Error(),
		})
		return
	}

	s.writeResponseJson(w, jsonBytes)
}

func (s *ApiHttpServer) handleDevicesFetchByBrand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var devices api_model.Devices
	var err error

	args := r.PathValue("brands")
	brands := strings.Split(args, ",")

	if devices, err = s.db.FetchByBrand(brands); err != nil {
		s.writeApiReponse(w, HttpApiResponse{
			Status: "error",
			Reason: err.Error(),
		})

		return
	}

	jsonBytes, err := devices.ToJsonBytes()
	if err != nil {
		s.writeApiReponse(w, HttpApiResponse{
			Status: "error",
			Reason: err.Error(),
		})
		return
	}

	s.writeResponseJson(w, jsonBytes)

}

func (s *ApiHttpServer) handleDevicesFetchByState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var devices api_model.Devices
	var err error

	args := r.PathValue("states")
	states := strings.Split(args, ",")

	if devices, err = s.db.FetchByState(states); err != nil {
		s.writeApiReponse(w, HttpApiResponse{
			Status: "error",
			Reason: err.Error(),
		})

		return
	}

	jsonBytes, err := devices.ToJsonBytes()
	if err != nil {
		s.writeApiReponse(w, HttpApiResponse{
			Status: "error",
			Reason: err.Error(),
		})
		return
	}

	s.writeResponseJson(w, jsonBytes)
}

// handleDevices is the catch-all for devices operations
func (s *ApiHttpServer) handleDevices(w http.ResponseWriter, r *http.Request) {

	// Accepted methods are GET, POST, PUT, DELETE
	switch r.Method {
	case http.MethodPost:
		{
			s.handleDevicesCreate(w, r)
			break
		}

	case http.MethodPut:
		{
			s.handleDevicesUpdate(w, r)
			break
		}

	case http.MethodDelete:
		{
			s.handleDevicesDelete(w, r)
			break
		}

	default:
		{
			// Method not accepted. Return not supported
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
