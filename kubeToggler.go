/*
	Ryan Robinson, 2021

	kubeToggler is a lightweight command line tool built using the client-go API that can retreive kubernetes deployments by their labels or names
	and then set/get some of their attributes. Currently, kubeToggler can set/get deployment scales from labels, get deployment names from
	labels, and get the number of deployments in a given namespace with specified labels.

	TODO: Add unit tests and integration tests
*/

package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	v1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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
func getTimeElapsed(timestamp string) (time.Duration, error) {
	pastTime, err := time.Parse("2006-01-02|15:04:05 UTC", timestamp)
	if err != nil {
		return 0, err
	}
	d := time.Now().UTC().Sub(pastTime)
	return d, nil
}

/* checkMap returns true if an array is in this format {key=value, key=value ...} and false otherwise */
func checkMap(strArr []string) (bool, error) {
	dLCounter := 0
	if len(strArr) == 0 {
		return false, errors.New("error: empty array")
	}
	for _, str := range strArr {

		//Makes sure there isnt an empty string on any side of the equals sign
		if strings.Contains(str, "=") && (str[0] == '=' || str[len(str)-1] == '=') {
			return false, errors.New("error: invalid label argument(s)")
		} else if strings.Contains(str, "=") {
			dLCounter++
		} else {
			dLCounter--
		}
	}
	switch dLCounter {
	case len(strArr):
		return true, nil
	case -len(strArr):
		return false, nil
	default:
		return false, errors.New("error: arguments must be either labels or names, not both")
	}
}

/* convStringsToMap takes a string array of format {key=value, key=value ...} and returns a map of format {key: value, key: value, ...} */
func convStringsToMap(strArr []string) (map[string]string, error) {
	strMap := make(map[string]string)
	for _, str := range strArr {
		if str[0] == '=' || str[len(str)-1] == '=' || !strings.Contains(str, "=") {
			return nil, errors.New("error: invalid label argument(s)")
		}
		dL := strings.Index(str, "=")
		strMap[str[:dL]] = str[dL+1:]
	}
	return strMap, nil
}

/* kubeCmd is a struct that holds all required arguments to execute a kubeToggler command. */
type kubeCmd struct {
	cmd       string
	labels    map[string]string
	names     []string
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
		return nil, errors.New("error: deployment does not exist")
	}
	return names, nil
}

/* getNames takes a map of labels and an array of names. If the name argument is nil, getNames uses the labels to fetch each deployment's
   name and returns an array of names. Otherwise, get names just returns the unchanged names argument */
func getNames(labels map[string]string, names []string, namespace string) ([]string, error) {
	if names == nil && labels == nil {
		return nil, errors.New("error: there must be at least one targeting field (either names or labels)")
	}
	if names != nil {
		return names, nil
	}
	deploymentNames, err := GetDeploymentNamesWithLabels(labels, namespace)
	if err != nil {
		return nil, err
	}
	return deploymentNames, nil
}

/* getDeploymentScaleWithLabels finds the deployments in the given namespace with the given labels or names in the
   names array and then returns a map mapping deployment names to their current scales */
func GetDeploymentScales(labels map[string]string, names []string, namespace string) (map[string]string, error) {
	clientset, err := initClientSet()
	if err != nil {
		return nil, err
	}
	deploymentNames, err := getNames(labels, names, namespace)
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

/* setDeploymentScale finds the deployments in the given namespace with the given labels or names and then scales them to 'scale.'
   Returns an array of autoscalingv1.Scale structs (https://pkg.go.dev/k8s.io/api/autoscaling/v1#Scale) */
func SetDeploymentScales(labels map[string]string, names []string, scale int32, namespace string) ([]*v1.Scale, error) {
	deploymentNames, err := getNames(labels, names, namespace)
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

/* GetPods takes the name and namespace of a deployment and returns an array of pods currently running in that deployment */
func getPods(deploymentName string, namespace string) ([]corev1.Pod, error) {
	clientset, err := initClientSet()
	if err != nil {
		return nil, err
	}
	deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.Background(), deploymentName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	labelMap, err := metav1.LabelSelectorAsMap(deployment.Spec.Selector)
	if err != nil {
		return nil, err
	}
	options := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(labelMap).String(),
	}
	podList, err := clientset.CoreV1().Pods(namespace).List(context.Background(), options)
	if err != nil {
		return nil, err
	}

	return podList.Items, nil
}

/* GetPodCreationTimestamps takes the name and namespace of a deployment and returns a map mapping the deployment's
   pods to their time of creation. Time is formatted like 2006-01-02|15:04:05 UTC*/
func GetPodCreationTimestamps(deploymentName string, namespace string) (map[string]string, error) {
	pods, err := getPods(deploymentName, namespace)
	if err != nil {
		return nil, err
	}
	podTimeStamps := make(map[string]string)
	for _, pod := range pods {
		podTimeStamps[pod.Name] = pod.CreationTimestamp.UTC().Format("2006-01-02|15:04:05 UTC")
	}
	return podTimeStamps, nil
}

/* GetPodCreationTimestamps takes the name and namespace of a deployment and returns a map mapping the deployment's
   pods to their time of creation. Time is formatted like 2006-01-02|15:04:05 UTC*/
func GetPodLifetimes(deploymentName string, namespace string) (map[string]string, error) {
	podTimestamps, err := GetPodCreationTimestamps(deploymentName, namespace)
	if err != nil {
		return nil, err
	}
	podLifetimes := make(map[string]string)
	for name, timestamp := range podTimestamps {
		duration, err := getTimeElapsed(timestamp)
		if err != nil {
			return nil, err
		}
		podLifetimes[name] = duration.String()
	}
	return podLifetimes, nil
}

/* GetPodCreationTimestamps takes the name and namespace of a deployment and returns a map mapping the deployment's
   pods to their logs */
func GetPodLogs(deploymentName string, namespace string) (map[string]string, error) {
	clientset, err := initClientSet()
	if err != nil {
		return nil, err
	}
	pods, err := getPods(deploymentName, namespace)

	podLogOpts := corev1.PodLogOptions{}

	if err != nil {
		return nil, err
	}
	podLogs := make(map[string]string)
	for _, pod := range pods {
		logs := clientset.CoreV1().Pods(namespace).GetLogs(pod.Name, &podLogOpts)
		req, err := logs.Stream(context.Background())
		if err != nil {
			return nil, err
		}
		defer req.Close()

		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, req)
		if err != nil {
			return nil, err
		}
		logStr := buf.String()

		podLogs[pod.Name] = logStr
	}
	return podLogs, nil
}

