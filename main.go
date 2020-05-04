package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Krucible/krucible-go-client/krucible"
	"github.com/codefresh-io/stevedore/pkg/codefresh"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getEnvVarOrDie(name string) string {
	val, isSet := os.LookupEnv(name)
	if !isSet {
		panic(name + " is not set")
	}
	return val
}

func main() {
	codefreshToken, isSet := os.LookupEnv("CODEFRESH_API_KEY")
	if !isSet {
		panic("CODEFRESH_API_KEY not set")
	}

	krucibleClient := krucible.NewClient(krucible.ClientConfig{
		AccountID:    getEnvVarOrDie("KRUCIBLE_ACCOUNT_ID"),
		APIKeyId:     getEnvVarOrDie("KRUCIBLE_API_KEY_ID"),
		APIKeySecret: getEnvVarOrDie("KRUCIBLE_API_KEY_SECRET"),
	})

	clusterConfig := krucible.CreateClusterConfig{
		DisplayName: getEnvVarOrDie("KRUCIBLE_CLUSTER_NAME"),
	}

	snapshotID := strings.TrimSpace(os.Getenv("KRUCIBLE_SNAPSHOT_ID"))
	if snapshotID != "" {
		clusterConfig.SnapshotID = snapshotID
	}

	clusterDurationString := os.Getenv("KRUCIBLE_CLUSTER_DURATION")
	clusterDuration, err := strconv.Atoi(clusterDurationString)
	if clusterDuration >= 1 && clusterDuration <= 6 && err == nil {
		clusterConfig.DurationInHours = &clusterDuration
	} else {
		clusterConfig.DurationInHours = krucible.Permanent
	}

	cluster, clientset, err := krucibleClient.CreateCluster(clusterConfig)
	if err != nil {
		panic(err)
	}

	attempts := 0
	var serviceaccount *v1.ServiceAccount = nil
	for serviceaccount == nil || len(serviceaccount.Secrets) == 0 {
		serviceaccount, err = clientset.
			CoreV1().
			ServiceAccounts("default").
			Get(context.Background(), "default", metav1.GetOptions{})
		if err != nil || len(serviceaccount.Secrets) == 0 {
			if attempts > 30 {
				if err != nil {
					panic(err)
				}
				panic("service account has no secrets")
			} else {
				attempts += 1
				time.Sleep(2 * time.Second)
			}
		}
	}

	secretName := string(serviceaccount.Secrets[0].Name)
	namespace := serviceaccount.Namespace

	secret, e := clientset.CoreV1().Secrets(namespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if e != nil {
		message := fmt.Sprintf("Failed to get secrets with error:\n%s", e)
		panic(message)
	}

	token := secret.Data["token"]

	api := codefresh.NewCodefreshAPI("https://g.codefresh.io/", codefreshToken)
	body, err := api.Create(cluster.ConnectionDetails.Server, cluster.ID, token, []byte(cluster.ConnectionDetails.CertificateAuthority), false)
	fmt.Fprintln(os.Stderr, string(body))
	fmt.Print(cluster.ID)
	if err != nil {
		panic(err)
	}
}
