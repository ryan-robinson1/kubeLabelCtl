/*
	Ryan Robinson, 2021

	kubeLabelCtl is a lightweight command line tool built using the client-go API that can retreive kubernetes deployments by their labels 
	and then set/get some of their attributes. Currently, kubeLabelCtl can set/get deployment scales from labels, get deployment names from 
	labels, and get the number of deployments in a given namespace with specified labels.
*/

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	v1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func isNumeric(s string) bool {
	_, err := strconv.ParseInt(s, 10, 32)
	return err == nil
}
func isSubset(subset map[string]string, set map[string]string) bool {
	for key, val := range subset {
		if set[key] != val {
			return false
		}
	}
	return true
}

/* Takes a string array of format {key, key, value, value, ...} and returns a map of format {key: value, key: value, ...} {key value} */
func convStringsToMap(strArr []string) (map[string]string, error) {
	if len(strArr)%2 != 0 {
		return nil, errors.New("args: label argument has unmatched pair")
	}
	strMap := make(map[string]string)
	keys := strArr[:len(strArr)/2]
	vals := strArr[len(strArr)/2:]

	for i := 0; i < len(strArr)/2; i++ {
		strMap[keys[i]] = vals[i]
	}
	return strMap, nil
}

/* initClientSet scans for a kubernetes config file in the local '.kube' diretory. If one is found, it uses it to create and return a
   kubernetes.Clientset struct (https://pkg.go.dev/k8s.io/client-go/kubernetes#Clientset) */
func initClientSet() (kubernetes.Clientset, error) {

	//Scaning for kubernetes .config in local /.kube directory
	rules := clientcmd.NewDefaultClientConfigLoadingRules()

	//Creates a clientcmd.ClientConfig struct which is used to create a *rest.Config struct
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, &clientcmd.ConfigOverrides{})
	config, err := kubeconfig.ClientConfig()

	if err != nil {
		return kubernetes.Clientset{}, err
	}
	
	//Attempts to create a kubernetes.Clientset struct from 'config,' panics if failure
	return *kubernetes.NewForConfigOrDie(config), nil
}

/* getDeploymentNameWithLabels searches the given namespace for deployments that contain the labels specified in the labels map.
   If a deployment is found that contains the labels speciifed in the map, the function returns the name of the deployment */
func GetDeploymentNameWithLabels(labels map[string]string, namespace string) (string, error) {
	clientset, err := initClientSet()
	if err != nil {
		return "", err
	}

	//Gets a list of deployments in the given namespace
	deployments, err := clientset.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	//Loops through all deployments. If the given labels match a deployment's labels, return the deployment name
	//TODO: Make kubeLabelCtl functions work with multiple deployments at once
	depName := ""
	numDeps := 0
	for _, deps := range deployments.Items {
		if isSubset(labels, deps.GetLabels()) {
			depName = deps.GetName()
			numDeps++
		}
	}

	if numDeps == 0 {
		return "", errors.New("Error: Deployment does not exist")
	} else if numDeps == 1 {
		return depName, nil
	} else {
		return "", errors.New("Error: Multiple deployments with same tag(s)")
	}

}

/* setDeploymentScale finds the deployment in the given namespace with the given labels, and then scales it to 'scale.'
   Returns a v1.Scale struct (https://pkg.go.dev/k8s.io/api/autoscaling/v1#Scale) */
func SetDeploymentScale(labels map[string]string, scale int32, namespace string) (*v1.Scale, error) {
	deploymentName, err := GetDeploymentNameWithLabels(labels, namespace)
	if err != nil {
		return nil, err
	}
	clientset, err := initClientSet()
	if err != nil {
		return nil, err
	}

	//Gets the deployment with the given name in the given namespace
	deployment, err := clientset.AppsV1().Deployments(namespace).GetScale(context.Background(), deploymentName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	//Sets the deployment replica number to the value 'scale' and then updates the scale of the deployment
	deploymentPoiner := *deployment
	deploymentPoiner.Spec.Replicas = scale
	v1scale, err := clientset.AppsV1().Deployments(namespace).UpdateScale(context.Background(), deploymentName, &deploymentPoiner, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	return v1scale, nil
}

/* getDeploymentScaleWithLabels finds the deployment in the given namespace with the given labels and then returns
   its scale as an int */
func GetDeploymentScaleWithLabels(labels map[string]string, namespace string) (int, error) {
	clientset, err := initClientSet()
	if err != nil {
		return -1, err
	}
	deploymentName, err := GetDeploymentNameWithLabels(labels, namespace)
	if err != nil {
		return -1, err
	}

	//Gets the deployment with the given name in the given namespace
	deployment, err := clientset.AppsV1().Deployments(namespace).GetScale(context.Background(), deploymentName, metav1.GetOptions{})
	if err != nil {
		return -1, err
	}
	return int(deployment.Spec.Replicas), nil
}

/* getNumDeploymentsWithLabels returns the count of the number of deployments that contain the given labels in the given namespace */
func GetNumDeploymentsWithLabels(labels map[string]string, namespace string) (int, error) {
	clientset, err := initClientSet()
	if err != nil {
		return -1, err
	}

	//Gets a list of deployments in the given namespace
	deployments, err := clientset.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return -1, err
	}

	//If a deployment contains all labels given, increment the counter
	counter := 0
	for _, deps := range deployments.Items {
		if isSubset(labels, deps.GetLabels()) {
			counter++
		}
	}
	return counter, nil
}

/* Takes the array of command line arguments, parses them, and returns a function that can execute the requested action */
func parseArgs(args []string) (func(), error) {
	if args[0] == "getNumWithLabels" {
		return func() {
			labels, err := convStringsToMap(args[1 : len(args)-1])
			if err != nil {
				log.Fatalln(err)
			}
			num, err := GetNumDeploymentsWithLabels(labels, args[len(args)-1])
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Println(num)
		}, nil
	} else if args[0] == "getName" {
		return func() {
			labels, err := convStringsToMap(args[1 : len(args)-1])
			if err != nil {
				log.Fatalln(err)
			}
			name, err := GetDeploymentNameWithLabels(labels, args[len(args)-1])
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Println(name)
		}, nil
	} else if args[0] == "getScale" {
		return func() {
			labels, err := convStringsToMap(args[1 : len(args)-1])
			if err != nil {
				log.Fatalln(err)
			}
			scale, err := GetDeploymentScaleWithLabels(labels, args[len(args)-1])
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Println(strconv.Itoa(scale))
		}, nil
	} else if args[0] == "setScale" {
		return func() {
			if !isNumeric(args[len(args)-2]) {
				panic(errors.New("args: setScale requires an integer scale value"))
			}
			labels, err := convStringsToMap(args[1 : len(args)-2])
			if err != nil {
				log.Fatalln(err)
			}
			scale, err := strconv.ParseInt(args[len(args)-2], 10, 32)
			if err != nil {
				log.Fatalln(err)
			}
			SetDeploymentScale(labels, int32(scale), args[len(args)-1])

		}, nil
	}
	return nil, errors.New("args: cannot read arguments")
}

func main() {
	args := os.Args[1:]
	action, err := parseArgs(args)
	if err != nil {
		log.Fatalln(err)
	}
	action()
}
