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
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	sls_common "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"

	base "github.com/Cray-HPE/hms-base/v2"
	"github.com/Cray-HPE/hms-sls/v2/internal/database"
	"github.com/Cray-HPE/hms-xname/xnametypes"

	"github.com/Cray-HPE/hms-sls/v2/internal/datastore"
	"github.com/gorilla/mux"
)

// struct for the post /hardware request body
type postHardwareRequest struct {
	Xname              string                 `json:"Xname"`
	Class              sls_common.CabinetType `json:"Class"`
	ExtraPropertiesRaw interface{}            `json:"ExtraProperties,omitempty"`
}

// struct for the put /hardware/{xname} request body
type putHardwareRequest struct {
	Class              sls_common.CabinetType `json:"Class"`
	ExtraPropertiesRaw interface{}            `json:"ExtraProperties,omitempty"`
}

//  /hardware POST API

func doHardwarePost(w http.ResponseWriter, r *http.Request) {
	var jdata postHardwareRequest

	// Decode the JSON to see what we are to post

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("ERROR reading request body:", err)
		sendJsonRsp(w, http.StatusInternalServerError, "error reading REST request")
		return
	}
	err = json.Unmarshal(body, &jdata)
	if err != nil {
		log.Println("ERROR unmarshalling request body:", err)
		sendJsonRsp(w, http.StatusBadRequest, "error decoding JSON")
		return
	}

	if jdata.Xname == "" {
		log.Printf("ERROR, request JSON has empty Xname field.\n")
		sendJsonRsp(w, http.StatusBadRequest, "missing required Xname field")
		return
	}
	if !xnametypes.IsHMSCompIDValid(jdata.Xname) {
		log.Printf("ERROR, request JSON has invalid Xname field: '%s'.\n",
			jdata.Xname)
		sendJsonRsp(w, http.StatusBadRequest, "invalid Xname field")
		return
	}
	if xnametypes.GetHMSCompParent(jdata.Xname) == "" {
		log.Printf("ERROR, could not determine Parent for xname: %s.\n", jdata.Xname)
		sendJsonRsp(w, http.StatusInternalServerError, "Could not determine Parent for xname")
		return
	}

	if jdata.Class == "" {
		log.Printf("ERROR, request JSON has empty Class field.\n")
		sendJsonRsp(w, http.StatusBadRequest, "missing required Class field")
		return
	}
	if !sls_common.IsCabinetTypeValid(jdata.Class) {
		log.Printf("ERROR, request JSON has invalid Class field: '%s'.\n",
			string(jdata.Class))
		sendJsonRsp(w, http.StatusBadRequest, "invalid Class field")
		return
	}

	// Check if the component already exists.  If so, it's an error.

	cname, cerr := datastore.GetXname(jdata.Xname)
	if cerr != nil {
		log.Printf("ERROR accessing DB for '%s': %v", jdata.Xname, cerr)
		sendJsonRsp(w, http.StatusInternalServerError, "DB lookup error")
		return
	}
	if cname != nil {
		log.Printf("ERROR, DB object exists for '%s'.\n", jdata.Xname)
		sendJsonRsp(w, http.StatusConflict, "object already exists")
		return
	}

	// Create GenerricHardware object
	var hw sls_common.GenericHardware
	hw.Xname = xnametypes.NormalizeHMSCompID(jdata.Xname)
	hw.Class = jdata.Class
	hw.ExtraPropertiesRaw = jdata.ExtraPropertiesRaw
	hw.Parent = xnametypes.GetHMSCompParent(hw.Xname)
	hw.TypeString = xnametypes.GetHMSType(hw.Xname)
	hw.Type = sls_common.HMSTypeToHMSStringType(hw.TypeString)

	// Write these into the DB
	err, created := datastore.SetXname(hw.Xname, hw)
	if err != nil {
		log.Printf("ERROR inserting component '%s' into DB: %s\n", jdata.Xname, err)
		sendJsonRsp(w, http.StatusInternalServerError, "error inserting object into DB")
		return
	}

	http_code := http.StatusCreated
	if !created {
		// This is an unlikely race condition where the object was created by something else while
		// this POST was in progress and the POST ended up modifing an existing entry instead of
		// creating a new one. This is not a major concern.
		log.Printf("ERROR object was created while POST was already in progress '%s'\n", jdata.Xname)
		http_code = http.StatusOK
	}

	sendJsonRsp(w, http_code, "inserted new entry")
}

//  /hardware GET API

