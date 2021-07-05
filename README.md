
# kubeToggler
## Overview
KubeToggler is a lightweight command line tool built using the [client-go kubernetes API](https://pkg.go.dev/k8s.io/client-go). It can target kubernetes deployments by their labels or names and then set/get deployment their attributes. KubeToggler was primarily built as an example use case of client-go.
## Getting Started
### Dependencies
* Kubernetes Engine (minikube, k3s, etc...)
* Kubectl or another kubernetes interface
### Installation
* ``git clone https://github.com/ryan-robinson1/kubeToggler.git ``
### Setup
* cd into the kubeToggler directory and build the binary with ``go build``
* kubeToggler needs a local version of your kube config file in order to interface with kubernetes. On CentOS 7 you can find this file in your home directory here: ``~/.kube/config``. Copy the config file into this repository's local ``.kube`` directory.
* Because kubeToggler is built to access deployments from their labels, you'll need to make sure your deployments have labels. Assuming you have a kubernetes cluster running, use ``kubectl get deployments -n myNamespace --show-labels`` to get the deployment names and their labels. 
* To add labels to your deployments, you can use ``kubectl label deployments -n myNamespace myDeployment myLabel=label1``

## Commands

---
### toggleOn
 <font size="3">Toggles on the deployments that contain the specified labels or names by setting their scales to 1</font> 

### toggleOff
 <font size="3">Toggles off the deployments that contain the specified labels or names by setting their scales to 0</font> 

### reset
 <font size="3">Resets the deployments that contain the specified labels or names by setting their scales to 0 and then back to 1</font> 

#### getName 
 <font size="3">Retrieves the name of the deployments that contain the specified labels</font> 

#### getNumWithLabels
 <font size="3">Retrieves the number of deployments in a namespace that contain the specified labels </font> 

 #### getScale
 <font size="3">Retrieves the scale of the deployments that contain the specified labels or names</font>  


#### setScale
 <font size="3">Sets the scale of the deployments that contain the specified labels or names</font> 


## Usage
>All kubeToggler commands take one or more  key-value label pairs and a namespace.
<pre>$ ./kubeToggler toggleOn {<span style="color:magenta"><i><b>LABEL_KEY</b></i></span>=<span style="color:magenta"><i><b>LABEL_VALUE</b></i></span>|<span style="color:magenta"><i><b>DEPLOYMENT_NAME</b></i></span>} ... <span style="color:magenta"><i><b>NAMESPACE</b></i></span> </pre>
<pre>$ ./kubeToggler toggleOff {<span style="color:magenta"><i><b>LABEL_KEY</b></i></span>=<span style="color:magenta"><i><b>LABEL_VALUE</b></i></span>|<span style="color:magenta"><i><b>DEPLOYMENT_NAME</b></i></span>} ... <span style="color:magenta"><i><b>NAMESPACE</b></i></span> </pre>
<pre>$ ./kubeToggler reset {<span style="color:magenta"><i><b>LABEL_KEY</b></i></span>=<span style="color:magenta"><i><b>LABEL_VALUE</b></i></span>|<span style="color:magenta"><i><b>DEPLOYMENT_NAME</b></i></span>} ... <span style="color:magenta"><i><b>NAMESPACE</b></i></span> </pre>
<pre>$ ./kubeToggler getName <span style="color:magenta"><i><b>LABEL_KEY</b></i></span>=<span style="color:magenta"><i><b>LABEL_VALUE</b></i></span> ... <span style="color:magenta"><i><b>NAMESPACE</b></i></span> </pre>
<pre>$ ./kubeToggler getNumWithLabels <span style="color:magenta"><i><b>LABEL_KEY</b></i></span>=<span style="color:magenta"><i><b>LABEL_VALUE</b></i></span> ... <span style="color:magenta"><i><b>NAMESPACE</b></i></span> </pre>
<pre>$ ./kubeToggler getScale {<span style="color:magenta"><i><b>LABEL_KEY</b></i></span>=<span style="color:magenta"><i><b>LABEL_VALUE</b></i></span>|<span style="color:magenta"><i><b>DEPLOYMENT_NAME</b></i></span>} ... <span style="color:magenta"><i><b>NAMESPACE</b></i></span> </pre>
>The setScale command also requires an integer scale value to set the scale to.
<pre>$ ./kubeToggler setScale {<span style="color:magenta"><i><b>LABEL_KEY</b></i></span>=<span style="color:magenta"><i><b>LABEL_VALUE</b></i></span>|<span style="color:magenta"><i><b>DEPLOYMENT_NAME</b></i></span>} ... <span style="color:magenta"><i><b>SCALE_VALUE NAMESPACE</b></i></span> </pre>

## Examples
    $ ./kubeToggler getName myLabel1=value1 myLabel2=value2 myNamespace
    myConnector
    
    $ ./kubeToggler getScale myLabel1=value1 myLabel2=value2 myNamespace
    myConnector: 1
    
    $ ./kubeToggler getScale myConnector myNamespace
    myConnector: 1
 


---
