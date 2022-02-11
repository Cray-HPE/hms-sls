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
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	sls_common "github.com/Cray-HPE/hms-sls/pkg/sls-common"

	//"os"
	"reflect"
)

type nwTestData struct {
	op        string
	setURL    string
	getURL    string
	setString []byte
	getNWData sls_common.Network
	httpCode  int
}

const (
	nwURLBase = "http://localhost:8376/v1"
)

var nwHappyList = sls_common.NetworkArray{sls_common.Network{"HMN", "Hardware Management Network", []string{"10.1.1.0/28", "10.1.2.0/28"}, "ethernet", 0, "2014-07-16 20:55:46 +0000 UTC", nil},
	sls_common.Network{"NMN", "Node Management Network", []string{"10.100.1.0/28", "10.100.2.0/28"}, "ethernet", 0, "2014-07-16 20:55:46 +0000 UTC", nil},
}

var nwPostHappyPayloads = []nwTestData{
	nwTestData{"POST",
		nwURLBase + "/networks",
		nwURLBase + "/networks/HMN",
		json.RawMessage(`{"Name":"HMN","FullName":"Hardware Management Network","IPRanges":["10.1.1.0/28","10.1.2.0/28"],"Type":"ethernet"}`),
		sls_common.Network{"HMN", "Hardware Management Network", []string{"10.1.1.0/28", "10.1.2.0/28"}, "ethernet", 0, "2014-07-16 20:55:46 +0000 UTC", nil},
		http.StatusCreated,
	},
	nwTestData{"POST",
		nwURLBase + "/networks",
		nwURLBase + "/networks/NMN",
		json.RawMessage(`{"Name":"NMN","FullName":"Node Management Network","IPRanges":["10.100.1.0/28","10.100.2.0/28"],"Type":"ethernet"}`),
		sls_common.Network{"NMN", "Node Management Network", []string{"10.100.1.0/28", "10.100.2.0/28"}, "ethernet", 0, "2014-07-16 20:55:46 +0000 UTC", nil},
		http.StatusCreated,
	},
}

var nwPutNewPayload = nwTestData{"PUT",
	nwURLBase + "/networks/HSN",
	nwURLBase + "/networks/HSN",
	json.RawMessage(`{"Name":"HSN","FullName":"High Speed Network","IPRanges":["192.168.1.0/28","192.168.2.0/28"],"Type":"slingshot10"}`),
	sls_common.Network{"HSN", "High Speed Network", []string{"192.168.1.0/28", "192.168.2.0/28"}, "slingshot10", 0, "2014-07-16 20:55:46 +0000 UTC", nil},
	http.StatusCreated,
}

var nwPostErrPayloads = []nwTestData{
	nwTestData{"POST", //POST of existing network
		nwURLBase + "/networks",
		nwURLBase + "/networks/NMN",
		json.RawMessage(`{"Name":"NMN","FullName":"Node Management Network","IPRanges":["10.100.1.0/28","10.100.2.0/28"],"Type":"ethernet"}`),
		sls_common.Network{"NMN", "Node Management Network", []string{"10.100.1.0/28", "10.100.2.0/28"}, "ethernet", 0, "2014-07-16 20:55:46 +0000 UTC", nil},
		http.StatusConflict,
	},
	nwTestData{"POST", //POST with bad JSON
		nwURLBase + "/networks",
		nwURLBase + "/networks",
		json.RawMessage(`{"Name","HMN","FullName":"Hardware Management Network","IPRanges":["10.1.1.0/28","10.1.2.0/28"],"Type":"ethernet"}`),
		sls_common.Network{"HMN", "Hardware Management Network", []string{"10.1.1.0/28", "10.1.2.0/28"}, "ethernet", 0, "2014-07-16 20:55:46 +0000 UTC", nil},
		http.StatusBadRequest,
	},
	nwTestData{"POST", //POST bad range in IPRanges: 1010.100.2.0/28
		nwURLBase + "/networks",
		nwURLBase + "/networks",
		json.RawMessage(`{"Name":"NMN2","FullName":"Node Management Network","IPRanges":["10.100.1.0/28","1010.100.2.0/28"],"Type":"ethernet"}`),
		sls_common.Network{"NMN2", "Node Management Network", []string{"10.100.1.0/28", "1010.100.2.0/28"}, "ethernet", 0, "2014-07-16 20:55:46 +0000 UTC", nil},
		http.StatusBadRequest,
	},
	nwTestData{"POST", //POST empty range in IPRanges.
		nwURLBase + "/networks",
		nwURLBase + "/networks",
		json.RawMessage(`{"Name":"NMN2","FullName":"Node Management Network","IPRanges":["","10.100.2.0/28"],"Type":"ethernet"}`),
		sls_common.Network{"NMN2", "Node Management Network", []string{"", "10.100.2.0/28"}, "ethernet", 0, "2014-07-16 20:55:46 +0000 UTC", nil},
		http.StatusBadRequest,
	},
}