/* doCommand takes a kubeCmd struct and executes the command it specifies */
func doCommand(args kubeCmd) {
	switch args.cmd {
	case "empty":
		fmt.Println("A lightweight command line tool that can target Kubernetes deployments by their labels and retrieve/modify their attributes. Reference README for arguments.")
	case "getNumWithLabel":
		num, err := GetNumDeploymentsWithLabels(args.labels, args.namespace)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println(num)
	case "getName":
		names, err := GetDeploymentNamesWithLabels(args.labels, args.namespace)
		if err != nil {
			log.Fatalln(err)
		}
		printArr(names)
	case "getScale":
		scales, err := GetDeploymentScales(args.labels, args.names, args.namespace)
		if err != nil {
			log.Fatalln(err)
		}
		printMap(scales)
	case "setScale":
		_, err := SetDeploymentScales(args.labels, args.names, args.scale, args.namespace)
		if err != nil {
			log.Fatalln(err)
		}
	case "toggleOn":
		_, err := SetDeploymentScales(args.labels, args.names, 1, args.namespace)
		if err != nil {
			log.Fatalln(err)
		}
	case "toggleOff":
		_, err := SetDeploymentScales(args.labels, args.names, 0, args.namespace)
		if err != nil {
			log.Fatalln(err)
		}
	case "reset":
		_, err := SetDeploymentScales(args.labels, args.names, 0, args.namespace)
		if err != nil {
			log.Fatalln(err)
		}
		_, err = SetDeploymentScales(args.labels, args.names, 1, args.namespace)
		if err != nil {
			log.Fatalln(err)
		}
	case "getPodLifetimes":
		lifetimes, err := GetPodLifetimes(args.names[0], args.namespace)
		if err != nil {
			log.Fatalln(err)
		}
		printMap(lifetimes)
	case "getPodLogs":
		logs, err := GetPodLogs(args.names[0], args.namespace)
		if err != nil {
			log.Fatalln(err)
		}
		printMap(logs)
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

/* parseTargetArgs takes an array of arguments that are either in map format {key:value, key:value...} or array format {value, value} and
   returns the data correctly formatted in a map or array. It returns nil for the data structure that the array argument is not */
func parseTargetArgs(args []string) (labels map[string]string, names []string, err error) {
	isMap, err := checkMap(args)
	if err != nil {
		log.Fatalln(err)
	}
	if isMap {
		labels, err := convStringsToMap(args)
		if err != nil {
			return nil, nil, err
		}
		return labels, nil, nil
	} else {
		return nil, args, nil
	}
}

/* parseArgs parses an array of arguments, usually from os.Args, and returns a kubeCmd struct containing all the relevant arguments */
func parseArgs(osArgs []string) kubeCmd {
	cmd := getCommand(osArgs)
	args := kubeCmd{}
	args.cmd = cmd

	switch cmd {
	case "getNumWithLabels", "getNames":
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
	case "getScale", "toggleOn", "toggleOff", "reset":
		if len(osArgs) < 4 {
			args.cmd = "error"
			break
		}
		err := error(nil)
		args.labels, args.names, err = parseTargetArgs(osArgs[2 : len(osArgs)-1])
		if err != nil {
			log.Fatalln(err)
		}
		args.namespace = osArgs[len(osArgs)-1]
		args.scale = -1
	case "setScale":
		if len(osArgs) < 5 {
			args.cmd = "error"
			break
		}
		scale, err := strconv.ParseInt(osArgs[len(osArgs)-2], 10, 32)
		if err != nil {
			log.Fatalln(err)
		}
		args.labels, args.names, err = parseTargetArgs(osArgs[2 : len(osArgs)-2])
		if err != nil {
			log.Fatalln(err)
		}
		args.namespace = osArgs[len(osArgs)-1]
		args.scale = int32(scale)
	case "getPodLogs", "getPodLifetimes":
		if len(osArgs) != 4 {
			args.cmd = "error"
			break
		}
		names := []string{osArgs[2]}

		args.names = names
		args.namespace = osArgs[3]
		args.scale = -1
	default:
		args.cmd = "error"
	}

	return args
}

func main() {
	doCommand(parseArgs(os.Args))
}
