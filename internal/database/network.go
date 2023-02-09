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

package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	sls_common "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

func InsertNetworkContext(ctx context.Context, network sls_common.Network) error {
	tx, beginErr := DB.BeginTx(ctx, nil)
	if beginErr != nil {
		return errors.Errorf("unable to begin transaction: %s", beginErr)
	}

	err := InsertNetwork(tx, network)
	if err != nil {
		tx.Rollback()
		return err
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return errors.Errorf("unable to commit transaction: %s", commitErr)
	}

	return nil
}

func InsertNetwork(tx *sql.Tx, network sls_common.Network) (err error) {
	q := "INSERT INTO \n" +
		"    network (name, \n" +
		"             full_name, \n" +
		"             ip_ranges, \n" +
		"             type, \n" +
		"             extra_properties, \n" +
		"             last_updated_version) \n" +
		"VALUES \n" +
		"($1, \n" +
		" $2, \n" +
		" $3, \n" +
		" $4, \n" +
		" $5, \n" +
		" $6) "

	jsonBytes, jsonErr := json.Marshal(network.ExtraPropertiesRaw)
	if jsonErr != nil {
		err = errors.Errorf("unable to marshal ExtendedProperties: %s", jsonErr)
		return
	}

	version, err := IncrementVersion(tx, network.Name)
	if err != nil {
		err = errors.Errorf("insert to version_history failed: %s", err)
		return err
	}

	result, transErr := tx.Exec(q, network.Name, network.FullName, pq.Array(network.IPRanges), network.Type, string(jsonBytes), version)
	if transErr != nil {
		switch transErr.(type) {
		case *pq.Error:
			if transErr.(*pq.Error).Code.Name() == "unique_violation" {
				err = AlreadySuch
				return
			}
		}

		err = errors.Errorf("unable to exec transaction: %s", transErr)
		return
	}

	var counter int64
	counter, rowsErr := result.RowsAffected()
	if rowsErr != nil {
		err = errors.Errorf("insert network failed: %s", rowsErr)
		return
	}
	if counter < 1 {
		err = NoSuch
		return
	}

	return
}

func DeleteNetworkContext(ctx context.Context, networkName string) error {
	tx, beginErr := DB.BeginTx(ctx, nil)
	if beginErr != nil {
		return errors.Errorf("unable to begin transaction: %s", beginErr)

	}

	err := DeleteNetwork(tx, networkName)
	if err != nil {
		tx.Rollback()
		return err
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return errors.Errorf("unable to commit transaction: %s", commitErr)
	}

	return nil
}

func DeleteNetwork(tx *sql.Tx, networkName string) (err error) {
	q := "DELETE \n" +
		"FROM \n" +
		"    network \n" +
		"WHERE \n" +
		"    name = $1 "

	_, err = IncrementVersion(tx, networkName)
	if err != nil {
		err = errors.Errorf("insert to version_history failed: %s", err)
		return err
	}

	result, transErr := tx.Exec(q, networkName)
	if transErr != nil {
		err = errors.Errorf("unable to exec transaction: %s", transErr)
		return
	}

	var counter int64
	counter, rowsErr := result.RowsAffected()
	if rowsErr != nil {
		err = errors.Errorf("delete generic component failed: %s", rowsErr)
		return
	}
	if counter < 1 {
		err = NoSuch
		return
	}

	return
}

func DeleteAllNetworksContext(ctx context.Context) error {
	tx, beginErr := DB.BeginTx(ctx, nil)
	if beginErr != nil {
		return errors.Errorf("unable to begin transaction: %s", beginErr)
	}

	err := DeleteAllNetworks(tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return errors.Errorf("unable to commit transaction: %s", commitErr)
	}

	return nil
}

func DeleteAllNetworks(tx *sql.Tx) (err error) {
	q := "TRUNCATE " +
		"    network "

	_, err = IncrementVersion(tx, "delete all networks")
	if err != nil {
		err = errors.Errorf("insert to version_history failed: %s", err)
		return err
	}

	_, transErr := tx.Exec(q)
	if transErr != nil {
		err = errors.Errorf("unable to exec transaction: %s", transErr)
		return
	}
	return
}

func UpdateNetworkContext(ctx context.Context, network sls_common.Network) error {
	tx, beginErr := DB.BeginTx(ctx, nil)
	if beginErr != nil {
		return errors.Errorf("unable to begin transaction: %s", beginErr)
	}

	err := UpdateNetwork(tx, network)
	if err != nil {
		tx.Rollback()
		return err
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return errors.Errorf("unable to commit transaction: %s", commitErr)
	}

	return nil
}