func doHardwareGet(w http.ResponseWriter, r *http.Request) {
	hwList, err := datastore.GetAllXnameObjects()
	if err != nil {
		log.Println("ERROR getting all /hardware objects from DB:", err)
		sendJsonRsp(w, http.StatusInternalServerError, "failed hardware DB query")
		return
	}
	ba, baerr := json.Marshal(hwList)
	if baerr != nil {
		log.Println("ERROR: JSON marshal of /hardware failed:", baerr)
		sendJsonRsp(w, http.StatusInternalServerError, "JSON marshal error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(ba)
}

//  /hardware/{xname} GET API

func doHardwareObjGet(w http.ResponseWriter, r *http.Request) {
	// Decode the URL to get the XName
	vars := mux.Vars(r)
	xname := xnametypes.NormalizeHMSCompID(vars["xname"])

	if !xnametypes.IsHMSCompIDValid(xname) {
		log.Printf("ERROR, invalid xname in request URL: '%s'\n", xname)
		sendJsonRsp(w, http.StatusBadRequest, "invalid xname")
		return
	}

	// Fetch the item and all of its descendants from the database.  If
	// the item does not exist, error.

	cmp, err := datastore.GetXname(xname)
	if cmp == nil {
		log.Printf("ERROR, requested component not found in DB: '%s'\n",
			xname)
		sendJsonRsp(w, http.StatusNotFound, "no such component not in DB")
		return
	}
	if err != nil {
		log.Println("ERROR, DB query failed:", err)
		sendJsonRsp(w, http.StatusInternalServerError, "failed to query DB")
		return
	}

	// Return the HW component.

	sendJsonCompRsp(w, *cmp, http.StatusOK)
}

//  /hardware/{xname} PUT API

func doHardwareObjPut(w http.ResponseWriter, r *http.Request) {
	// Decode the URL to get the XName

	vars := mux.Vars(r)
	xname := xnametypes.NormalizeHMSCompID(vars["xname"])

	if !xnametypes.IsHMSCompIDValid(xname) {
		log.Printf("ERROR, PUT request with xname: '%s'\n", xname)
		sendJsonRsp(w, http.StatusBadRequest, "invalid xname")
		return
	}

	//Unmarshal the payload.

	body, berr := ioutil.ReadAll(r.Body)
	if berr != nil {
		log.Println("ERROR reading request body:", berr)
		sendJsonRsp(w, http.StatusInternalServerError, "unable to read request body")
		return
	}

	var jdata putHardwareRequest
	berr = json.Unmarshal(body, &jdata)
	if berr != nil {
		log.Println("ERROR unmarshalling request body:", berr)
		sendJsonRsp(w, http.StatusBadRequest, "unable to unmarshal JSON payload")
		return
	}

	// Insure mandatory fields are present

	// currently Class is the only manditory field
	errstr := ""
	if jdata.Class == "" {
		errstr = errstr + "Class "
	}
	if errstr != "" {
		errstr = "missing fields " + errstr
		log.Printf("ERROR in PUT JSON: '%s'\n", errstr)
		sendJsonRsp(w, http.StatusBadRequest, errstr)
		return
	}

	if !sls_common.IsCabinetTypeValid(jdata.Class) {
		log.Printf("ERROR, invalid component class '%s'\n",
			string(jdata.Class))
		sendJsonRsp(w, http.StatusBadRequest, "invalid component class")
		return
	}

	// Create GenericHardware
	var hw sls_common.GenericHardware
	hw.Xname = xname
	hw.Class = jdata.Class
	hw.ExtraPropertiesRaw = jdata.ExtraPropertiesRaw
	hw.Parent = xnametypes.GetHMSCompParent(xname)
	hw.TypeString = xnametypes.GetHMSType(xname)
	hw.Type = sls_common.HMSTypeToHMSStringType(hw.TypeString)

	// Write back to the DB
	err, created := datastore.SetXname(hw.Xname, hw)
	if err != nil {
		log.Println("ERROR updating DB:", err)
		sendJsonRsp(w, http.StatusInternalServerError, "DB update failed")
		return
	}

	http_code := http.StatusOK
	if created {
		http_code = http.StatusCreated
	}
	sendJsonCompRsp(w, hw, http_code)
}

// Recursive function used to get all components of a component
// tree and put them into a linear slice.

func getCompTree(gcomp sls_common.GenericHardware, compList *[]sls_common.GenericHardware) error {
	for _, cxname := range gcomp.Children {
		cmp, err := datastore.GetXname(cxname)
		if cmp == nil {
			log.Printf("WARNING: child component '%s' not found in DB.\n",
				cxname)
			continue
		}
		if err != nil {
			return err
		}
		err = getCompTree(*cmp, compList)
		if err != nil {
			return err
		}
	}

	*compList = append(*compList, gcomp)
	return nil
}

//  /hardware/{xname} DELETE API

func doHardwareObjDelete(w http.ResponseWriter, r *http.Request) {
	var compList []sls_common.GenericHardware

	// Decode the URL to get the XName
	vars := mux.Vars(r)
	xname := xnametypes.NormalizeHMSCompID(vars["xname"])

	if !xnametypes.IsHMSCompIDValid(xname) {
		log.Printf("ERROR, invalid Xname in request URL: '%s'\n", xname)
		sendJsonRsp(w, http.StatusBadRequest, "invalid xname")
		return
	}

	// Fetch the item and all of its descendants from the database.  If
	// the item does not exist, error.

	cmp, err := datastore.GetXname(xname)
	if err != nil {
		log.Println("ERROR, error in DB query:", err)
		sendJsonRsp(w, http.StatusInternalServerError, "failed to query DB")
		return
	}
	if cmp == nil {
		log.Printf("ERROR, no '%s' component in DB.\n", xname)
		sendJsonRsp(w, http.StatusNotFound, "no such component not in DB")
		return
	}

	err = getCompTree(*cmp, &compList)
	if err != nil {
		log.Println("ERROR, error in comp tree DB query:", err)
		sendJsonRsp(w, http.StatusInternalServerError, "failed to query DB")
		return
	}

	// Delete the item(s) from the database

	ok := true
	for _, component := range compList {
		log.Printf("INFO: Deleting: '%s'\n", component.Xname)
		err = datastore.DeleteXname(component.Xname)
		if err != nil {
			ok = false
		}
	}

	if !ok {
		sendJsonRsp(w, http.StatusInternalServerError, "failed to delete entry in DB")
		return
	}

	sendJsonRsp(w, http.StatusOK, "deleted entry and its descendants")
}

//  /search/hardware GET API

func doHardwareSearch(w http.ResponseWriter, r *http.Request) {
	hardware := sls_common.GenericHardware{
		Parent:             r.FormValue("parent"),
		Children:           nil,
		Xname:              r.FormValue("xname"),
		Type:               sls_common.HMSStringType(r.FormValue("type")),
		Class:              sls_common.CabinetType(r.FormValue("class")),
		ExtraPropertiesRaw: nil,
	}

	// Build up the extra properties section by gathering the various possible query object and adding them.
	properties := make(map[string]interface{})

	powerConnector := r.FormValue("power_connector")
	if powerConnector != "" {
		properties["PoweredBy"] = powerConnector
	}

	object := r.FormValue("object")
	if object != "" {
		properties["Object"] = object
	}

	nodeNics := r.FormValue("node_nics")
	if nodeNics != "" {
		properties["NodeNics"] = []string{nodeNics}
	}

	networks := r.FormValue("networks")
	if networks != "" {
		properties["Networks"] = []string{networks}
	}

	peers := r.FormValue("peers")
	if peers != "" {
		properties["Peers"] = []string{peers}
	}

	// The ExtraProperties section of SLS is probably the most powerful concept it has. To support generic queries
	// WITHOUT having to code in conditions for each possible field, look for everything that begins with:
	//   `extra_properties.`
	// And add each of them to the map for searching.
	for key, value := range r.Form {
		if strings.HasPrefix(key, "extra_properties.") || key == "extra_properties" {
			// What comes after the period is the name of the property.
			keyParts := strings.SplitN(key, ".", 2)
			if len(keyParts) != 2 || keyParts[1] == "" {
				log.Println("ERROR: ExtraProperties search does not include field")
				pdet := base.NewProblemDetails("about: blank",
					"Bad Request",
					"ExtraProperties search did not include the field name. The ExtraProperties query should be of the form: extra_properties.{fieldname}={value}",
					r.URL.Path, http.StatusBadRequest)
				base.SendProblemDetails(w, pdet, 0)
				return
			}

			// Support multiple values if they're provided.
			if len(value) == 1 {
				properties[keyParts[1]] = value[0]
			} else {
				properties[keyParts[1]] = value
			}
		}
	}

	hardware.ExtraPropertiesRaw = properties

	returnedHardware, validationErr, dbErr := datastore.SearchGenericHardware(hardware)
	if dbErr == database.NoSuch {
		log.Println("ERROR: ", dbErr)
		pdet := base.NewProblemDetails("about: blank",
			"Not Found",
			"Hardware not found in DB",
			r.URL.Path, http.StatusNotFound)
		base.SendProblemDetails(w, pdet, 0)
		return
	} else if validationErr != nil {
		log.Println("ERROR: Bad search request for hardware:", validationErr)
		pdet := base.NewProblemDetails("about: blank",
			"Bad Request",
			validationErr.Error(),
			r.URL.Path, http.StatusBadRequest)
		base.SendProblemDetails(w, pdet, 0)
		return
	} else if dbErr != nil {
		log.Println("ERROR: Failed to search for hardware:", dbErr)
		pdet := base.NewProblemDetails("about: blank",
			"Internal Server Error",
			"Failed to search hardware in DB",
			r.URL.Path, http.StatusInternalServerError)
		base.SendProblemDetails(w, pdet, 0)
		return
	}

	ba, err := json.Marshal(returnedHardware)
	if err != nil {
		log.Println("ERROR: JSON marshal of hardware failed:", err)
		pdet := base.NewProblemDetails("about: blank",
			"Internal Server Error",
			"JSON marshal error",
			r.URL.Path, http.StatusInternalServerError)
		base.SendProblemDetails(w, pdet, 0)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(ba)
}