var nwGetErrPayloads = []nwTestData{
	nwTestData{"GET", //No such network
		nwURLBase + "/networks",
		nwURLBase + "/networks/ZZZ",
		json.RawMessage(`{"Name":"HMN","FullName":"Hardware Management Network","IPRanges":["10.1.1.0/28","10.1.2.0/28"],"Type":"ethernet"}`),
		sls_common.Network{"HMN", "Hardware Management Network", []string{"10.1.1.0/28", "10.1.2.0/28"}, "ethernet", 0, "2014-07-16 20:55:46 +0000 UTC", nil},
		http.StatusNotFound,
	},
}

var nwPutErrPayloads = []nwTestData{
	nwTestData{"PUT", // bad JSON
		nwURLBase + "/networks/HSN2",
		nwURLBase + "/networks/HSN2",
		json.RawMessage(`{"Name","HSN2","FullName":"High Speed Network","IPRanges":["192.168.10.0/28","192.168.20.0/28"],"Type":"slingshot10"}`),
		sls_common.Network{"HSN2", "High Speed Network", []string{"192.168.10.0/28", "192.168.20.0/28"}, "slingshot10", 0, "2014-07-16 20:55:46 +0000 UTC", nil},
		http.StatusBadRequest,
	},
	nwTestData{"PUT", //POST empty range in IPRanges.
		nwURLBase + "/networks/HSN2",
		nwURLBase + "/networks/HSN2",
		json.RawMessage(`{"Name":"HSN2","FullName":"Node Management Network","IPRanges":["","10.100.2.0/28"],"Type":"ethernet"}`),
		sls_common.Network{"HSN2", "Node Management Network", []string{"", "10.100.2.0/28"}, "ethernet", 0, "2014-07-16 20:55:46 +0000 UTC", nil},
		http.StatusBadRequest,
	},
}

func doNWSet(pl nwTestData) (error, int) {
	preq, preqerr := http.NewRequest(pl.op, pl.setURL,
		bytes.NewBuffer(pl.setString))
	if preqerr != nil {
		return fmt.Errorf("ERROR creating http POST request: %v", preqerr), 0
	}

	pw := httptest.NewRecorder()
	router.ServeHTTP(pw, preq)

	//Check response code

	if (pw.Code != http.StatusOK) && (pw.Code != http.StatusCreated) {
		return fmt.Errorf("ERROR response in %s operation: %d/%s",
				pl.op, pw.Code, http.StatusText(pw.Code)),
			pw.Code
	}

	return nil, pw.Code
}

func doNWListGet() (sls_common.NetworkArray, error) {
	var jdata sls_common.NetworkArray

	greq, greqerr := http.NewRequest("GET", nwURLBase+"/v1/networks", nil)
	if greqerr != nil {
		return jdata, greqerr
	}

	gw := httptest.NewRecorder()
	router.ServeHTTP(gw, greq)

	//Check response code

	if gw.Code != http.StatusOK {
		return jdata, fmt.Errorf("Bad response in GET op: %d/%s",
			gw.Code, http.StatusText(gw.Code))
	}

	jerr := json.Unmarshal(gw.Body.Bytes(), &jdata)
	if jerr != nil {
		return jdata, jerr
	}

	return jdata, nil
}

