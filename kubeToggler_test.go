package main

import (
	"os"
	"reflect"
	"strconv"
	"testing"
)

/*
	In order for integration testing to work, there needs to be a deployment with the name testconnector-connector and labels "expose.name=usmc1", "expose.group=usmc" in the namespace defined below
*/
var namespace = "jicd42dev"

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

//Tests GetDeploymentNamesWithLabels using label for specific connector. Should return that connector name
func TestGetDeploymentNamesWithLabels_ExistingLabels(t *testing.T) {
	usmcLabel := map[string]string{"expose.name": "usmc1"}
	usmcName := []string{"testconnector-connector"}
	nameLocal, err := GetDeploymentNamesWithLabels(usmcLabel, namespace)
	if err != nil || !reflect.DeepEqual(nameLocal, usmcName) {
		t.Errorf("Returned incorrectly names for %v, got: %v, want: %v, error: %v", usmcLabel, nameLocal, usmcName, err)
	}
}

//Tests GetDeploymentNamesWithLabels using labels for specific connector. Should return that connector name
func TestGetDeploymentNamesWithLabels_MultipleExistingLabels(t *testing.T) {
	usmcLabel := map[string]string{"expose.name": "usmc1", "expose.group": "usmc"}
	usmcName := []string{"testconnector-connector"}
	nameLocal, err := GetDeploymentNamesWithLabels(usmcLabel, namespace)
	if err != nil || !reflect.DeepEqual(nameLocal, usmcName) {
		t.Errorf("Returned incorrectly names for %v, got: %v, want: %v, error: %v", usmcLabel, nameLocal, usmcName, err)
	}
}

//Tests GetDeploymentNamesWithLabels using non existing labels. Should return an error
func TestGetDeploymentNamesWithLabels_NonExistingLabels(t *testing.T) {
	nonExistentLabel := map[string]string{"expose.type": "test"}
	nameLocal, err := GetDeploymentNamesWithLabels(nonExistentLabel, namespace)
	if err == nil {
		t.Errorf("Expected error for %v, got: %v, error: %v", nonExistentLabel, nameLocal, err)
	}
}

/*
	Integration test getNames
*/

//Tests GetNames with labels for a specific connector and no name input. Should return that connector name
func TestGetNames_LabelsMap(t *testing.T) {
	usmcLabel := map[string]string{"expose.name": "usmc1"}
	usmcName := []string{"testconnector-connector"}
	nameLocal, err := getNames(usmcLabel, nil, namespace)
	if err != nil || !reflect.DeepEqual(nameLocal, usmcName) {
		t.Errorf("Returned incorrectly names for %v, got: %v, want: %v, error: %v", usmcLabel, nameLocal, usmcName, err)
	}
}

//Tests GetNames with no label input and name input. Should just return the inputed name
func TestGetNames_NamesArray(t *testing.T) {
	usmcName := []string{"testconnector-connector"}
	nameLocal, err := getNames(nil, usmcName, namespace)
	if err != nil || !reflect.DeepEqual(nameLocal, usmcName) {
		t.Errorf("Returned incorrectly names for %v, got: %v, want: %v, error: %v", usmcName, nameLocal, usmcName, err)
	}
}

//Tests GetNames with label and name input. Should ignore the labels and return the inputed name
func TestGetNames_NamesAndLabelsArray(t *testing.T) {
	usmcLabel := map[string]string{"expose.name": "usmc1"}
	usmcName := []string{"testconnector-connector"}
	nameLocal, err := getNames(usmcLabel, usmcName, namespace)
	if err != nil || !reflect.DeepEqual(nameLocal, usmcName) {
		t.Errorf("Returned incorrectly names for %v, got: %v, want: %v, error: %v", usmcName, nameLocal, usmcName, err)
	}
}

//Tests GetNames with labels that don't exist and no name input
func TestGetNames_NonExistentLabelsAndNames(t *testing.T) {
	nonExistentLabel := map[string]string{"expose.type": "test"}
	nameLocal, err := GetDeploymentNamesWithLabels(nonExistentLabel, namespace)
	if err == nil {
		t.Errorf("Expected error for %v, got: %v, error: %v", nonExistentLabel, nameLocal, err)
	}
}

