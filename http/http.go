package dvapi_http

import (
	"encoding/json"
	"fmt"
	dvapi_db "github.com/lapuglisi/dvapi/database"
	dvapi_model "github.com/lapuglisi/dvapi/model"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// Constants
const (
	ApiServerDefaultPort int    = 9098
	ApiServerDefaultHost string = "0.0.0.0"
)

// The structure that holds the ApiHttpServer implementation
type ApiHttpServer struct {
	listenUri string
	db        *dvapi_db.DuckDatabase
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
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	total, err := w.Write(bytes)

	log.Printf("writeReponseJson: wrote %d bytes\n", total)

	return err
}

// Setup sets up our ApiHttpServer instance
func (s *ApiHttpServer) Setup(host string, port int, db *dvapi_db.DuckDatabase) {
	if len(host) == 0 {
		host = ApiServerDefaultHost
	}

	if port <= 0 {
		port = ApiServerDefaultPort
	}

	// Format the listen address based on the arguments
	s.listenUri = fmt.Sprintf("%s:%d", host, port)

	if s.db = db; s.db == nil {
		log.Fatal("ApiHttpServer: No database handle defined")
	}

	// Setup the endpoints here
	http.HandleFunc("/devices", s.handleDevices)
	http.HandleFunc("GET /fetch", s.HandleDevicesFetchAll)
	http.HandleFunc("GET /fetch/id/{id}", s.HandleDevicesFetch)
	http.HandleFunc("GET /fetch/brand/{brands}", s.HandleDevicesFetchByBrand)
	http.HandleFunc("GET /fetch/state/{states}", s.HandleDevicesFetchByState)
}

func (s *ApiHttpServer) Run() error {
	fmt.Println("\033[32minfo\033[0m: dvapi listening on", s.listenUri)
	return http.ListenAndServe(s.listenUri, nil)
}

// HandleDevicesCreate is triggered when handleDevices receives a POST request
func (s *ApiHttpServer) HandleDevicesCreate(w http.ResponseWriter, r *http.Request) {
	var device dvapi_model.Device
	var jsonBytes []byte = make([]byte, 0)
	var err error

	jsonBytes, err = io.ReadAll(r.Body)
	if err != nil {
		s.writeApiReponse(w, HttpApiResponse{
			Status: "error",
			Reason: fmt.Sprintf("create device: %s", err.Error()),
		})

		return
	}
	defer r.Body.Close()

	log.Printf("[handleDevicesCreate] received data '%s'\n", string(jsonBytes))

	if err = device.FromJsonBytes(jsonBytes); err != nil {
		s.writeApiReponse(w, HttpApiResponse{
			Status: "error",
			Reason: fmt.Sprintf("create device: %s", err.Error()),
		})

		return
	}

	// Insert the new device into the database
	if err = s.db.CreateDevice(&device); err != nil {
		s.writeApiReponse(w, HttpApiResponse{
			Status: "error",
			Reason: fmt.Sprintf("create device: %s", err.Error()),
		})

		return
	}

	s.writeApiReponse(w, HttpApiResponse{
		Status: "success",
		Reason: fmt.Sprintf("device created succesfully."),
	})
}

// HandleDevicesUpdate is triggered when handleDevices receives a PATCH request
func (s *ApiHttpServer) HandleDevicesUpdate(w http.ResponseWriter, r *http.Request) {
	var device dvapi_model.Device
	var jsonBytes []byte = make([]byte, 0)
	var err error

	jsonBytes, err = io.ReadAll(r.Body)
	if err != nil {
		s.writeApiReponse(w, HttpApiResponse{
			Status: "error",
			Reason: fmt.Sprintf("update device: %s", err.Error()),
		})

		return
	}
	defer r.Body.Close()

	log.Printf("[handleDevicesUpdate] received data '%s'\n", string(jsonBytes))

	if err = device.FromJsonBytes(jsonBytes); err != nil {
		s.writeApiReponse(w, HttpApiResponse{
			Status: "error",
			Reason: fmt.Sprintf("update device: %s", err.Error()),
		})

		return
	}

	// Uupdate the in the database
	if err = s.db.UpdateDevice(device); err != nil {
		s.writeApiReponse(w, HttpApiResponse{
			Status: "error",
			Reason: fmt.Sprintf("update device: %s", err.Error()),
		})

		return
	}

	s.writeApiReponse(w, HttpApiResponse{
		Status: "success",
		Reason: "device updated succesfully",
	})
}

// HandleDevicesDelete is triggered when handleDevices receives a DELETE request
func (s *ApiHttpServer) HandleDevicesDelete(w http.ResponseWriter, r *http.Request) {
	var device dvapi_model.Device
	var jsonBytes []byte = make([]byte, 0)
	var err error

	jsonBytes, err = io.ReadAll(r.Body)
	if err != nil {
		s.writeApiReponse(w, HttpApiResponse{
			Status: "error",
			Reason: fmt.Sprintf("delete device: %s", err.Error()),
		})

		return
	}
	defer r.Body.Close()

	log.Printf("[handleDevicesDelete] received data '%s'\n", string(jsonBytes))

	if err = device.FromJsonBytes(jsonBytes); err != nil {
		s.writeApiReponse(w, HttpApiResponse{
			Status: "error",
			Reason: fmt.Sprintf("delete device: %s", err.Error()),
		})

		return
	}

	// Delte the device from the database
	if err = s.db.DeleteDevice(device); err != nil {
		s.writeApiReponse(w, HttpApiResponse{
			Status: "error",
			Reason: fmt.Sprintf("delete device: %s", err.Error()),
		})

		return
	}

	s.writeApiReponse(w, HttpApiResponse{
		Status: "success",
		Reason: "device deleted succesfully",
	})

}

// HandleDevicesFetch is triggered when the API receives a 'GET /fetch/id/{id}' request
func (s *ApiHttpServer) HandleDevicesFetch(w http.ResponseWriter, r *http.Request) {
	/* Leave it here just as a reminder
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	*/

	var devices dvapi_model.Devices = nil
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

// HandleDevicesFetchAll is triggered when the API receives a 'GET /fetch' request
func (s *ApiHttpServer) HandleDevicesFetchAll(w http.ResponseWriter, r *http.Request) {
	/* Leave it here as a reminder
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	*/

	var devices dvapi_model.Devices
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

// HandleDevicesFetchByBrand is triggered when
// the API receives a 'GET /fetch/brand/{brands}' request
// - {brands} is a comma delimited string: Brand1[,Brand2,Brand3]
func (s *ApiHttpServer) HandleDevicesFetchByBrand(w http.ResponseWriter, r *http.Request) {
	/* Leave it here as a reminder
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	*/

	var devices dvapi_model.Devices
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

// HandleDevicesFetchByState is triggered when
// the API receives a 'GET /fetch/state/{states}' request
// - {states} is a comma delimited string: state1[,state2,state3]
func (s *ApiHttpServer) HandleDevicesFetchByState(w http.ResponseWriter, r *http.Request) {
	/* Have I said to leave it here as a reminder already?
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	*/

	var devices dvapi_model.Devices
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

// handleDevices is the catch-all handler for devices operations
// It does not need to be exported
func (s *ApiHttpServer) handleDevices(w http.ResponseWriter, r *http.Request) {

	// Accepted methods are POST, PUT, DELETE
	switch r.Method {
	case http.MethodPost:
		{
			s.HandleDevicesCreate(w, r)
			break
		}

	case http.MethodPatch:
		{
			s.HandleDevicesUpdate(w, r)
			break
		}

	case http.MethodDelete:
		{
			s.HandleDevicesDelete(w, r)
			break
		}

	default:
		{
			// Method not accepted. Return not supported
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