func doNWObjGet(pl nwTestData) (sls_common.Network, error) {
	var jdata sls_common.Network

	greq, greqerr := http.NewRequest("GET", pl.getURL, nil)
	if greqerr != nil {
		return jdata, greqerr
	}

	gw := httptest.NewRecorder()
	router.ServeHTTP(gw, greq)

	//Check response code

	if gw.Code != http.StatusOK {
		return jdata, fmt.Errorf("Bad response in GET op: %d/%s",
			gw.Code, http.StatusText(gw.Code))
	}

	jerr := json.Unmarshal(gw.Body.Bytes(), &jdata)
	if jerr != nil {
		return jdata, jerr
	}

	return jdata, nil
}

func cleanDB() {
	nwl, err := doNWListGet()
	if err != nil {
		log.Printf("ERROR getting stale NW components, probably OK.\n")
		return
	}

	//Delete each item from the list

	for _, item := range nwl {
		var pld = nwTestData{op: "DELETE", setURL: nwURLBase + "/networks" + "/" + item.Name}
		log.Printf("cleanDB(): deleting: '%s'\n", item.Name)
		serr, _ := doNWSet(pld)
		if serr != nil {
			log.Printf("ERROR deleting: '%s'\n", item.Name)
		} else {
			log.Printf("SUCCESS deleting: '%s'\n", item.Name)
		}
	}
}

func logNWTestContext(t *testing.T, description string, data nwTestData) {
	t.Logf(
		"Test context for: %s\n"+
			"    setUrl: %s\n"+
			"    getUrl: %s\n"+
			"    setString: %s\n"+
			"    NWData: %v\n",
		description, data.setURL, data.getURL, data.setString, data.getNWData)
}