/*
	Integration test getDeploymentScales and setDeploymentScales
*/

//Tests GetDeploymentScale and SetDeploymentScale by using the connector name to set the deployment scale of testconnector-connector
//to scale 3 and getting the new scale to make sure the values match
func TestGetAndSetDeployment_ByName(t *testing.T) {
	testScale := 3
	_, err1 := SetDeploymentScales(nil, []string{"testconnector-connector"}, int32(testScale), namespace)
	out, err2 := GetDeploymentScales(nil, []string{"testconnector-connector"}, namespace)
	outInt, err3 := strconv.ParseInt(out["testconnector-connector"], 10, 64)
	if err1 != nil || err3 != nil || outInt != int64(testScale) {
		t.Errorf("Returned incorrect scale for set input %v, got: %v, setDeploymentScalesError: %v, getDeploymentScalesError: %v, parseReturnError: %v", testScale, outInt, err1, err2, err3)
	}
}

//Tests GetDeploymentScale and SetDeploymentScale by using the connector labels to set the deployment scale of testconnector-connector
//to scale 1 and getting the new scale to make sure the values match
func TestGetAndSetDeployment_ByLabel(t *testing.T) {
	testScale := 1
	_, err1 := SetDeploymentScales(map[string]string{"expose.name": "usmc1"}, nil, int32(testScale), namespace)
	out, err2 := GetDeploymentScales(map[string]string{"expose.name": "usmc1"}, nil, namespace)
	outInt, err3 := strconv.ParseInt(out["testconnector-connector"], 10, 64)
	if err1 != nil || err2 != nil || err3 != nil || outInt != int64(testScale) {
		t.Errorf("Returned incorrect scale for set input %v, got: %v, setDeploymentScalesError: %v, getDeploymentScalesError: %v, parseReturnError: %v", testScale, outInt, err1, err2, err3)
	}
}

//Tests GetDeploymentScale and SetDeploymentScale by requesting a deployment with nil labels and nil names. Should return an error
func TestGetAndSetDeployment_NonExistentNameOrLabel(t *testing.T) {
	testScale := 5
	_, err1 := SetDeploymentScales(nil, nil, int32(testScale), namespace)
	out, err2 := GetDeploymentScales(nil, nil, namespace)
	outInt, err3 := strconv.ParseInt(out["testconnector-connector"], 10, 64)
	if err1 == nil || err2 == nil || err3 == nil {
		t.Errorf("Expected 3 errors for nil input but returned less than 3 for set input %v, got: %v, setDeploymentScalesError: %v, getDeploymentScalesError: %v, parseReturnError: %v", testScale, outInt, err1, err2, err3)
	}
}

/*
	Integration test getNumDeploymentsWithLabels
*/

//Tests GetNumDeploymentsWithLabels by requesting the number of deployments with the label "expose.name:usmc1" which should equal 1
func TestGetNumDeploymentsWithLabels_1(t *testing.T) {
	exOut := 1
	out, err := GetNumDeploymentsWithLabels(map[string]string{"expose.name": "usmc1"}, namespace)
	if err != nil || out != exOut {
		t.Errorf("Return incorrect number of deployments with label 'expose.name: usmc1' or return an error. expected: %v, got: %v, error: %v", exOut, out, err)
	}
}

//Tests GetNumDeploymentsWithLabels by requesting the number of deployments with a label that is not used which should equal 0
func TestGetNumDeploymentsWithLabels_0(t *testing.T) {
	exOut := 0
	out, err := GetNumDeploymentsWithLabels(map[string]string{"expose.type": "test"}, namespace)
	if err != nil || out != exOut {
		t.Errorf("Return incorrect number of deployments with label 'expose.name: usmc1' or return an error. expected: %v, got: %v, error: %v", exOut, out, err)
	}
}
