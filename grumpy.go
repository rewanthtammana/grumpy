package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
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
	// glog.Info("Received request")

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
	glog.Infof("admissionRequestUID = %s\n", admissionRequestUID)
	// glog.Info(admissionRequestUID)
	
	pod := v1.Pod{}
	if err := json.Unmarshal(raw, &pod); err != nil {
		glog.Error("error deserializing pod")
		return
	}
	// podName := fmt.Sprintf("Pod name = %s\n", pod.Name)
	// glog.Info("podName")
	// glog.Info(podName)
	// podInfo := fmt.Sprintf(">>>> Podx image = %v\n", pod)
	// glog.Info("PodInfo")
	// glog.Info(podInfo)
	// podUID := fmt.Sprintf("Pod name = %s\n", pod.UID)
	// glog.Info("PodUID")
	// glog.Info(podUID)

	// var containerName string
	app := "./notary-slim"
	subcommand := "lookup"
	arg0 := "-s"
	arg1 := os.Getenv("NOTARY_SERVER")

	for i := 0; i < len(pod.Spec.Containers); i++ {
		imageName := pod.Spec.Containers[i].Image

		// glog.Infof("===> Pod name = %s ; Container name = %s\n", podName, imageName)

		// arg2 := os.Getenv("GUN")
		// arg3 := os.Getenv("TARGET")
		gun := strings.Split(imageName, ":")[0]
		target := strings.Split(imageName, ":")[1]
		r, err := regexp.Compile(target + " sha256:[a-f0-9]+")
		if err != nil {
			glog.Errorf("Regex compile error = %v\n", err)
			arResponse := v1beta1.AdmissionReview{
				Response: &v1beta1.AdmissionResponse{
					UID: admissionRequestUID,
					Allowed: false,
					Result: &metav1.Status{
						Message: "Regex compile error",
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
			// return
		}

		cmd := exec.Command(app, subcommand, arg0, arg1, gun, target)
		stdout, err := cmd.Output()

		if err != nil {
			glog.Errorf("No signing information found for = %s\nError message: %v", imageName, err)
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
			// return
		}

		// Print the output
		// glog.Infof("--- Notary output for %s\n", imageName)
		m := r.FindStringIndex(string(stdout))
		if len(m) == 0 {
			glog.Errorf("No matching signature found for %v\n", imageName)
			arResponse := v1beta1.AdmissionReview{
				Response: &v1beta1.AdmissionResponse{
					UID: admissionRequestUID,
					Allowed: false,
					Result: &metav1.Status{
						Message: "Length = 0; No signing information found for " + imageName,					},
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
		}
		glog.Infof("+++ %v\n", string(stdout))
		// fmt.Println(string(stdout))
	}

	// if pod.Name == "smooth-app" {
	// 	return
	// }

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

	resp, err := json.Marshal(arResponse)

	if err != nil {
		glog.Errorf("Can't encode response: %v", err)
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
	}
	if _, err := w.Write(resp); err != nil {
		glog.Errorf("Can't write response: %v", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}