func UpdateNetwork(tx *sql.Tx, network sls_common.Network) (err error) {
	q := "UPDATE network \n" +
		"SET \n" +
		"    full_name        = $2, \n" +
		"    ip_ranges        = $3, \n" +
		"    type             = $4, \n" +
		"    extra_properties = $5, \n" +
		"    last_updated_version = $6 \n" +
		"WHERE \n" +
		"    name = $1 "

	jsonBytes, jsonErr := json.Marshal(network.ExtraPropertiesRaw)
	if jsonErr != nil {
		err = errors.Errorf("unable to marshal ExtendedProperties: %s", jsonErr)
		return
	}

	version, err := IncrementVersion(tx, network.Name)
	if err != nil {
		err = errors.Errorf("insert to version_history failed: %s", err)
		return err
	}

	result, transErr := tx.Exec(q, network.Name, network.FullName, pq.Array(network.IPRanges), network.Type, string(jsonBytes), version)
	if transErr != nil {
		err = errors.Errorf("unable to exec transaction: %s", transErr)
		return
	}

	var counter int64
	counter, rowsErr := result.RowsAffected()
	if rowsErr != nil {
		err = errors.Errorf("update network failed: %s", rowsErr)
		return
	}
	if counter < 1 {
		err = NoSuch
		return
	}

	return
}

func SetNetworkContext(ctx context.Context, network sls_common.Network) error {
	tx, beginErr := DB.BeginTx(ctx, nil)
	if beginErr != nil {
		return errors.Errorf("unable to begin transaction: %s", beginErr)
	}

	_, err := GetNetworkForName(tx, network.Name)
	if (err != nil) && (err != NoSuch) {
		tx.Rollback()
		return err
	}

	if (err != nil) && (err == NoSuch) {
		err = InsertNetwork(tx, network)
		if err != nil {
			tx.Rollback()
			return err
		}
	} else {
		err = UpdateNetwork(tx, network)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return errors.Errorf("unable to commit transaction: %s", commitErr)
	}

	return nil
}

func GetAllNetworksContext(ctx context.Context) ([]sls_common.Network, error) {
	tx, beginErr := DB.BeginTx(ctx, nil)
	if beginErr != nil {
		return nil, errors.Errorf("unable to begin transaction: %s", beginErr)
	}

	networks, err := GetAllNetworks(tx)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return nil, errors.Errorf("unable to commit transaction: %s", commitErr)
	}

	return networks, nil
}

func GetAllNetworks(tx *sql.Tx) (networks []sls_common.Network, err error) {
	q := "SELECT \n" +
		"    name, \n" +
		"    full_name, \n" +
		"    ip_ranges, \n" +
		"    type, \n" +
		"    timestamp, \n" +
		"    extra_properties \n" +
		"FROM \n" +
		"    network \n" +
		"INNER JOIN \n" +
		"    version_history \n" +
		"ON network.last_updated_version = version_history.version \n"

	rows, rowsErr := tx.Query(q)
	if rowsErr != nil {
		err = errors.Errorf("unable to query network: %s", rowsErr)
		return
	}

	for rows.Next() {
		var thisNetwork sls_common.Network
		var lastUpdated time.Time

		var extraPropertiesBytes []byte
		scanErr := rows.Scan(&thisNetwork.Name,
			&thisNetwork.FullName,
			pq.Array(&thisNetwork.IPRanges),
			&thisNetwork.Type,
			&lastUpdated,
			&extraPropertiesBytes)
		if scanErr != nil {
			err = errors.Errorf("unable to scan network row: %s", scanErr)
			return
		}

		thisNetwork.LastUpdated = lastUpdated.Unix()
		thisNetwork.LastUpdatedTime = lastUpdated.String()

		unmarshalErr := json.Unmarshal(extraPropertiesBytes, &thisNetwork.ExtraPropertiesRaw)
		if unmarshalErr != nil {
			err = errors.Errorf("unable to unmarshal extra properties: %s", unmarshalErr)
			return
		}

		networks = append(networks, thisNetwork)
	}

	return
}

func GetNetworkForNameContext(ctx context.Context, name string) (sls_common.Network, error) {
	tx, beginErr := DB.BeginTx(ctx, nil)
	if beginErr != nil {
		return sls_common.Network{}, errors.Errorf("unable to begin transaction: %s", beginErr)
	}

	network, err := GetNetworkForName(tx, name)
	if err != nil {
		tx.Rollback()
		return sls_common.Network{}, err
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return sls_common.Network{}, errors.Errorf("unable to commit transaction: %s", commitErr)
	}

	return network, nil
}

