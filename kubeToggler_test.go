package main

import (
	"os"
	"reflect"
	"testing"
)

/*
	Unit test checkMap
*/

//Test for 3 values in map format. Should return true
func TestCheckMap_Map3Vals(t *testing.T) {
	exOut := true
	mapArray := []string{"key1=val1", "key2=val2", "key3=val3"}
	isMap, err := checkMap(mapArray)
	if !isMap || err != nil {
		t.Errorf("Returned incorrect bool or error for %v, got: %t, want: %t, error: %v", mapArray, isMap, exOut, err)
	}
}

//Test for 0 values. Should return an error
func TestCheckMap_0Vals(t *testing.T) {
	mapArray := []string{}
	isMap, err := checkMap(mapArray)
	if err == nil {
		t.Errorf("Expected error for %v, got: %t, error: %v", mapArray, isMap, err)
	}
}

//Test for 3 values in an array format. Should return false
func TestCheckMap_Array(t *testing.T) {
	exOut := false
	mapArray := []string{"key1", "key2", "key3"}
	isMap, err := checkMap(mapArray)
	if isMap || err != nil {
		t.Errorf("Returned incorrect bool or error for %v, got: %t, want: %t, error: %v", mapArray, isMap, exOut, err)
	}
}

//Test for array with mixed formating. Should return an error
func TestCheckMap_Mixed(t *testing.T) {
	mapArray := []string{"key", "key2=val2", "key3"}
	isMap, err := checkMap(mapArray)
	if err == nil {
		t.Errorf("Expected error for %v, got: %t, error: %v", mapArray, isMap, err)
	}
}

//Test for an empty string "argument". Should return an error"
func TestCheckMap_InvalidLabels1(t *testing.T) {
	mapArray := []string{"k="}
	isMap, err := checkMap(mapArray)
	if err == nil {
		t.Errorf("Expected error for %v, got: %t, error: %v", mapArray, isMap, err)
	}
}

//Test for an empty string "argument" with a larger length key. Should return an error"
func TestCheckMap_InvalidLabels2(t *testing.T) {
	mapArray := []string{"key="}
	isMap, err := checkMap(mapArray)
	if err == nil {
		t.Errorf("Expected error for %v, got: %t, error: %v", mapArray, isMap, err)
	}
}

//Test for an empty string argument mixed with a valid map value. Should return an error
func TestCheckMap_InvalidLabels3(t *testing.T) {
	mapArray := []string{"=value", "key=value"}
	isMap, err := checkMap(mapArray)
	if err == nil {
		t.Errorf("Expected error for %v, got: %t, error: %v", mapArray, isMap, err)
	}
}

//Test for two empty string arguments. Should return an error
func TestCheckMap_InvalidLabels4(t *testing.T) {
	mapArray := []string{"="}
	isMap, err := checkMap(mapArray)
	if err == nil {
		t.Errorf("Expected error for %v, got: %t, error: %v", mapArray, isMap, err)
	}
}

/*
	Unit test convStringsToMap
*/

func TestConvStringsToMap_ValidMapInput(t *testing.T) {
	testMap := map[string]string{"foo": "one", "bar": "two"}
	testArr := []string{"foo=one", "bar=two"}
	inputMap, err := convStringsToMap(testArr)
	if !reflect.DeepEqual(testMap, inputMap) || err != nil {
		t.Errorf("Returned incorrectly formatted map or error for %v, got: %v, want: %v, error: %v", testArr, inputMap, testMap, err)
	}
}

//Test for an empty string argument, should return error
func TestConvStringsToMap_InValidMapInput1(t *testing.T) {
	testArr := []string{"foo=one", "bar="}
	inputMap, err := convStringsToMap(testArr)
	if err == nil {
		t.Errorf("Expected error for %v, got: %v, error: %v", testArr, inputMap, err)
	}
}

//Test for an invalid map format, should return error
func TestConvStringsToMap_InValidMapInput2(t *testing.T) {
	testArr := []string{"foo", "bar"}
	inputMap, err := convStringsToMap(testArr)
	if err == nil {
		t.Errorf("Expected error for %v, got: %v, error: %v", testArr, inputMap, err)
	}
}

//Test for an empty input map, should return an empty map
func TestConvStringsToMap_EmptyInput(t *testing.T) {
	testMap := map[string]string{}
	testArr := []string{}
	inputMap, err := convStringsToMap(testArr)
	if !reflect.DeepEqual(testMap, inputMap) || err != nil {
		t.Errorf("Expected error for %v, got: %v, error: %v", testArr, inputMap, err)
	}
}

/*
	Integration test initClientSet
*/
//Tests to confirm there is a .kube/config file
func TestInitClientSet(t *testing.T) {
	_, err := os.Stat(".kube/config")
	if os.IsNotExist(err) {
		t.Errorf("Config file in local .kube directory does not exist")
	}
}

/*
	Integration test GetDeploymentNamesWithLabels
*/
