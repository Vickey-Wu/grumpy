package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

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

	//if r.URL.Path != "/validate" {
	//	glog.Error("no validate")
	//	http.Error(w, "no validate", http.StatusBadRequest)
	//	return
	//}
	if (r.URL.Path != "/mutate" || r.URL.Path != "/validate") {
		glog.Error("no mutate or validate")
		http.Error(w, "no mutate or validate", http.StatusBadRequest)
		return
        }

	arRequest := v1beta1.AdmissionReview{}
	if err := json.Unmarshal(body, &arRequest); err != nil {
		glog.Error("incorrect body")
		http.Error(w, "incorrect body", http.StatusBadRequest)
	}

	raw := arRequest.Request.Object.Raw
	pod := v1.Pod{}

	if err := json.Unmarshal(raw, &pod); err != nil {
		glog.Error("error deserializing pod")
		return
	}

	type patchOperation struct {
		Op    string      `json:"op"`
		Path  string      `json:"path"`
		Value interface{} `json:"value,omitempty"`
	}

	func createPatch() ([]byte, error) {
		var patch []patchOperation
		patch = append(patch, patchOperation{
			Op:    "add",
			Path:  "/metadata/labels",
			Value: "address: vickey-wu.com",
        	})
		return json.Marshal(patch)
	}

	patchBytes, err := createPatch()

	if (pod.Name != "smooth-app" && r.URL.Path == "/mutate") {
		arResponse := v1beta1.AdmissionReview{
			Response: &v1beta1.AdmissionResponse{
				Allowed: true,
				Patch:   patchBytes,
				PatchType: func() *v1beta1.PatchType {
					pt := v1beta1.PatchTypeJSONPatch
					return &pt
			}
		}
	} else if pod.Name != "smooth-app" {
		arResponse := v1beta1.AdmissionReview{
			Response: &v1beta1.AdmissionResponse{
				Allowed: false,
				Result: &metav1.Status{
					Message: "validating pod attr!!!",
				},
			},
		}
		//return
        }

	//arResponse := v1beta1.AdmissionReview{
	//	Response: &v1beta1.AdmissionResponse{
	//		Allowed: false,
	//		Result: &metav1.Status{
	//			Message: "Keep calm and don't add more crap to the cluster!",
	//		},
	//	},
	//}

	resp, err := json.Marshal(arResponse)
	if err != nil {
		glog.Errorf("Can't encode response: %v", err)
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
	}
	glog.Infof("Ready to write reponse ...")
	if _, err := w.Write(resp); err != nil {
		glog.Errorf("Can't write response: %v", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}
