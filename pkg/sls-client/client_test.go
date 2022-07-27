// MIT License
//
// (C) Copyright 2022 Hewlett Packard Enterprise Development LP
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

package sls_client

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	sls_common "github.com/Cray-HPE/hms-sls/pkg/sls-common"
	"github.com/stretchr/testify/suite"
)

type SLSClientTestSuite struct {
	suite.Suite
}

func (suite *SLSClientTestSuite) TestGetAllHardware() {
	expectedUserAgent := "sls-client"

	allHardware := []sls_common.GenericHardware{
		sls_common.NewGenericHardware("x1000c0", sls_common.ClassMountain, nil),
		sls_common.NewGenericHardware("x1000c0b0", sls_common.ClassMountain, nil),
		sls_common.NewGenericHardware("x1000c0s0b0n0", sls_common.ClassMountain, nil),
		sls_common.NewGenericHardware("x1000c0s0b0n1", sls_common.ClassMountain, nil),
	}

	var requestCount int
	var expectedUserAgentPresent bool

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/hardware" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Update our book keeping
		requestCount++
		expectedUserAgentPresent = r.Header.Get("User-Agent") == expectedUserAgent

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(allHardware)

	}))

	// Create the SLS Client
	slsClient := NewSLSClient(ts.URL, ts.Client(), expectedUserAgent)

	returnedHardware, err := slsClient.GetAllHardware(context.TODO())
	suite.NoError(err)
	suite.Len(returnedHardware, 4)
	suite.Equal(allHardware, returnedHardware)

	suite.Equal(1, requestCount, "Requests made to /v1/hardware")
	suite.True(expectedUserAgentPresent, "User Agent present")
}

func (suite *SLSClientTestSuite) TestPutHardware_Existing() {
	expectedUserAgent := "sls-client"

	expectedHardwareRaw := `{"Parent":"x1000c0","Xname":"x1000c0b0","Type":"comptype_chassis_bmc","Class":"Mountain","TypeString":"ChassisBMC"}`

	var requestCount int
	var expectedUserAgentPresent bool
	var expectedRequestBodyProvided bool

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/hardware/x1000c0b0" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		defer r.Body.Close()

		// Update our book keeping
		requestCount++
		expectedUserAgentPresent = r.Header.Get("User-Agent") == expectedUserAgent
		expectedRequestBodyProvided = expectedHardwareRaw == string(body)

		w.WriteHeader(http.StatusOK)
	}))

	// Create the SLS Client
	slsClient := NewSLSClient(ts.URL, ts.Client(), expectedUserAgent)

	hardware := sls_common.NewGenericHardware("x1000c0b0", sls_common.ClassMountain, nil)
	err := slsClient.PutHardware(context.TODO(), hardware)
	suite.NoError(err)

	suite.Equal(1, requestCount, "Requests made to /v1/hardware/x1000c0b0")
	suite.True(expectedUserAgentPresent, "User Agent present")
	suite.True(expectedRequestBodyProvided, "Expected Request body provided")
}

func (suite *SLSClientTestSuite) TestPutHardware_New() {
	expectedUserAgent := "sls-client"

	expectedHardwareRaw := `{"Parent":"x1000c0","Xname":"x1000c0b0","Type":"comptype_chassis_bmc","Class":"Mountain","TypeString":"ChassisBMC"}`

	var requestCount int
	var expectedUserAgentPresent bool
	var expectedRequestBodyProvided bool

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/hardware/x1000c0b0" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		defer r.Body.Close()

		// Update our book keeping
		requestCount++
		expectedUserAgentPresent = r.Header.Get("User-Agent") == expectedUserAgent
		expectedRequestBodyProvided = expectedHardwareRaw == string(body)

		w.WriteHeader(http.StatusCreated)
	}))

	// Create the SLS Client
	slsClient := NewSLSClient(ts.URL, ts.Client(), expectedUserAgent)

	hardware := sls_common.NewGenericHardware("x1000c0b0", sls_common.ClassMountain, nil)
	err := slsClient.PutHardware(context.TODO(), hardware)
	suite.NoError(err)

	suite.Equal(1, requestCount, "Requests made to /v1/hardware/x1000c0b0")
	suite.True(expectedUserAgentPresent, "User Agent present")
	suite.True(expectedRequestBodyProvided, "Expected Request body provided")
}

func TestSLSClientTestSuite(t *testing.T) {
	suite.Run(t, new(SLSClientTestSuite))
}
