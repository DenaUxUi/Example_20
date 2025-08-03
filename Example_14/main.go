package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/ec2"
    ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type Response struct {
    Message string `json:"system-response"`
}

var instanceStack []string // глобальный список инстансов

func createVMHandler(w http.ResponseWriter, r *http.Request) {
    ctx := context.Background()

    cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("eu-central-1"))
    if err != nil {
        http.Error(w, "Failed to load AWS config", http.StatusInternalServerError)
        return
    }

    ec2Client := ec2.NewFromConfig(cfg)

    input := &ec2.RunInstancesInput{
        ImageId:      aws.String("ami-0a72753edf3e631b7"),
        InstanceType: ec2Types.InstanceTypeT3Micro,
        MinCount:     aws.Int32(1),
        MaxCount:     aws.Int32(1),
    }

    result, err := ec2Client.RunInstances(ctx, input)
    if err != nil {
        log.Printf("EC2 error: %v\n", err)
        http.Error(w, "Failed to create instance", http.StatusInternalServerError)
        return
    }

    instanceID := *result.Instances[0].InstanceId
    instanceStack = append(instanceStack, instanceID)

    response := Response{Message: fmt.Sprintf("EC2 instance created: %s", instanceID)}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func deleteVMHandler(w http.ResponseWriter, r *http.Request) {
    if len(instanceStack) == 0 {
        http.Error(w, "No instance to delete", http.StatusBadRequest)
        return
    }

    lastID := instanceStack[len(instanceStack)-1]
    instanceStack = instanceStack[:len(instanceStack)-1]

    ctx := context.Background()
    cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("eu-central-1"))
    if err != nil {
        http.Error(w, "Failed to load AWS config", http.StatusInternalServerError)
        return
    }

    ec2Client := ec2.NewFromConfig(cfg)

    _, err = ec2Client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
        InstanceIds: []string{lastID},
    })
    if err != nil {
        log.Printf("Terminate error: %v\n", err)
        http.Error(w, "Failed to terminate instance", http.StatusInternalServerError)
        return
    }

    response := Response{Message: fmt.Sprintf("EC2 instance %s terminated", lastID)}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

    if r.Method == http.MethodOptions {
        w.WriteHeader(http.StatusOK)
        return
    }

    if r.Method != http.MethodGet {
        http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
        return
    }

    name := r.URL.Query().Get("name")
    if name == "" {
        http.Error(w, "Missing 'name' parameter", http.StatusBadRequest)
        return
    }

    response := Response{Message: fmt.Sprintf("Hello, %s!", name)}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func main() {
    http.Handle("/", http.FileServer(http.Dir("./static")))
    http.HandleFunc("/api/aws/vm/create", createVMHandler)
    http.HandleFunc("/api/aws/vm/delete", deleteVMHandler)
    http.HandleFunc("/api/hello", helloHandler)

    fmt.Println("Server started at http://localhost:8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal(err)
    }
}

