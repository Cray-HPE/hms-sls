// MIT License
//
// (C) Copyright [2019, 2021-2022] Hewlett Packard Enterprise Development LP
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	sls_common "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"
	"github.com/Cray-HPE/hms-xname/xnametypes"

	"github.com/Cray-HPE/hms-sls/v2/internal/database"

	base "github.com/Cray-HPE/hms-base/v2"
	"github.com/Cray-HPE/hms-sls/v2/internal/datastore"
)

// Used for response functions

type Response struct {
	C int    `json:"code"`
	M string `json:"message"`
}

var mapVersion int
var mapTimestamp string

const (
	SLS_VERSION_KEY = "slsVersion"
	SLS_HEALTH_KEY  = "SLS_HEALTH_KEY"
	SLS_HEALTH_VAL  = "SLS_OK"
)

// Fetch the version info from the DB.

func getVersionFromDB() (version sls_common.SLSVersion, err error) {
	currentVersion, err := database.GetCurrentVersion()
	if err != nil {
		log.Println("ERROR: Can't get current version:", err)
		return
	}

	lastModified, err := database.GetLastModified()
	if err != nil {
		log.Println("ERROR: Can't get last modified:", err)
		return
	}

	version = sls_common.SLSVersion{
		Counter:     currentVersion,
		LastUpdated: lastModified,
	}

	return
}

// Check if the database is ready.

func dbReady() bool {
	_, serr := getVersionFromDB()
	if serr != nil {
		log.Println("INFO: Readiness check failed, can't get version info from DB:", serr)
		return false
	}

	return true
}

// /verion API: Get the current version information.

func doVersionGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		log.Printf("ERROR: Bad request type for '%s': %s\n", r.URL.Path,
			r.Method)
		pdet := base.NewProblemDetails("about:blank",
			"Invalid Request",
			"Only GET operations supported.",
			r.URL.Path, http.StatusMethodNotAllowed)
		w.Header().Add("Allow", "GET")
		base.SendProblemDetails(w, pdet, 0)
	}

	// Grab the version info from the DB

	slsVersion, slserr := getVersionFromDB()
	if slserr != nil {
		log.Println("ERROR: Can't get version info from DB:", slserr)
		pdet := base.NewProblemDetails("about: blank",
			"Internal Server Error",
			"Unable to get version info from DB",
			r.URL.Path, http.StatusInternalServerError)
		base.SendProblemDetails(w, pdet, 0)
	}

	ba, err := json.Marshal(slsVersion)
	if err != nil {
		log.Println("ERROR: JSON marshal of version info failed:", err)
		pdet := base.NewProblemDetails("about: blank",
			"Internal Server Error",
			"JSON marshal error",
			r.URL.Path, http.StatusInternalServerError)
		base.SendProblemDetails(w, pdet, 0)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(ba)
}

// HealthResponse - used to report service health stats
type HealthResponse struct {
	Vault        string `json:"Vault"`
	DBConnection string `json:"DBConnection"`
}