func GetNetworkForName(tx *sql.Tx, name string) (network sls_common.Network, err error) {
	q := "SELECT \n" +
		"    name, \n" +
		"    full_name, \n" +
		"    ip_ranges, \n" +
		"    type, \n" +
		"    timestamp, \n" +
		"    extra_properties \n" +
		"FROM \n" +
		"    network  \n" +
		"INNER JOIN \n" +
		"    version_history \n" +
		"ON network.last_updated_version = version_history.version \n" +
		"WHERE \n" +
		"    name = $1 "

	row := tx.QueryRow(q, name)

	var extraPropertiesBytes []byte
	var lastUpdated time.Time
	baseErr := row.Scan(&network.Name,
		&network.FullName,
		pq.Array(&network.IPRanges),
		&network.Type,
		&lastUpdated,
		&extraPropertiesBytes)
	network.LastUpdated = lastUpdated.Unix()
	network.LastUpdatedTime = lastUpdated.String()
	if baseErr == sql.ErrNoRows {
		err = NoSuch
	} else if baseErr != nil {
		err = errors.Errorf("unable to scan network row: %s", baseErr)
	} else {
		unmarshalErr := json.Unmarshal(extraPropertiesBytes, &network.ExtraPropertiesRaw)
		if unmarshalErr != nil {
			err = errors.Errorf("unable to unmarshal extra properties: %s", unmarshalErr)
			return
		}
	}

	return
}

func GetNetworksContainingIPContext(ctx context.Context, addr string) ([]sls_common.Network, error) {
	tx, beginErr := DB.BeginTx(ctx, nil)
	if beginErr != nil {
		return nil, errors.Errorf("unable to begin transaction: %s", beginErr)
	}

	networks, err := GetNetworksContainingIP(tx, addr)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return nil, errors.Errorf("unable to commit transaction: %s", commitErr)
	}

	return networks, nil
}

func GetNetworksContainingIP(tx *sql.Tx, addr string) (networks []sls_common.Network, err error) {
	return SearchNetworks(tx,
		map[string]string{
			"ip_ranges": addr,
		}, map[string]interface{}{},
	)
}

func SearchNetworksContext(ctx context.Context, conditions map[string]string, properties map[string]interface{}) ([]sls_common.Network, error) {
	tx, beginErr := DB.BeginTx(ctx, nil)
	if beginErr != nil {
		return nil, errors.Errorf("unable to begin transaction: %s", beginErr)
	}

	networks, err := SearchNetworks(tx, conditions, properties)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return nil, errors.Errorf("unable to commit transaction: %s", commitErr)
	}

	return networks, nil
}

