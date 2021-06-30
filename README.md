
# KubeLabelCtl

## Overview
KubeLabelCtl is a lightweight command line tool built using the [client-go kubernetes API](https://pkg.go.dev/k8s.io/client-go). It can retreive kubernetes deployments by their labels and then set/get deployment attributes. KubeLabelCtl was primarily built as an example use case of client-go.
## Getting Started
### Dependencies
* Kubernetes Engine (minikube, k3s, etc...)
* Kubectl or another kubernetes interface
### Installation
* ``git clone https://github.com/ryan-robinson1/kubeLabelCtl.git ``
### Setup
* KubeLabelCtl needs a local version of your kube config file in order to interface with kubernetes. On CentOS 7 you can find this file in your home directory here: ``~/.kube/config``. Copy the config file into this repository's local ``.kube`` directory.
* Because KubeLabelCtl is built to access deployments from their labels, you'll need to make sure your deployments have labels. Assuming you have a kubernetes cluster running, use ``kubectl get deployments -n myNamespace --show-labels`` to get the deployment names and their labels. 
* To add labels to your deployments, you can use ``kubectl label deployments -n myNamespace myDeployment myLabel=label1``

## Commands
#### getName 
 <font size="3">Retrieves the name of the deployment that contains the specified labels</font> 
#### getScale
 <font size="3">Retrieves the scale of the deployment that contains the specified labels</font>  
#### getNumWithLabels
 <font size="3">Retrieves the number of deployments in a namespace that contain the specified labels </font> 
#### setScale
 <font size="3">Sets the scale of the deployment that contains the specified labels</font> 

## Usage
All kubeLabelCtl commands take one or more  key-value label pairs and a namespace.
<pre>$ ./kubeLabelCtl getName <span style="color:magenta"><i><b>LABEL_KEY</b></i></span> ... <span style="color:magenta"><i><b>LABEL_VALUE</b></i></span> ... <span style="color:magenta"><i><b>NAMESPACE</b></i></span> </pre>
<pre>$ ./kubeLabelCtl getScale <span style="color:magenta"><i><b>LABEL_KEY</b></i></span> ... <span style="color:magenta"><i><b>LABEL_VALUE</b></i></span> ... <span style="color:magenta"><i><b>NAMESPACE</b></i></span> </pre>
<pre>$ ./kubeLabelCtl getNumWithLabels <span style="color:magenta"><i><b>LABEL_KEY</b></i></span> ... <span style="color:magenta"><i><b>LABEL_VALUE</b></i></span> ... <span style="color:magenta"><i><b>NAMESPACE</b></i></span> </pre>

The setScale command also requires an integer scale value to set the scale to.
<pre>$ ./kubeLabelCtl setScale <span style="color:magenta"><i><b>LABEL_KEY</b></i></span> ... <span style="color:magenta"><i><b>LABEL_VALUE</b></i></span> ... <span style="color:magenta"><i><b>SCALE_VALUE NAMESPACE</b></i></span> </pre>


---
