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

package main

import (
	"context"
	"errors"
	"testing"

	sls_common "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

// SLSClient Mock
type mockSLSClient struct {
	mock.Mock
}

func (m *mockSLSClient) GetAllHardware(ctx context.Context) ([]sls_common.GenericHardware, error) {
	args := m.Called(ctx)

	var allHardware []sls_common.GenericHardware

	if hardware, ok := args.Get(0).([]sls_common.GenericHardware); ok {
		allHardware = hardware // TODO I don't like this naming
	}

	return allHardware, args.Error(1)
}

func (m *mockSLSClient) PutHardware(ctx context.Context, hardware sls_common.GenericHardware) error {
	args := m.Called(ctx, hardware)

	return args.Error(0)
}

// Migrator Test Suite data
var saneHardwareList = []sls_common.GenericHardware{
	// Chassis
	sls_common.NewGenericHardware("x1000c0", sls_common.ClassMountain, nil),
	sls_common.NewGenericHardware("x9000c0", sls_common.ClassHill, nil),
	// ChassisBMC
	sls_common.NewGenericHardware("x1000c0b0", sls_common.ClassMountain, nil),
	sls_common.NewGenericHardware("x9000c0b0", sls_common.ClassHill, nil),
	// GigabyteCMC
	sls_common.NewGenericHardware("x3000c0s1b999", sls_common.ClassRiver, nil),
	// NodeBMC
	sls_common.NewGenericHardware("x1000c0s1b0", sls_common.ClassMountain, nil),
	sls_common.NewGenericHardware("x3000c0s1b0", sls_common.ClassRiver, nil),
	sls_common.NewGenericHardware("x9000c0s1b0", sls_common.ClassHill, nil),
	// Node
	sls_common.NewGenericHardware("x1000c0s1b0n0", sls_common.ClassMountain, nil),
	sls_common.NewGenericHardware("x3000c0s1b0n0", sls_common.ClassRiver, nil),
	sls_common.NewGenericHardware("x9000c0s1b0n0", sls_common.ClassHill, nil),
	// RouterBMC
	sls_common.NewGenericHardware("x1000c0r1b0", sls_common.ClassMountain, nil),
	sls_common.NewGenericHardware("x3000c0r1b0", sls_common.ClassRiver, nil),
	sls_common.NewGenericHardware("x9000c0r1b0", sls_common.ClassHill, nil),
}

var malformedChassis = sls_common.GenericHardware{
	Parent:          "x1003",
	Xname:           "x1003c7",
	Type:            "comptype_chassis_bmc",
	Class:           "Mountain",
	TypeString:      "ChassisBMC",
	LastUpdated:     1655153192,
	LastUpdatedTime: "2022-06-13 20:46:32.184966 +0000 +0000",
}

var malformedGigabyteCMC = sls_common.GenericHardware{
	Parent:          "x3000",
	Xname:           "x3000c0s19b999",
	Type:            "comptype_chassis_bmc",
	Class:           "River",
	TypeString:      "ChassisBMC",
	LastUpdated:     1657806665,
	LastUpdatedTime: "2022-07-14 13:51:05.018704 +0000 +0000",
}

var malformedRouterBMC = sls_common.GenericHardware{
	Parent:          "x3000",
	Xname:           "x3000c0r39b0",
	Type:            "comptype_rtr_bmc",
	Class:           "River",
	TypeString:      "RouterBMC",
	LastUpdated:     1655153192,
	LastUpdatedTime: "2022-06-13 20:46:32.184966 +0000 +0000",
	ExtraPropertiesRaw: map[string]interface{}{
		"Password": "vault://hms-creds/x3000c0r39b0",
		"Username": "vault://hms-creds/x3000c0r39b0",
	},
}

// Migrator Test Suite
type MigratorTestSuite struct {
	suite.Suite

	logger *zap.Logger
}

func (suite *MigratorTestSuite) SetupSuite() {
	suite.logger, _ = zap.NewDevelopment()
}

func (suite *MigratorTestSuite) TestMigrateMalformedChassisData_SaneHardware() {
	for _, hardware := range saneHardwareList {
		// Setup Mock SLS client
		slsClient := new(mockSLSClient)

		m := Migrator{
			performChanges: true,
			slsClient:      slsClient,
			logger:         suite.logger,
		}

		err := m.migrateMalformedChassisData(context.TODO(), hardware)
		suite.NoError(err)
		slsClient.AssertNumberOfCalls(suite.T(), "PutHardware", 0)
		slsClient.AssertNumberOfCalls(suite.T(), "GetAllHardware", 0)
	}
}

func (suite *MigratorTestSuite) TestMigrateMalformedChassisData_MalformedChassis() {
	hardware := malformedChassis

	expectedChassisBMC := sls_common.GenericHardware{
		Parent:     "x1003c7",
		Xname:      "x1003c7b0",
		Class:      "Mountain",
		Type:       "comptype_chassis_bmc",
		TypeString: "ChassisBMC",
	}

	// Setup Mock SLS client
	slsClient := new(mockSLSClient)
	slsClient.On("GetAllHardware", mock.Anything).Return(nil, nil)
	slsClient.On("PutHardware", mock.Anything, expectedChassisBMC).Return(nil)

	m := Migrator{
		performChanges: true,
		slsClient:      slsClient,
		logger:         suite.logger,
	}
	err := m.migrateMalformedChassisData(context.TODO(), hardware)
	suite.NoError(err)

	slsClient.AssertNumberOfCalls(suite.T(), "GetAllHardware", 0)
	slsClient.AssertNumberOfCalls(suite.T(), "PutHardware", 1)
	slsClient.AssertCalled(suite.T(), "PutHardware", mock.Anything, expectedChassisBMC)
}

func (suite *MigratorTestSuite) TestMigrateMalformedChassisData_PUTError() {
	// Setup Mock SLS client
	slsClient := new(mockSLSClient)
	slsClient.On("GetAllHardware", mock.Anything).Return(nil, nil)
	slsClient.On("PutHardware", mock.Anything, mock.Anything).Return(errors.New("put error"))

	m := Migrator{
		performChanges: true,
		slsClient:      slsClient,
		logger:         suite.logger,
	}
	err := m.migrateMalformedChassisData(context.TODO(), malformedChassis)
	suite.EqualError(err, "put error")
}

func (suite *MigratorTestSuite) TestMigrateMalformedDerivedFields_SaneHardware() {
	for _, hardware := range saneHardwareList {
		// Setup Mock SLS client
		slsClient := new(mockSLSClient)

		m := Migrator{
			performChanges: true,
			slsClient:      slsClient,
			logger:         suite.logger,
		}

		err := m.migrateMalformedDerivedFields(context.TODO(), hardware)
		suite.NoError(err)
		slsClient.AssertNumberOfCalls(suite.T(), "PutHardware", 0)
		slsClient.AssertNumberOfCalls(suite.T(), "GetAllHardware", 0)
	}
}

func (suite *MigratorTestSuite) TestMigrateMalformedDerivedFields_MalformedChassis() {
	hardware := malformedChassis

	// Setup Mock SLS client
	slsClient := new(mockSLSClient)
	slsClient.On("GetAllHardware", mock.Anything).Return(nil, nil)
	slsClient.On("PutHardware", mock.Anything, hardware).Return(nil)

	m := Migrator{
		performChanges: true,
		slsClient:      slsClient,
		logger:         suite.logger,
	}
	err := m.migrateMalformedDerivedFields(context.TODO(), hardware)
	suite.NoError(err)

	slsClient.AssertNumberOfCalls(suite.T(), "GetAllHardware", 0)
	slsClient.AssertNumberOfCalls(suite.T(), "PutHardware", 1)
	slsClient.AssertCalled(suite.T(), "PutHardware", mock.Anything, hardware)
}

func (suite *MigratorTestSuite) TestMigrateMalformedDerivedFields_MalformedGigabyteCMC() {
	hardware := malformedGigabyteCMC

	// Setup Mock SLS client
	slsClient := new(mockSLSClient)
	slsClient.On("GetAllHardware", mock.Anything).Return(nil, nil)
	slsClient.On("PutHardware", mock.Anything, hardware).Return(nil)

	m := Migrator{
		performChanges: true,
		slsClient:      slsClient,
		logger:         suite.logger,
	}
	err := m.migrateMalformedDerivedFields(context.TODO(), hardware)
	suite.NoError(err)

	slsClient.AssertNumberOfCalls(suite.T(), "GetAllHardware", 0)
	slsClient.AssertNumberOfCalls(suite.T(), "PutHardware", 1)
	slsClient.AssertCalled(suite.T(), "PutHardware", mock.Anything, hardware)
}

func (suite *MigratorTestSuite) TestMigrateMalformedDerivedFields_MalformedRouterBMC() {
	hardware := malformedRouterBMC

	// Setup Mock SLS client
	slsClient := new(mockSLSClient)
	slsClient.On("GetAllHardware", mock.Anything).Return(nil, nil)
	slsClient.On("PutHardware", mock.Anything, hardware).Return(nil)

	m := Migrator{
		performChanges: true,
		slsClient:      slsClient,
		logger:         suite.logger,
	}
	err := m.migrateMalformedDerivedFields(context.TODO(), hardware)
	suite.NoError(err)

	slsClient.AssertNumberOfCalls(suite.T(), "GetAllHardware", 0)
	slsClient.AssertNumberOfCalls(suite.T(), "PutHardware", 1)
	slsClient.AssertCalled(suite.T(), "PutHardware", mock.Anything, hardware)
}

func (suite *MigratorTestSuite) TestMigrateMalformedDerivedFields_PUTError() {
	// Setup Mock SLS client
	slsClient := new(mockSLSClient)
	slsClient.On("GetAllHardware", mock.Anything).Return(nil, nil)
	slsClient.On("PutHardware", mock.Anything, mock.Anything).Return(errors.New("put error"))

	m := Migrator{
		performChanges: true,
		slsClient:      slsClient,
		logger:         suite.logger,
	}
	err := m.migrateMalformedDerivedFields(context.TODO(), malformedChassis)
	suite.EqualError(err, "put error")
}

func (suite *MigratorTestSuite) TestRun_SaneHardware() {
	// Setup Mock SLS client
	slsClient := new(mockSLSClient)
	slsClient.On("GetAllHardware", mock.Anything).Return(saneHardwareList, nil)
	slsClient.On("PutHardware", mock.Anything, mock.Anything).Return(nil)

	m := Migrator{
		performChanges: true,
		slsClient:      slsClient,
		logger:         suite.logger,
	}

	err := m.Run(context.TODO())
	suite.NoError(err)
	slsClient.AssertNumberOfCalls(suite.T(), "GetAllHardware", 1)
	slsClient.AssertNumberOfCalls(suite.T(), "PutHardware", 0) // No modifications should have taken place.
	slsClient.AssertCalled(suite.T(), "GetAllHardware", mock.Anything)
}

func (suite *MigratorTestSuite) TestRun_MalformedHardware() {
	// Add on the 3 malformed hardware objects to this list of sane objects.
	hardwareList := append(saneHardwareList, malformedChassis, malformedGigabyteCMC, malformedRouterBMC)

	// Setup Mock SLS client
	slsClient := new(mockSLSClient)
	slsClient.On("GetAllHardware", mock.Anything).Return(hardwareList, nil)
	slsClient.On("PutHardware", mock.Anything, mock.Anything).Return(nil)

	m := Migrator{
		performChanges: true,
		slsClient:      slsClient,
		logger:         suite.logger,
	}

	err := m.Run(context.TODO())
	suite.NoError(err)
	slsClient.AssertNumberOfCalls(suite.T(), "GetAllHardware", 1)
	slsClient.AssertNumberOfCalls(suite.T(), "PutHardware", 4) // 4 PUTs should have taken place. 3 Modifications + 1 Creation.
	slsClient.AssertCalled(suite.T(), "GetAllHardware", mock.Anything)
}

func (suite *MigratorTestSuite) TestRun_GetAllHardwareError() {
	// Setup Mock SLS client
	slsClient := new(mockSLSClient)
	slsClient.On("GetAllHardware", mock.Anything).Return(nil, errors.New("get error"))
	slsClient.On("PutHardware", mock.Anything, mock.Anything).Return(nil)

	m := Migrator{
		performChanges: true,
		slsClient:      slsClient,
		logger:         suite.logger,
	}

	err := m.Run(context.TODO())
	suite.EqualError(err, "get error")
	slsClient.AssertNumberOfCalls(suite.T(), "GetAllHardware", 1)
	slsClient.AssertNumberOfCalls(suite.T(), "PutHardware", 0) // No modifications should have taken place.
	slsClient.AssertCalled(suite.T(), "GetAllHardware", mock.Anything)
}

func TestMigratorTestSuite(t *testing.T) {
	suite.Run(t, new(MigratorTestSuite))
}
