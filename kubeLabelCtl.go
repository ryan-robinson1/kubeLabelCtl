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
	"strings"

	v1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

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
   Returns an autoscalingv1.Scale struct (https://pkg.go.dev/k8s.io/api/autoscaling/v1#Scale) */
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
	deploymentScale, err := clientset.AppsV1().Deployments(namespace).GetScale(context.Background(), deploymentName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	//Sets the deployment replica number to the value 'scale' and then updates the scale of the deployment
	deploymentScalePoiner := *deploymentScale
	deploymentScalePoiner.Spec.Replicas = scale
	v1scale, err := clientset.AppsV1().Deployments(namespace).UpdateScale(context.Background(), deploymentName, &deploymentScalePoiner, metav1.UpdateOptions{})
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
	deploymentScale, err := clientset.AppsV1().Deployments(namespace).GetScale(context.Background(), deploymentName, metav1.GetOptions{})
	if err != nil {
		return -1, err
	}
	return int(deploymentScale.Spec.Replicas), nil
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
		fmt.Println("A lightweight command line tool that can get Kubernetes deployments by their labels and retrieve/modify their attributes. Reference README for arguments.")
	case "getNumWithLabels":
		num, err := GetNumDeploymentsWithLabels(args.labels, args.namespace)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println(num)
	case "getName":
		name, err := GetDeploymentNameWithLabels(args.labels, args.namespace)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println(name)
	case "getScale":
		scale, err := GetDeploymentScaleWithLabels(args.labels, args.namespace)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println(strconv.Itoa(scale))
	case "setScale":
		_, err := SetDeploymentScale(args.labels, args.scale, args.namespace)
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
