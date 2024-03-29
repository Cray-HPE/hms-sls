// MIT License
//
// (C) Copyright [2019, 2021-2023] Hewlett Packard Enterprise Development LP
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
	"github.com/Cray-HPE/hms-xname/xnametypes"
	"github.com/lib/pq"

	"github.com/pkg/errors"
)

func InsertGenericHardwareContext(ctx context.Context, hardware sls_common.GenericHardware) error {
	tx, beginErr := DB.BeginTx(ctx, nil)
	if beginErr != nil {
		return errors.Errorf("unable to begin transaction: %s", beginErr)
	}

	err := InsertGenericHardware(tx, hardware)
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

func InsertGenericHardware(tx *sql.Tx, hardware sls_common.GenericHardware) (err error) {
	q := "INSERT INTO \n" +
		"    components (xname, \n" +
		"                parent, \n" +
		"                comp_type, \n" +
		"                comp_class, \n" +
		"                extra_properties, \n" +
		"				 last_updated_version) \n" +
		"VALUES \n" +
		"($1, \n" +
		" $2, \n" +
		" $3, \n" +
		" $4, \n" +
		" $5, \n" +
		" $6) "

	jsonBytes, jsonErr := json.Marshal(hardware.ExtraPropertiesRaw)
	if jsonErr != nil {
		err = errors.Errorf("unable to marshal ExtendedProperties: %s", jsonErr)
		return err
	}

	version, err := IncrementVersion(tx, hardware.Xname)
	if err != nil {
		err = errors.Errorf("insert to version_history failed: %s", err)
		return err
	}

	result, transErr := tx.Exec(q, hardware.Xname, hardware.Parent, hardware.Type, hardware.Class, string(jsonBytes), version)
	if transErr != nil {
		switch transErr.(type) {
		case *pq.Error:
			if transErr.(*pq.Error).Code.Name() == "unique_violation" {
				err = AlreadySuch
				return
			}
		}

		// I bet we're getting back the wrong (or no) ID
		println(transErr)
		err = errors.Errorf("unable to exec transaction: %s", transErr)
		return
	}

	var counter int64
	counter, rowsErr := result.RowsAffected()
	if rowsErr != nil {
		err = errors.Errorf("insert generic component failed: %s", rowsErr)
		return
	}
	if counter < 1 {
		err = NoSuch
		return
	}

	return
}

func DeleteGenericHardwareContext(ctx context.Context, hardware sls_common.GenericHardware) error {
	tx, beginErr := DB.BeginTx(ctx, nil)
	if beginErr != nil {
		return errors.Errorf("unable to begin transaction: %s", beginErr)
	}

	err := DeleteGenericHardware(tx, hardware)
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

func DeleteGenericHardware(tx *sql.Tx, hardware sls_common.GenericHardware) (err error) {
	q := "DELETE \n" +
		"FROM \n" +
		"    components \n" +
		"WHERE \n" +
		"    xname = $1 "

	_, err = IncrementVersion(tx, hardware.Xname)
	if err != nil {
		err = errors.Errorf("insert to version_history failed: %s", err)
		return err
	}

	result, transErr := tx.Exec(q, hardware.Xname)
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

func DeleteGenericHardwareByXnameContext(ctx context.Context, xname string) error {
	tx, beginErr := DB.BeginTx(ctx, nil)
	if beginErr != nil {
		return errors.Errorf("unable to begin transaction: %s", beginErr)
	}

	err := DeleteGenericHardwareByXname(tx, xname)
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

func DeleteGenericHardwareByXname(tx *sql.Tx, xname string) error {
	// check if xname exists
	_, err := GetGenericHardwareFromXname(tx, xnametypes.NormalizeHMSCompID(xname))
	if err != nil {
		return err
	}
	gh := sls_common.GenericHardware{}
	gh.Xname = xnametypes.NormalizeHMSCompID(xname)
	return DeleteGenericHardware(tx, gh)
}

func DeleteAllGenericHardwareContext(ctx context.Context) error {
	tx, beginErr := DB.BeginTx(ctx, nil)
	if beginErr != nil {
		return errors.Errorf("unable to begin transaction: %s", beginErr)
	}

	err := DeleteAllGenericHardware(tx)
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

func DeleteAllGenericHardware(tx *sql.Tx) (err error) {
	q := "TRUNCATE " +
		"    components "

	_, err = IncrementVersion(tx, "delete all hardware")
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

func UpdateGenericHardwareContext(ctx context.Context, hardware sls_common.GenericHardware) error {
	tx, beginErr := DB.BeginTx(ctx, nil)
	if beginErr != nil {
		return errors.Errorf("unable to begin transaction: %s", beginErr)
	}

	err := UpdateGenericHardware(tx, hardware)
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

func UpdateGenericHardware(tx *sql.Tx, hardware sls_common.GenericHardware) (err error) {
	q := "UPDATE components \n" +
		"SET \n" +
		"    parent           = $2, \n" +
		"    comp_type        = $3, \n" +
		"    comp_class       = $4, \n" +
		"    extra_properties = $5, \n" +
		"    last_updated_version = $6 \n" +
		"WHERE \n" +
		"    xname = $1 "

	jsonBytes, jsonErr := json.Marshal(hardware.ExtraPropertiesRaw)
	if jsonErr != nil {
		err = errors.Errorf("unable to marshal ExtendedProperties: %s", jsonErr)
		return
	}

	version, err := IncrementVersion(tx, hardware.Xname)
	if err != nil {
		err = errors.Errorf("insert to version_history failed: %s", err)
		return err
	}

	result, transErr := tx.Exec(q, hardware.Xname, hardware.Parent, hardware.Type, hardware.Class, string(jsonBytes), version)
	if transErr != nil {
		err = errors.Errorf("unable to exec transaction: %s", transErr)
		return
	}

	var counter int64
	counter, rowsErr := result.RowsAffected()
	if rowsErr != nil {
		err = errors.Errorf("update generic component failed: %s", rowsErr)
		return
	}
	if counter < 1 {
		err = NoSuch
		return
	}

	return
}

func SetGenericHardwareContext(ctx context.Context, obj sls_common.GenericHardware) (err error, created bool) {
	tx, beginErr := DB.BeginTx(ctx, nil)
	if beginErr != nil {
		return errors.Errorf("unable to begin transaction: %s", beginErr), false
	}

	err, created = SetGenericHardware(tx, obj)
	if err != nil {
		tx.Rollback()
		return err, false
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return errors.Errorf("unable to commit transaction: %s", commitErr), false
	}

	return err, created
}

func SetGenericHardware(tx *sql.Tx, obj sls_common.GenericHardware) (err error, created bool) {
	created = false

	// check if xname exists
	_, err = GetGenericHardwareFromXname(tx, obj.Xname)
	if err != nil && err != NoSuch {
		return err, created
	} else if err == NoSuch {
		err = InsertGenericHardware(tx, obj)
		created = true
	} else {
		err = UpdateGenericHardware(tx, obj)
	}

	return err, created
}

func GetAllGenericHardwareContext(ctx context.Context) (hardware []sls_common.GenericHardware, err error) {
	tx, beginErr := DB.BeginTx(ctx, nil)
	if beginErr != nil {
		return nil, errors.Errorf("unable to begin transaction: %s", beginErr)
	}

	hardware, err = GetAllGenericHardware(tx)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return nil, errors.Errorf("unable to commit transaction: %s", commitErr)
	}

	return
}

func GetAllGenericHardware(tx *sql.Tx) (hardware []sls_common.GenericHardware, err error) {
	// First, get the base object and all its associated data
	baseQ := "SELECT \n" +
		"    c1.xname,  \n" +
		"    c1.parent, \n" +
		"    c1.comp_type, \n" +
		"    c1.comp_class,  \n" +
		"    version_history.timestamp, \n" +
		"    c1.extra_properties, \n" +
		"    ARRAY_REMOVE(ARRAY_AGG(distinct c2.xname), NULL) as children \n" +
		"FROM \n" +
		"    components c1 \n" +
		"INNER JOIN version_history \n" +
		"    ON c1.last_updated_version = version_history.version \n" +
		"LEFT JOIN components c2 \n" +
		"    ON c1.xname = c2.parent \n" +
		"    GROUP BY c1.xname, c1.parent, c1.comp_type, c1.comp_class, version_history.timestamp, c1.extra_properties"

	baseRows, baseErr := tx.Query(baseQ)
	if baseErr != nil {
		err = errors.Errorf("unable to query generic hardware: %s", baseErr)
		return
	}

	for baseRows.Next() {
		var thisGenericHardware sls_common.GenericHardware
		var lastUpdated time.Time

		var extraPropertiesBytes []byte
		baseErr := baseRows.Scan(&thisGenericHardware.Xname,
			&thisGenericHardware.Parent,
			&thisGenericHardware.Type,
			&thisGenericHardware.Class,
			&lastUpdated,
			&extraPropertiesBytes,
			pq.Array(&thisGenericHardware.Children))
		if baseErr != nil {
			err = errors.Errorf("unable to scan base row: %s", baseErr)
			return
		}

		thisGenericHardware.LastUpdated = lastUpdated.Unix()
		thisGenericHardware.LastUpdatedTime = lastUpdated.String()

		thisGenericHardware.TypeString = sls_common.HMSStringTypeToHMSType(thisGenericHardware.Type)

		unmarshalErr := json.Unmarshal(extraPropertiesBytes, &thisGenericHardware.ExtraPropertiesRaw)
		if unmarshalErr != nil {
			err = errors.Errorf("unable to unmarshal extended properties: %s", unmarshalErr)
			return
		}

		hardware = append(hardware, thisGenericHardware)
	}

	return
}

func GetGenericHardwareFromXnameContext(ctx context.Context, xname string) (hardware sls_common.GenericHardware, err error) {
	tx, beginErr := DB.BeginTx(ctx, nil)
	if beginErr != nil {
		return sls_common.GenericHardware{}, errors.Errorf("unable to begin transaction: %s", beginErr)
	}

	hardware, err = GetGenericHardwareFromXname(tx, xname)
	if err != nil {
		tx.Rollback()
		return sls_common.GenericHardware{}, err
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return sls_common.GenericHardware{}, errors.Errorf("unable to commit transaction: %s", commitErr)
	}

	return
}

func GetGenericHardwareFromXname(tx *sql.Tx, xname string) (hardware sls_common.GenericHardware, err error) {
	// First, get the base object and all its associated data
	baseQ := "SELECT \n" +
		"    c1.xname, \n" +
		"    c1.parent, \n" +
		"    c1.comp_type, \n" +
		"    c1.comp_class, \n" +
		"    version_history.timestamp, \n" +
		"    c1.extra_properties, \n" +
		"    ARRAY_REMOVE(ARRAY_AGG(distinct c2.xname), NULL) as children \n" +
		"FROM \n" +
		"    components c1 \n" +
		"INNER JOIN version_history \n" +
		"    ON c1.last_updated_version = version_history.version \n" +
		"LEFT JOIN components c2 \n" +
		"    ON c1.xname = c2.parent \n" +
		"WHERE \n" +
		"    c1.xname = $1 \n" +
		"GROUP BY c1.xname, c1.parent, c1.comp_type, c1.comp_class, version_history.timestamp, c1.extra_properties"

	baseRow := tx.QueryRow(baseQ, xname)

	var extraPropertiesBytes []byte
	var lastUpdated time.Time
	baseErr := baseRow.Scan(&hardware.Xname,
		&hardware.Parent,
		&hardware.Type,
		&hardware.Class,
		&lastUpdated,
		&extraPropertiesBytes,
		pq.Array(&hardware.Children))
	if baseErr == sql.ErrNoRows {
		err = NoSuch
		return
	} else if baseErr != nil {
		err = errors.Errorf("unable to scan generic hardware row: %s", baseErr)
		return
	}

	hardware.TypeString = sls_common.HMSStringTypeToHMSType(hardware.Type)
	hardware.LastUpdated = lastUpdated.Unix()
	hardware.LastUpdatedTime = lastUpdated.String()

	unmarshalErr := json.Unmarshal(extraPropertiesBytes, &hardware.ExtraPropertiesRaw)
	if unmarshalErr != nil {
		err = errors.Errorf("unable to unmarshal extended properties: %s", unmarshalErr)
		return
	}

	return
}

func GetGenericHardwareForExtraProperties(ctx context.Context, properties map[string]interface{}) (hardware []sls_common.GenericHardware,
	err error) {
	return SearchGenericHardwareContext(ctx, nil, properties)
}

func SearchGenericHardwareContext(ctx context.Context, conditions map[string]string, properties map[string]interface{}) (
	hardware []sls_common.GenericHardware, err error) {
	tx, beginErr := DB.BeginTx(ctx, nil)
	if beginErr != nil {
		return nil, errors.Errorf("unable to begin transaction: %s", beginErr)
	}

	hardware, err = SearchGenericHardware(tx, conditions, properties)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return nil, errors.Errorf("unable to commit transaction: %s", commitErr)
	}

	return
}

func SearchGenericHardware(tx *sql.Tx, conditions map[string]string, properties map[string]interface{}) (
	hardware []sls_common.GenericHardware, err error) {
	if len(conditions) == 0 && len(properties) == 0 {
		err = errors.Errorf("no conditions/properties with which to search")
		return
	}

	q := "SELECT \n" +
		"    c1.xname, \n" +
		"    c1.parent, \n" +
		"    c1.comp_type, \n" +
		"    c1.comp_class, \n" +
		"    version_history.timestamp, \n" +
		"    c1.extra_properties, \n" +
		"    ARRAY_REMOVE(ARRAY_AGG(distinct c2.xname), NULL) as children \n" +
		"FROM \n" +
		"    components c1 \n" +
		"INNER JOIN version_history \n" +
		"    ON c1.last_updated_version = version_history.version \n" +
		"LEFT JOIN components c2 \n" +
		"    ON c1.xname = c2.parent \n" +
		"WHERE \n     "

	// Build the conditions for the regular columns.
	index := 1
	parameters := make([]interface{}, 0)
	for key, value := range conditions {
		if index != 1 {
			q = q + "  AND"
		}

		q = q + fmt.Sprintf(" c1.%s = $%d \n", key, index)
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
			q = q + fmt.Sprintf(" c1.extra_properties ->> $%d = $%d \n", index, index+1)
			parameters = append(parameters, key, valueString)
			index += 2
		} else if valueArray, ok := value.([]string); ok {
			keyIndex := index
			index++
			var paramArray string
			var values []interface{}
			index, values, paramArray = ToParameterArray(index, valueArray)
			q = q + fmt.Sprintf(" c1.extra_properties -> $%d ?| array[%s] \n", keyIndex, paramArray)
			parameters = append(parameters, key)
			parameters = append(parameters, values...)
		}
	}

	q = q + "GROUP BY c1.xname, c1.parent, c1.comp_type, c1.comp_class, version_history.timestamp, c1.extra_properties \n"

	rows, queryErr := tx.Query(q, parameters...)
	if queryErr != nil {
		err = errors.Errorf("unable to query extra properties: %s", queryErr)
		return
	}

	for rows.Next() {
		newGenericHardware := sls_common.GenericHardware{}
		var extraPropertiesBytes []byte
		var lastUpdated time.Time

		scanErr := rows.Scan(&newGenericHardware.Xname,
			&newGenericHardware.Parent,
			&newGenericHardware.Type,
			&newGenericHardware.Class,
			&lastUpdated,
			&extraPropertiesBytes,
			pq.Array(&newGenericHardware.Children))
		if scanErr != nil {
			err = errors.Errorf("unable to scan row: %s", scanErr)
			return
		}

		newGenericHardware.TypeString = sls_common.HMSStringTypeToHMSType(newGenericHardware.Type)
		newGenericHardware.LastUpdated = lastUpdated.Unix()
		newGenericHardware.LastUpdatedTime = lastUpdated.String()

		unmarshalErr := json.Unmarshal(extraPropertiesBytes, &newGenericHardware.ExtraPropertiesRaw)
		if unmarshalErr != nil {
			err = errors.Errorf("unable to unmarshal extended properties: %s", unmarshalErr)
			return
		}

		hardware = append(hardware, newGenericHardware)
	}

	return
}

func ReplaceAllGenericHardwareContext(ctx context.Context, hardware []sls_common.GenericHardware) error {
	tx, beginErr := DB.BeginTx(ctx, nil)
	if beginErr != nil {
		return errors.Errorf("unable to begin transaction: %s", beginErr)
	}

	err := ReplaceAllGenericHardware(tx, hardware)
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

func ReplaceAllGenericHardware(tx *sql.Tx, hardware []sls_common.GenericHardware) (err error) {
	version, err := IncrementVersion(tx, "replaced all components")
	if err != nil {
		err = errors.Errorf("insert to version_history failed: %s", err)
		return err
	}

	// Start by deleting all the components currently there.
	q := "TRUNCATE " +
		"    components "

	_, transErr := tx.Exec(q)
	if transErr != nil {
		err = errors.Errorf("unable to exec transaction: %s", transErr)
		return
	}

	// Now bulk load the passed in hardware into the database using a prepared statement.
	statement, prepareErr := tx.Prepare(pq.CopyIn("components",
		"xname", "parent", "comp_type", "comp_class", "last_updated_version", "extra_properties"))
	if prepareErr != nil {
		err = errors.Errorf("unable to prepare statement: %s", prepareErr)
		return
	}

	for _, component := range hardware {
		jsonBytes, jsonErr := json.Marshal(component.ExtraPropertiesRaw)
		if jsonErr != nil {
			err = errors.Errorf("unable to marshal ExtendedProperties: %s", jsonErr)
			return
		}

		_, execErr := statement.Exec(component.Xname, component.Parent, component.Type, component.Class,
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
