// MIT License
//
// (C) Copyright [2022] Hewlett Packard Enterprise Development LP
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
	"fmt"
	"testing"
)

func TestToParameterArray_Empty(t *testing.T) {
	index := 7
	values := make([]string, 0)
	newIndex, newValues, params := ToParameterArray(index, values)

	if newIndex != index {
		t.Errorf("Unexpected index, actual: %d, expected: %d", newIndex, index)
	}

	if params != "" {
		t.Errorf("Unexpected non empty params, actual: %s, expected: an empty string", params)
	}

	if len(newValues) != 0 {
		t.Errorf("Unexpected values length, actual: %d, expected: %d, values: %v", len(newValues), 0, newValues)
	}
}

func TestToParameterArray_ThreeValues(t *testing.T) {
	index := 7
	values := []string{"first", "second", "third"}
	newIndex, newValues, params := ToParameterArray(index, values)

	expectedIndex := index + 3
	if newIndex != expectedIndex {
		t.Errorf("Unexpected index, actual: %d, expected: %d", newIndex, expectedIndex)
	}

	expectedParams := "$7, $8, $9"
	if params != expectedParams {
		t.Errorf("Unexpected params, actual: '%s', expected: '%s'", params, expectedParams)
	}

	if len(values) != len(newValues) {
		t.Errorf("Mismatched values length, actual: %d, expected: %d, actual values: %v", len(newValues), len(values), newValues)
		return
	}

	for i, newValue := range newValues {
		value := values[i]
		if fmt.Sprintf("%v", newValue) != value {
			t.Errorf("Unexpected value at index %d, actual: '%v', expected: '%s'", i, newValue, value)
		}
	}
}
