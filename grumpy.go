package main

import (
	"encoding/json"
	"fmt"
	// "os"
	// "regexp"
	// "strings"

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

	pod := v1.Pod{}
	if err := json.Unmarshal(raw, &pod); err != nil {
		glog.Error("error deserializing pod")
		return
	}

	app := "./cosign"
	subcommand := "verify"
	arg0 := "--key"
	// Replace with env variable
	arg1 := "cosign.pub"
	arg2 := "--allow-insecure-registry"

	for i := 0; i < len(pod.Spec.Containers); i++ {
		imageName := pod.Spec.Containers[i].Image

		cmd := exec.Command(app, subcommand, arg0, arg1, arg2, imageName)
		stdout, err := cmd.Output()

		if err != nil || len(stdout) == 0 {
			glog.Errorf("No signing information found for image = %s; Error message: %v", imageName, err)
			arResponse := v1beta1.AdmissionReview{
				Response: &v1beta1.AdmissionResponse{
					UID: admissionRequestUID,
					Allowed: false,
					Result: &metav1.Status{
						Message: "No signing information found for " + imageName,
					},
				},
				TypeMeta: metav1.TypeMeta{
					APIVersion: "admission.k8s.io/v1",
					Kind:       "AdmissionReview",
				},
			}
		
			resp, err := json.Marshal(arResponse)
		
			if err != nil {
				glog.Errorf("Can't encode response: %v", err)
				http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
			}
			if _, err := w.Write(resp); err != nil {
				glog.Errorf("Can't write response: %v", err)
				http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
			}
			return
		}
	}
}