func Test_doNetworkIO(t *testing.T) {
	if router == nil {
		routes = generateRoutes()
		router = newRouter(routes)
	}
	dbInit()

	//Clean out whatever is there first

	cleanDB()

	for ii, pl := range nwPostHappyPayloads {
		t.Logf("POST /networks happy test %d, inserting '%s'...", ii, pl.getNWData.Name)

		//Set up and execute the POST
		psterr, code := doNWSet(pl)
		if psterr != nil {
			t.Errorf("ERROR in POST /networks happy test %d: %v", ii, psterr)
			logNWTestContext(t, "POST /netorks", pl)
		} else if code != pl.httpCode {
			log.Printf("ERROR in POST /networks expected http code %d but got %d", pl.httpCode, code)
			logNWTestContext(t, "POST /netorks", pl)
		}

		//Set up and execute the GET

		jdata, gterr := doNWObjGet(pl)
		if gterr != nil {
			t.Errorf("ERROR in POST /networks happy test %d GET op: %v", ii, gterr)
		}

		// Force the LastUpdated property to be identical; we dont care about it doing comparisons.
		jdata.LastUpdated = pl.getNWData.LastUpdated
		jdata.LastUpdatedTime = pl.getNWData.LastUpdatedTime

		if !reflect.DeepEqual(pl.getNWData, jdata) {
			t.Errorf("Data miscompare in POST /happy networks test %d\nexp: %v\ngot: %v",
				ii, pl.getNWData, jdata)
		}
	}

	for ii, pl := range nwPostErrPayloads {
		t.Logf("POST /networks error test %d...", ii)

		//Set up and execute the POST
		psterr, code := doNWSet(pl)
		if psterr == nil {
			t.Errorf("ERROR, POST /networks error test %d didn't fail!", ii)
			logNWTestContext(t, "POST /networks error test", pl)
		} else if code != pl.httpCode {
			log.Printf("ERROR in POST /networks expected http code %d but got %d", pl.httpCode, code)
			logNWTestContext(t, "POST /netorks", pl)
		}
	}

	for ii, pl := range nwGetErrPayloads {
		t.Logf("GET /networks error test %d...", ii)

		//Set up and execute the PUT
		_, psterr := doNWObjGet(pl)
		if psterr == nil {
			t.Errorf("ERROR, GET /networks error test %d didn't fail!", ii)
		}
	}

	//This will just PUT the same stuff that POST already did.

	t.Logf("PUT /networks happy test ...")

	//Set up and execute the POST
	pltmp := nwPostHappyPayloads[0]
	pltmp.op = "PUT"
	pltmp.setURL = pltmp.setURL + "/" + pltmp.getNWData.Name
	psterr, _ := doNWSet(pltmp)
	if psterr != nil {
		t.Errorf("ERROR in PUT /networks happy test %v", psterr)
	}

	//Set up and execute the GET

	jdata, gterr := doNWObjGet(pltmp)
	if gterr != nil {
		t.Errorf("ERROR in PUT /networks happy test GET op: %v", gterr)
	}

	// Force the LastUpdated property to be identical; we dont care about it doing comparisons.
	jdata.LastUpdated = pltmp.getNWData.LastUpdated
	jdata.LastUpdatedTime = pltmp.getNWData.LastUpdatedTime

	if !reflect.DeepEqual(pltmp.getNWData, jdata) {
		t.Errorf("Data miscompare in POST /happy networks test\nexp: %v\ngot: %v",
			pltmp.getNWData, jdata)
	}

	//DELETE a network, insure it is deleted.

	pltmp = nwPostHappyPayloads[0]
	t.Logf("DELETE /networks happy test, deleting: '%s'", pltmp.getNWData.Name)
	pltmp.op = "DELETE"
	pltmp.setURL = pltmp.setURL + "/" + pltmp.getNWData.Name
	psterr, _ = doNWSet(pltmp)
	if psterr != nil {
		t.Errorf("ERROR in DELETE /networks happy test: %v", psterr)
	}

	//Set up and execute the GET

	jdata, gterr = doNWObjGet(pltmp)
	if gterr == nil {
		t.Errorf("ERROR NW object '%s' was not deleted!", pltmp.getNWData.Name)
	}

	//Delete the same network -- should fail, since it's gone

	t.Logf("DELETE of deleted network '%s', (expected to fail)", pltmp.getNWData.Name)
	psterr, _ = doNWSet(pltmp)
	if psterr == nil {
		t.Errorf("ERROR: Deleting already-deleted NW object '%s' didn't fail!",
			pltmp.getNWData.Name)
	}

	//PUT of a non-existent NW

	t.Logf("PUT of non-existent network '%s'", nwPutNewPayload.getNWData.Name)
	psterr, code := doNWSet(nwPutNewPayload)
	if psterr != nil {
		t.Errorf("ERROR PUT /networks non-existent NW test: %v", psterr)
		logNWTestContext(t, "PUT /netorks", nwPutNewPayload)
	} else if code != nwPutNewPayload.httpCode {
		log.Printf("ERROR in PUT /networks expected http code %d but got %d", nwPutNewPayload.httpCode, code)
		logNWTestContext(t, "PUT /netorks", nwPutNewPayload)
	}

	//PUT error, bad JSON

	for ii, pl := range nwPutErrPayloads {
		t.Logf("PUT /networks/{network} error test %d...", ii)

		//Set up and execute the PUT
		psterr, code := doNWSet(pl)
		if psterr == nil {
			t.Errorf("ERROR, PUT /networks error test %d didn't fail!", ii)
			logNWTestContext(t, "PUT /netorks error test", pl)
		} else if code != pl.httpCode {
			log.Printf("ERROR in PUT /networks expected http code %d but got %d", pl.httpCode, code)
			logNWTestContext(t, "PUT /netorks error test", pl)
		}
	}
}