func SearchNetworks(tx *sql.Tx, conditions map[string]string, properties map[string]interface{}) (networks []sls_common.Network, err error) {
	if len(conditions) == 0 && len(properties) == 0 {
		err = errors.Errorf("no properties with which to search")
		return
	}

	q := "SELECT \n" +
		"    name, \n" +
		"    full_name, \n" +
		"    ip_ranges, \n" +
		"    type, \n" +
		"    timestamp, \n" +
		"    extra_properties \n" +
		"FROM \n" +
		"    network  \n" +
		"INNER JOIN \n" +
		"    version_history \n" +
		"ON network.last_updated_version = version_history.version \n" +
		"WHERE \n     "

	// Now build up the WHERE clause with the given conditions.
	index := 1
	parameters := make([]interface{}, 0)
	for key, value := range conditions {
		if index != 1 {
			q = q + "  AND"
		}

		// postgres operator
		//   inet <<= inet returns boolean - returns true if the subnet contains or equals the other subnet
		if key == "ip_ranges" {
			q = q + fmt.Sprintf(" $%d <<= ANY(ip_ranges) \n", index)
		} else {
			q = q + fmt.Sprintf(" %s = $%d \n", key, index)
		}
		parameters = append(parameters, value)
		index++
	}

	// Build the conditions for the extra properties JSON column.
	for key, value := range properties {
		if index != 1 {
			q = q + "  AND"
		}

		// Some day I want to come back around and make this work with infinite levels of depth, but for now just
		// investigate the type of the value interface. If it's a string then use one syntax, if it's an array use
		// another. The rational being that nested types need different syntax to query.
		//
		// postgres operators: https://www.postgresql.org/docs/current/functions-json.html
		//   jsonb -> text returns jsonb - finds the given json field and returns the value as json
		//   jsonb ->> text returns text - finds the given json field and returns the value as text
		//   jsonb ?| text[] returns boolean - returns true if any of the strings in text[] exist as top level keys or array elements
		valueString, ok := value.(string)
		if ok {
			q = q + fmt.Sprintf(" extra_properties ->> $%d = $%d \n", index, index+1)
			parameters = append(parameters, key, valueString)
			index += 2
		} else if valueArray, ok := value.([]string); ok {
			keyIndex := index
			index++
			var paramArray string
			var values []interface{}
			index, values, paramArray = ToParameterArray(index, valueArray)
			q = q + fmt.Sprintf(" extra_properties -> $%d ?| array[%s] \n", keyIndex, paramArray)
			parameters = append(parameters, key)
			parameters = append(parameters, values...)
		} else {
			err = fmt.Errorf("Unable to query on parameter %s: %v", key, value)
			return
		}
	}

	rows, rowsErr := tx.Query(q, parameters...)
	if rowsErr != nil {
		err = errors.Errorf("unable to query network: %s", rowsErr)
		return
	}

	for rows.Next() {
		var thisNetwork sls_common.Network
		var lastUpdated time.Time

		var extraPropertiesBytes []byte
		scanErr := rows.Scan(&thisNetwork.Name,
			&thisNetwork.FullName,
			pq.Array(&thisNetwork.IPRanges),
			&thisNetwork.Type,
			&lastUpdated,
			&extraPropertiesBytes)
		if scanErr != nil {
			err = errors.Errorf("unable to scan network row: %s", scanErr)
			return
		}

		unmarshalErr := json.Unmarshal(extraPropertiesBytes, &thisNetwork.ExtraPropertiesRaw)
		if unmarshalErr != nil {
			err = errors.Errorf("unable to unmarshal extra properties: %s", unmarshalErr)
			return
		}

		thisNetwork.LastUpdated = lastUpdated.Unix()
		thisNetwork.LastUpdatedTime = lastUpdated.String()

		networks = append(networks, thisNetwork)
	}

	if len(networks) == 0 {
		err = NoSuch
	}

	return
}

func ReplaceAllNetworksContext(ctx context.Context, networks []sls_common.Network) error {
	tx, beginErr := DB.BeginTx(ctx, nil)
	if beginErr != nil {
		return errors.Errorf("unable to begin transaction: %s", beginErr)
	}

	err := ReplaceAllNetworks(tx, networks)
	if err != nil {
		tx.Rollback()
		return err
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return errors.Errorf("unable to commit transaction: %s", commitErr)
	}

	return nil
}

func ReplaceAllNetworks(tx *sql.Tx, networks []sls_common.Network) (err error) {
	version, err := IncrementVersion(tx, "replaced all networks")
	if err != nil {
		err = errors.Errorf("insert to version_history failed: %s", err)
		_ = tx.Rollback()
		return err
	}

	// Start by deleting all the networks currently there.
	q := "TRUNCATE " +
		"    network "

	_, transErr := tx.Exec(q)
	if transErr != nil {
		err = errors.Errorf("unable to exec transaction: %s", transErr)
		return
	}

	// Now bulk load the passed in hardware into the database using a prepared statement.
	statement, prepareErr := tx.Prepare(pq.CopyIn("network",
		"name", "full_name", "ip_ranges", "type", "last_updated_version", "extra_properties"))
	if prepareErr != nil {
		err = errors.Errorf("unable to prepare statement: %s", prepareErr)
		return
	}

	for _, network := range networks {
		jsonBytes, jsonErr := json.Marshal(network.ExtraPropertiesRaw)
		if jsonErr != nil {
			err = errors.Errorf("unable to marshal ExtendedProperties: %s", jsonErr)
			return
		}

		_, execErr := statement.Exec(network.Name, network.FullName, pq.Array(network.IPRanges), network.Type,
			version, string(jsonBytes))
		if execErr != nil {
			err = errors.Errorf("unable to exec statement: %s", execErr)
			return
		}
	}

	_, statementErr := statement.Exec()
	if statementErr != nil {
		err = errors.Errorf("unable to exec statement: %s", statementErr)
		return
	}

	statementErr = statement.Close()
	if statementErr != nil {
		err = errors.Errorf("unable to close statement: %s", statementErr)
		return
	}

	return
}
