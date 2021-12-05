package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"io/ioutil"
	"net/http"
	"os/exec"

	"github.com/golang/glog"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//GrumpyServerHandler listen to admission requests and serve responses
type GrumpyServerHandler struct {
}

func (gs *GrumpyServerHandler) serve(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	if len(body) == 0 {
		glog.Error("empty body")
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}
	glog.Info("Received request")

	if r.URL.Path != "/validate" {
		glog.Error("no validate")
		http.Error(w, "no validate", http.StatusBadRequest)
		return
	}

	arRequest := v1beta1.AdmissionReview{}
	if err := json.Unmarshal(body, &arRequest); err != nil {
		glog.Error("incorrect body")
		http.Error(w, "incorrect body", http.StatusBadRequest)
	}

	raw := arRequest.Request.Object.Raw
	admissionRequestUID := arRequest.Request.UID
	glog.Info("admissionRequestUID")
	glog.Info(admissionRequestUID)
	
	pod := v1.Pod{}
	if err := json.Unmarshal(raw, &pod); err != nil {
		glog.Error("error deserializing pod")
		return
	}
	podName := fmt.Sprintf("Pod name = %s\n", pod.Name)
	glog.Info("podName")
	glog.Info(podName)
	podInfo := fmt.Sprintf(">>>> Podx image = %v\n", pod)
	glog.Info("PodInfo")
	glog.Info(podInfo)
	podUID := fmt.Sprintf("Pod name = %s\n", pod.UID)
	glog.Info("PodUID")
	glog.Info(podUID)

	// var containerName string
	app := "./notary-slim"
	subcommand := "lookup"
	arg0 := "-s"
	arg1 := os.Getenv("NOTARY_SERVER")

	for i := 0; i < len(pod.Spec.Containers); i++ {
		imageName := pod.Spec.Containers[i].Image

		glog.Infof("===> Pod name = %s ; Container name = %s\n", podName, imageName)

		// arg2 := os.Getenv("GUN")
		// arg3 := os.Getenv("TARGET")
		gun := strings.Split(imageName, ":")[0]
		target := strings.Split(imageName, ":")[1]

		cmd := exec.Command(app, subcommand, arg0, arg1, gun, target)
		stdout, err := cmd.Output()

		if err != nil {
			glog.Errorf("Notary error = %v\n", err)
			return
		}

		// Print the output
		glog.Infof("--- Notary output for %s\n", imageName)
		glog.Infof("+++ %v\n", string(stdout))
		fmt.Println(string(stdout))
	}

	if pod.Name == "smooth-app" {
		return
	}

	arResponse := v1beta1.AdmissionReview{
		Response: &v1beta1.AdmissionResponse{
			UID: admissionRequestUID,
			Allowed: false,
			Result: &metav1.Status{
				Message: "Keep calm and not add more crap in the cluster!",
			},
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
	}

	arResponseDebug := fmt.Sprintf(">>>> Podx image = %v\n", arResponse)
	glog.Info("arResponseDebug")
	glog.Info(arResponseDebug)

	resp, err := json.Marshal(arResponse)
	
	// type respArr struct {
	// 	ApiVersion string
	// 	Kind string
	// 	Response *v1beta1.AdmissionReview
	// }
	// respUpdate := &respArr{
	// 	ApiVersion: "admission.k8s.io/v1",
	// 	Kind: "AdmissionReview",
	// 	Response: &arResponse,
	// }
	// respUpdateJson, _ := json.Marshal(respUpdate)

	if err != nil {
		glog.Errorf("Can't encode response: %v", err)
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
	}
	glog.Info("resp")
	glog.Info(string(resp))
	glog.Infof("Ready to write reponse ...")
	if _, err := w.Write(resp); err != nil {
	// if _, err := w.Write(respUpdateJson); err != nil {
		glog.Errorf("Can't write response: %v", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
	glog.Info("resp-end")
	glog.Info(string(resp))
}
