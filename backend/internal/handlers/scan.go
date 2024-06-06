package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/akshat-crypto/TagTracker/backend/internal/services/aws"
)

func ScanHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		IAMRoleARN string `json:"iam_role_arn"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := aws.StartScan(request.IAMRoleARN)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ec2Instances, err := aws.FetchEC2Instances()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rdsInstances, err := aws.FetchRDSInstances()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		EC2Instances []aws.EC2Instance `json:"ec2_instances"`
		RDSInstances []aws.RDSInstance `json:"rds_instances"`
	}{
		EC2Instances: ec2Instances,
		RDSInstances: rdsInstances,
	}

	fmt.Println(response)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	//TODO: Print only scanning status
	// w.WriteHeader(http.StatusOK)
	// json.NewEncoder(w).Encode(map[string]string{"status": "scan initiated"})
}
