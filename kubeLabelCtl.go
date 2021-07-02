/*
	Ryan Robinson, 2021

	kubeLabelCtl is a lightweight command line tool built using the client-go API that can retreive kubernetes deployments by their labels
	and then set/get some of their attributes. Currently, kubeLabelCtl can set/get deployment scales from labels, get deployment names from
	labels, and get the number of deployments in a given namespace with specified labels.

	TODO: Create label getters/setters, add getting/setting by name func
*/

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	v1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func printMap(m map[string]string) {
	for k, v := range m {
		fmt.Printf("%s: %s\n", k, v)
	}
}
func printArr(arr []string) {
	for _, v := range arr {
		fmt.Println(v)
	}
}
func isSubset(subset map[string]string, set map[string]string) bool {
	for key, val := range subset {
		if set[key] != val {
			return false
		}
	}
	return true
}

/* convStringsToMap takes a string array of format {key=value, key=value ...} and returns a map of format {key: value, key: value, ...} */
func convStringsToMap(strArr []string) (map[string]string, error) {
	strMap := make(map[string]string)
	for _, str := range strArr {
		if len(str) < 3 || !strings.Contains(str, "=") {
			return nil, errors.New("Error: invalid label argument(s)")
		}
		dL := strings.Index(str, "=")
		strMap[str[:dL]] = str[dL+1:]
	}
	return strMap, nil
}

/* kubeCmd is a struct that holds all required arguments to execute a kubeLabelCtl command. */
type kubeCmd struct {
	cmd       string
	labels    map[string]string
	scale     int32
	namespace string
}

/* initClientSet scans for a kubernetes config file in the local '.kube' diretory. If one is found, it uses it to create and return a
   kubernetes.Clientset struct (https://pkg.go.dev/k8s.io/client-go/kubernetes#Clientset) */
func initClientSet() (kubernetes.Clientset, error) {

	//Scaning for kubernetes .config in local .kube directory
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

/* getDeploymentNameWithLabels searches the given namespace for deployments that contain the labels specified in the labels map
   and returns a slice of all their names */
func GetDeploymentNamesWithLabels(labels map[string]string, namespace string) ([]string, error) {
	clientset, err := initClientSet()
	if err != nil {
		return nil, err
	}

	//Gets a list of deployments in the given namespace
	deployments, err := clientset.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	//Loops through all deployments. If the given labels match a deployment's labels, return the deployment name
	names := []string{}
	for _, deps := range deployments.Items {
		if isSubset(labels, deps.GetLabels()) {
			names = append(names, deps.GetName())
		}
	}
	if len(names) == 0 {
		return nil, errors.New("Error: Deployment does not exist")
	}
	return names, nil
}

/* getDeploymentScaleWithLabels finds the deployments in the given namespace with the given labels and then returns
   a map mapping deployment names to their current scales */
func GetDeploymentScalesWithLabels(labels map[string]string, namespace string) (map[string]string, error) {
	clientset, err := initClientSet()
	if err != nil {
		return nil, err
	}
	deploymentNames, err := GetDeploymentNamesWithLabels(labels, namespace)
	if err != nil {
		return nil, err
	}

	//Maps deployment names to their scales
	scales := make(map[string]string)
	for _, n := range deploymentNames {

		//Gets the deployment with the given name in the given namespace
		deploymentScale, err := clientset.AppsV1().Deployments(namespace).GetScale(context.Background(), n, metav1.GetOptions{})

		if err != nil {
			return nil, err
		}
		scales[n] = strconv.Itoa(int(deploymentScale.Spec.Replicas))
	}

	return scales, nil
}

/* setDeploymentScale finds the deployments in the given namespace with the given labels, and then scales them to 'scale.'
   Returns an array of autoscalingv1.Scale structs (https://pkg.go.dev/k8s.io/api/autoscaling/v1#Scale) */
func SetDeploymentScales(labels map[string]string, scale int32, namespace string) ([]*v1.Scale, error) {
	deploymentNames, err := GetDeploymentNamesWithLabels(labels, namespace)
	if err != nil {
		return nil, err
	}
	clientset, err := initClientSet()
	if err != nil {
		return nil, err
	}

	v1scales := []*v1.Scale{}
	for _, n := range deploymentNames {

		//Gets the deployment's autoscalingv1.Scale struct
		deploymentScale, err := clientset.AppsV1().Deployments(namespace).GetScale(context.Background(), n, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		//Updates the autoscalingv1.Scale struct to the new value, updates the deployment scale
		deploymentScalePoiner := *deploymentScale
		deploymentScalePoiner.Spec.Replicas = scale
		v1scale, err := clientset.AppsV1().Deployments(namespace).UpdateScale(context.Background(), n, &deploymentScalePoiner, metav1.UpdateOptions{})
		if err != nil {
			return nil, err
		}
		v1scales = append(v1scales, v1scale)
	}

	return v1scales, nil
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

/* doCommand takes a kubeCmd struct and executes the command it specifies */
func doCommand(args kubeCmd) {
	switch args.cmd {
	case "empty":
		fmt.Println("A lightweight command line tool that can target Kubernetes deployments by their labels and retrieve/modify their attributes. Reference README for arguments.")
	case "getNumWithLabels":
		num, err := GetNumDeploymentsWithLabels(args.labels, args.namespace)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println(num)
	case "getNames":
		names, err := GetDeploymentNamesWithLabels(args.labels, args.namespace)
		if err != nil {
			log.Fatalln(err)
		}
		printArr(names)
	case "getScales":
		scales, err := GetDeploymentScalesWithLabels(args.labels, args.namespace)
		if err != nil {
			log.Fatalln(err)
		}
		printMap(scales)
	case "setScales":
		_, err := SetDeploymentScales(args.labels, args.scale, args.namespace)
		if err != nil {
			log.Fatalln(err)
		}
	case "error":
		log.Fatalln(errors.New("args: cannot read arguments"))
	}
}

/* getCommand takes an array of arguments, usually from os.Args, and returns the command (conventionally the second arg). If there is not
   a second argument, getCommand returns the string "empty" */
func getCommand(osArgs []string) string {
	if len(osArgs) < 2 {
		return "empty"
	}
	return osArgs[1]
}

/* parseArgs parses an array of arguments, usually from os.Args, and returns a kubeCmd struct containing all the relevant arguments */
func parseArgs(osArgs []string) kubeCmd {
	cmd := getCommand(osArgs)
	args := kubeCmd{}
	args.cmd = cmd

	switch cmd {
	case "getNumWithLabels", "getName", "getScale":
		if len(osArgs) < 4 {
			args.cmd = "error"
			break
		}
		labels, err := convStringsToMap(osArgs[2 : len(osArgs)-1])
		if err != nil {
			log.Fatalln(err)
		}
		args.labels = labels
		args.namespace = osArgs[len(osArgs)-1]
		args.scale = -1
	case "setScale":
		if len(osArgs) < 5 {
			args.cmd = "error"
			break
		}
		labels, err := convStringsToMap(osArgs[2 : len(osArgs)-2])
		if err != nil {
			log.Fatalln(err)
		}
		scale, err := strconv.ParseInt(osArgs[len(osArgs)-2], 10, 32)
		if err != nil {
			log.Fatalln(err)
		}
		args.labels = labels
		args.namespace = osArgs[len(osArgs)-1]
		args.scale = int32(scale)
	default:
		args.cmd = "error"
	}

	return args
}

func main() {
	doCommand(parseArgs(os.Args))
}
