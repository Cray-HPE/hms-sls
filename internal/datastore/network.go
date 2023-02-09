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

package datastore

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/Cray-HPE/hms-sls/v2/internal/database"
	sls_common "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"
)

var InvalidNetworkType = errors.New("invalid network type")
var InvalidNetworkName = errors.New("invalid network name")

func verifyNetworkType(networkType sls_common.NetworkType) error {
	networkTypeLower := strings.ToLower(string(networkType))
	if networkTypeLower != sls_common.NetworkTypeCassini.String() &&
		networkTypeLower != sls_common.NetworkTypeEthernet.String() &&
		networkTypeLower != sls_common.NetworkTypeInfiniband.String() &&
		networkTypeLower != sls_common.NetworkTypeMixed.String() &&
		networkTypeLower != sls_common.NetworkTypeOPA.String() &&
		networkTypeLower != sls_common.NetworkTypeSS10.String() &&
		networkTypeLower != sls_common.NetworkTypeSS11.String() {
		return InvalidNetworkType
	}

	return nil
}

func verifyNetworkName(networkName string) error {
	if strings.Contains(networkName, " ") || networkName == "" {
		return InvalidNetworkName
	}

	return nil
}

// Helper function to verify network is of a correct type and name.
func VerifyNetwork(nw sls_common.Network) error {
	typeErr := verifyNetworkType(nw.Type)
	if typeErr != nil {
		return typeErr
	}

	nameErr := verifyNetworkName(nw.Name)
	if nameErr != nil {
		return nameErr
	}

	ipRangeErr := verifyIPRanges(nw.IPRanges)
	if ipRangeErr != nil {
		return ipRangeErr
	}

	return nil
}

// Validates that the IP ranges are in a valid CIDR notation. Examples: 10.254.0.0/17, 10.94.100.0/24
func verifyIPRanges(ipRanges []string) (err error) {
	for _, ipRange := range ipRanges {
		_, _, parseErr := net.ParseCIDR(ipRange)
		if parseErr != nil {
			err = fmt.Errorf("invalid IP Range: %s", ipRange)
			return
		}
	}
	return
}

// GetNetwork returns the network object matching the given name.
func GetNetwork(ctx context.Context, name string) (sls_common.Network, error) {
	return database.GetNetworkForNameContext(ctx, name)
}

// InsertNetwork adds a given network into the database assuming it passes validation.
func InsertNetwork(ctx context.Context, network sls_common.Network) (validationErr error, dbErr error) {
	validationErr = VerifyNetwork(network)
	if validationErr != nil {
		return
	}

	dbErr = database.InsertNetworkContext(ctx, network)

	return
}

// UpdateNetwork updates all of the fields for a given network in the DB *except* for the name which is read-only.
// Therefore, this function does no validation on network name.
func UpdateNetwork(ctx context.Context, network sls_common.Network) error {
	return database.UpdateNetworkContext(ctx, network)
}

// Insert or update a network
func SetNetwork(ctx context.Context, network sls_common.Network) (verificationErr error, dbErr error) {
	verificationErr = VerifyNetwork(network)
	if verificationErr != nil {
		return
	}

	dbErr = database.SetNetworkContext(ctx, network)

	return
}

// DeleteNetwork removes a network from the DB.
func DeleteNetwork(ctx context.Context, networkName string) error {
	return database.DeleteNetworkContext(ctx, networkName)
}

// GetAllNetworks returns all the network objects in the DB.
func GetAllNetworks(ctx context.Context) ([]sls_common.Network, error) {
	return database.GetAllNetworksContext(ctx)
}

func SearchNetworks(ctx context.Context, network sls_common.Network) (networks []sls_common.Network, err error) {
	conditions := make(map[string]string)

	if network.Name != "" {
		err = verifyNetworkName(network.Name)
		if err != nil {
			return
		}

		conditions["name"] = network.Name
	}
	if network.FullName != "" {
		conditions["full_name"] = network.FullName
	}
	if len(network.IPRanges) == 1 && network.IPRanges[0] != "" {
		conditions["ip_ranges"] = network.IPRanges[0]
	}
	if network.Type != "" {
		err = verifyNetworkType(network.Type)
		if err != nil {
			return
		}

		conditions["type"] = string(network.Type)
	}

	propertiesMap, ok := network.ExtraPropertiesRaw.(map[string]interface{})
	if !ok {
		err = InvalidExtraProperties
		return
	}

	networks, err = database.SearchNetworksContext(ctx, conditions, propertiesMap)

	return
}

func ReplaceAllNetworks(ctx context.Context, networks []sls_common.Network) error {
	return database.ReplaceAllNetworksContext(ctx, networks)
}