func doHealthGet(w http.ResponseWriter, r *http.Request) {
	// NOTE: this is provided as a debugging aid for administrators to
	//  find out what is going on with the system.  This should return
	//  information in a human-readable format that will help to
	//  determine the state of this service.

	log.Printf("INFO: entering health check")
	if r.Method != "GET" {
		log.Printf("ERROR: Bad request type for '%s': %s\n", r.URL.Path,
			r.Method)
		pdet := base.NewProblemDetails("about:blank",
			"Invalid Request",
			"Only GET operations supported.",
			r.URL.Path, http.StatusMethodNotAllowed)
		w.Header().Add("Allow", "GET")
		base.SendProblemDetails(w, pdet, 0)
	}

	var stats HealthResponse

	// Vault is no longer used by SLS
	stats.Vault = "Not checked"

	//Check that ETCD/DB connection is available
	if database.DB == nil {
		log.Printf("INFO: DB not initialized")
		stats.DBConnection = "Not Initialized"
	} else {
		// NOTE - the Ping command will restore a dropped connection
		dberr := database.DB.Ping()
		if dberr != nil {
			log.Printf("INFO: DB ping error:%s", dberr.Error())
			stats.DBConnection = fmt.Sprintf("Ping error:%s", dberr.Error())
		} else if dbReady() == false {
			// active query from something in database
			log.Printf("INFO: DB not Ready")
			stats.DBConnection = "Not Ready"
		} else {
			log.Printf("INFO: DB Ready")
			stats.DBConnection = "Ready"
		}
	}

	// marshal and send the response
	ba, err := json.Marshal(stats)
	if err != nil {
		log.Println("ERROR: JSON marshal of readiness info failed:", err)
		pdet := base.NewProblemDetails("about: blank",
			"Internal Server Error",
			"JSON marshal error",
			r.URL.Path, http.StatusInternalServerError)
		base.SendProblemDetails(w, pdet, 0)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(ba)
}

func doLivenessGet(w http.ResponseWriter, r *http.Request) {
	// NOTE: this is coded in accordance with kubernetes best practices
	//  for liveness/readiness checks.  This function should only be
	//  used to indicate the server is still alive and processing requests.

	if r.Method != "GET" {
		log.Printf("ERROR: Bad request type for '%s': %s\n", r.URL.Path, r.Method)
		pdet := base.NewProblemDetails("about:blank",
			"Invalid Request",
			"Only GET operations supported.",
			r.URL.Path, http.StatusMethodNotAllowed)
		w.Header().Add("Allow", "GET")
		base.SendProblemDetails(w, pdet, 0)
	}

	w.WriteHeader(http.StatusNoContent)
}

func doReadinessGet(w http.ResponseWriter, r *http.Request) {
	// NOTE: this is coded in accordance with kubernetes best practices
	//  for liveness/readiness checks.  This function should only be
	//  used to indicate if something is wrong with this service that
	//  prevents usage.  If this fails too many times, the instance
	//  will be killed and re-started.  Only fail this if restarting
	//  this service is likely to fix the problem.

	if r.Method != "GET" {
		log.Printf("ERROR: Bad request type for '%s': %s\n", r.URL.Path,
			r.Method)
		pdet := base.NewProblemDetails("about:blank",
			"Invalid Request",
			"Only GET operations supported.",
			r.URL.Path, http.StatusMethodNotAllowed)
		w.Header().Add("Allow", "GET")
		base.SendProblemDetails(w, pdet, 0)
	}

	ready := true
	if dbReady() == false {
		log.Printf("INFO: readiness check fails, db not ready")
		ready = false
	}

	if ready {
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}

// POST /dumpstate API
// This used to be supported but is no longer supported.
func doPostDumpState(w http.ResponseWriter, r *http.Request) {
	pdet := base.NewProblemDetails("about: blank",
		"Method Not Allowed",
		"POST for dumpstate is not supported. Use GET instead.",
		r.URL.Path, http.StatusMethodNotAllowed)
	base.SendProblemDetails(w, pdet, 0)
	return
}

// GET /dumpstate API
func doDumpState(w http.ResponseWriter, r *http.Request) {
	ret := sls_common.SLSState{
		Hardware: make(map[string]sls_common.GenericHardware),
		Networks: make(map[string]sls_common.Network),
	}

	allHardware, err := datastore.GetAllHardware()
	if err != nil {
		log.Println("ERROR: unable to get hardware: ", err)
		pdet := base.NewProblemDetails("about: blank",
			"Internal Server Error",
			"Failed to get hardware",
			r.URL.Path, http.StatusInternalServerError)
		base.SendProblemDetails(w, pdet, 0)
		return
	}

	for _, hardware := range allHardware {
		ret.Hardware[hardware.Xname] = hardware
	}

	allNetworks, err := datastore.GetAllNetworks()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		log.Printf("ERROR: Unable to load existing networks: %s", err)
		return
	}
	for _, network := range allNetworks {
		ret.Networks[network.Name] = network
	}

	jsonBytes, err := json.Marshal(ret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		log.Printf("ERROR: Unable to create json: %s", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

//  /loadstate API

func doLoadState(w http.ResponseWriter, r *http.Request) {
	var inputData sls_common.SLSState
	var buf bytes.Buffer

	// Check for the deprecated and ignored private_key form file
	_, _, privateKeyErr := r.FormFile("private_key")
	if privateKeyErr != http.ErrMissingFile || privateKeyErr == nil {
		log.Println("INFO: loadstate: Ignoring the private_key option, which is no longer supported.")
	}

	// Now get the config file to read back in.
	configFile, _, configFileErr := r.FormFile("sls_dump")
	if configFileErr != nil {
		log.Println("ERROR: loadstate: Unable to parse SLS dump form file: ", configFileErr)
		pdet := base.NewProblemDetails("about: blank",
			"Bad Request",
			"Unable to parse SLS dump form file",
			r.URL.Path, http.StatusBadRequest)
		base.SendProblemDetails(w, pdet, 0)
		return
	}

	io.Copy(&buf, configFile)
	configFile.Close()

	// Read the body, let's turn it into back into JSON
	marshalErr := json.Unmarshal(buf.Bytes(), &inputData)
	if marshalErr != nil {
		log.Println("ERROR: loadstate: Unable to unmarshal config file: ", marshalErr)
		pdet := base.NewProblemDetails("about: blank",
			"Bad Request",
			"Unable to unmarshal config file",
			r.URL.Path, http.StatusBadRequest)
		base.SendProblemDetails(w, pdet, 0)
		return
	}

	containsVaultData := false
	// Validate hardware
	for i, h := range inputData.Hardware {
		originalXname := h.Xname
		h.Xname = xnametypes.NormalizeHMSCompID(h.Xname)
		if !xnametypes.IsHMSCompIDValid(h.Xname) {
			pdet := base.NewProblemDetails("about: blank",
				"Bad Request",
				fmt.Sprintf("Invalid xname: %s", h.Xname),
				r.URL.Path, http.StatusBadRequest)
			base.SendProblemDetails(w, pdet, 0)
			log.Printf("ERROR: loadstate: invalid hardware: %s - %s\n", pdet.Title, pdet.Detail)
			return
		}
		if h.Xname != originalXname {
			log.Printf("INFO: loadstate: Normalized xname. requested: %s, changed to: %s\n", originalXname, h.Xname)
		}

		class := h.Class
		if class == "" {
			pdet := base.NewProblemDetails("about: blank",
				"Bad Request",
				fmt.Sprintf("Missing Class field for %s", h.Xname),
				r.URL.Path, http.StatusBadRequest)
			base.SendProblemDetails(w, pdet, 0)
			log.Printf("ERROR: loadstate: invalid hardware: %s - %s\n", pdet.Title, pdet.Detail)
			return
		}

		err := datastore.NormalizeFields(&h)
		if err != nil {
			pdet := base.NewProblemDetails("about: blank",
				"Bad Request",
				fmt.Sprintf("Failure normalizing fields for %s. %v", h.Xname, err),
				r.URL.Path, http.StatusBadRequest)
			base.SendProblemDetails(w, pdet, 0)
			log.Printf("ERROR: loadstate: invalid hardware: %s - %s\n", pdet.Title, pdet.Detail)
			return
		}

		p := xnametypes.GetHMSCompParent(h.Xname)
		if p != h.Parent {
			log.Printf("INFO: loadstate: Parent mismatch. xname: %s, requested: %s, changed to: %s\n", h.Xname, h.Parent, p)
		}
		h.Parent = p

		ts := xnametypes.GetHMSType(h.Xname)
		if ts != h.TypeString {
			log.Printf("INFO: loadstate: TypeString mismatch. xname: %s, requested: %s, changed to: %s\n", h.Xname, h.TypeString, ts)
		}
		h.TypeString = ts

		t := sls_common.HMSTypeToHMSStringType(h.TypeString)
		if t != h.Type {
			log.Printf("INFO: loadstate: Type mismatch. xname: %s, requested: %s, changed to: %s\n", h.Xname, h.Type, t)
		}
		h.Type = t

		err = datastore.ValidateFields(h)
		if err != nil {
			pdet := base.NewProblemDetails("about: blank",
				"Bad Request",
				fmt.Sprintf("Invalid hardware data for %s. %v", h.Xname, err),
				r.URL.Path, http.StatusBadRequest)
			base.SendProblemDetails(w, pdet, 0)
			log.Printf("ERROR: loadstate: invalid hardware: %s - %s\n", pdet.Title, pdet.Detail)
			return
		}
		if h.VaultData != nil {
			containsVaultData = true
			h.VaultData = nil
		}
		inputData.Hardware[i] = h
	}

	if containsVaultData {
		log.Println("INFO: loadstate: Ignoring VaultData. Loading the vault data is no longer supported.")
	}

	// Validate networks
	for _, n := range inputData.Networks {
		err := datastore.VerifyNetwork(n)
		if err != nil {
			pdet := base.NewProblemDetails("about: blank",
				"Bad Request",
				fmt.Sprintf("Invalid network data for %s. %v", n.Name, err),
				r.URL.Path, http.StatusBadRequest)
			base.SendProblemDetails(w, pdet, 0)
			log.Printf("ERROR: loadstate: invalid networks: %s - %s\n", pdet.Title, pdet.Detail)
			return
		}
	}

	// Finally we are ready to put the info back into the database.
	var hardware []sls_common.GenericHardware
	var networks []sls_common.Network

	for _, obj := range inputData.Hardware {
		hardware = append(hardware, obj)
	}

	hardwareErr := datastore.ReplaceGenericHardware(hardware)
	if hardwareErr != nil {
		log.Println("ERROR: unable to replace hardware:", hardwareErr)
		pdet := base.NewProblemDetails("about: blank",
			"Internal Server Error",
			"Failed to replace hardware",
			r.URL.Path, http.StatusInternalServerError)
		base.SendProblemDetails(w, pdet, 0)
		return
	}

	for _, obj := range inputData.Networks {
		networks = append(networks, obj)
	}

	networksErr := datastore.ReplaceAllNetworks(networks)
	if networksErr != nil {
		log.Println("ERROR: unable to replace networks:", networksErr)
		pdet := base.NewProblemDetails("about: blank",
			"Internal Server Error",
			"Failed to replace networks",
			r.URL.Path, http.StatusInternalServerError)
		base.SendProblemDetails(w, pdet, 0)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Send a JSON response.  If the ecode indicates an error, send
// a properly formatted RFC7807 problem.
// If it does not, fall back to the original CAPMC format, which will
// now just be used for success cases or odd HTTP status codes that
// don't suggest a RFC7807 problem response.
// We use the 7807 problem format for 4xx and 5xx errors, though
// in practice the latter (server errors) will probably not be used here
// as they do not describe invalid requests but server-specific issues.

func sendJsonRsp(w http.ResponseWriter, ecode int, message string) {
	if ecode < 400 {
		sendJsonResponse(w, ecode, message)
	} else {
		// Use library function in HMS base.  Problem will be
		// a generic one with title matching the HTTP Status code text
		// with message as the details field.  For this type of problem
		// title can just be set to "about:blank" so we need no
		// custom URL.  The optional instance field is omitted as well
		// so no URL/URI is needed there either.
		base.SendProblemDetailsGeneric(w, ecode, message)
	}
}

// Send a simple message for cases where need a non-error response.  If
// a more feature filled message needs to be returned then do it with a
// different function.  Code is the http status response, converted to
// zero for success-related responses.
func sendJsonResponse(w http.ResponseWriter, ecode int, message string) {
	// if the HTTP call was a success then put zero in the returned json
	// error field. This is what capmc does.
	http_code := ecode
	if ecode >= 200 && ecode <= 299 {
		ecode = 0
	}
	data := Response{ecode, message}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if http_code != http.StatusNoContent {
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			log.Printf("Yikes, I couldn't encode a JSON status response: %s\n", err)
		}
	}
}

func sendJsonCompRsp(w http.ResponseWriter, comp sls_common.GenericHardware, http_code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	err := json.NewEncoder(w).Encode(comp)
	if err != nil {
		log.Printf("Couldn't encode a JSON command response: %s\n", err)
	}
}

func sendJsonCompArrayRsp(w http.ResponseWriter, comps []sls_common.GenericHardware) {
	http_code := 200
	if len(comps) == 0 {
		http_code = 204
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if len(comps) == 0 {
		err := json.NewEncoder(w).Encode(comps)
		if err != nil {
			log.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}
