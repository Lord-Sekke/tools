package cmd

import (
	"fmt"
	"github.com/armory/spinnaker-tools/internal/pkg/debug"
	"github.com/armory/spinnaker-tools/internal/pkg/k8s"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var sourceKubeconfig string
var destKubeconfig string
var context string
var namespace string
var serviceAccountName string
var targetNamespaces string
var verbose bool

// createServiceAccount creates a service account and kubeconfig
var createServiceAccount = &cobra.Command{
	Use:   "create-service-account",
	Short: "Create a service account and Kubeconfig",
	Long: `Given a Kubernetes kubeconfig and context, will create the following:
	* Kubernetes ServiceAccount
	* Kubernetes ClusterRole granting the service account access to cluster-admin
	* kubeconfig file with credentials for the ServiceAccount
	* Kubernetes Secret for the service account (if not already present)`,
	Run: func(cmd *cobra.Command, args []string) {

		// Create a debug context
		ctx, err := debug.NewContext(true)
		if err != nil {
			fmt.Println("TODO: This needs error handling")
		}

		cluster := k8s.Cluster{
			KubeconfigFile: sourceKubeconfig,
			Context:        k8s.ClusterContext{ContextName: context},
		}
		// TODO: change parameters
		serr, err := cluster.DefineCluster(ctx, verbose)
		if err != nil || serr != "" {
			color.Red("Defining cluster failed, exiting")
			color.Red(serr)
			color.Red(err.Error())
			os.Exit(1)
		}

		sa := k8s.ServiceAccount{
			Namespace:          namespace,
			ServiceAccountName: serviceAccountName,
			TargetNamespaces:   nil,
		}

		if len(targetNamespaces) != 0 {
			sa.TargetNamespaces = strings.Split(targetNamespaces, ",")
		}

		// Define service account
		serr, err = cluster.DefineServiceAccount(ctx, &sa, verbose)
		if err != nil || serr != "" {
			color.Red("Defining service account failed, exiting")
			color.Red(serr)
			color.Red(err.Error())
			os.Exit(1)
		}

		// Create service account
		serr, err = cluster.CreateServiceAccount(ctx, &sa, verbose)
		if err != nil || serr != "" {
			color.Red("Creating service account failed, exiting")
			color.Red(serr)
			color.Red(err.Error())
			os.Exit(1)
		}

		// Now create the secret manually
		serr, err = createServiceAccountSecret(ctx, &sa)
		if err != nil || serr != "" {
			color.Red("Creating service account secret failed, exiting")
			color.Red(serr)
			color.Red(err.Error())
			os.Exit(1)
		}

		// Define kubeconfig
		f, serr, err := cluster.DefineKubeconfig(destKubeconfig, &sa, verbose)
		if err != nil || serr != "" {
			color.Red("Defining kubeconfig failed, exiting")
			color.Red(serr)
			color.Red(err.Error())
			os.Exit(1)
		}

		// Create Kubeconfig
		o, serr, err := cluster.CreateKubeconfigUsingKubectl(ctx, f, sa, verbose)
		if err != nil || serr != "" {
			color.Red("Creating Kubeconfig failed, exiting")
			color.Red(serr)
			color.Red(err.Error())
			os.Exit(1)
		}
		color.Green("Created kubeconfig file at %s", o)
	},
}

// createServiceAccountSecret creates a Kubernetes secret for the service account
func createServiceAccountSecret(ctx *debug.Context, sa *k8s.ServiceAccount) (string, error) {
	secretYaml := fmt.Sprintf(`apiVersion: v1
kind: Secret
metadata:
  name: %s-token
  annotations:
    kubernetes.io/service-account.name: %s
type: kubernetes.io/service-account-token`, sa.ServiceAccountName, sa.ServiceAccountName)

	// Create the secret using kubectl apply
	cmd := fmt.Sprintf("kubectl apply -f - --context %s", ctx.ContextName)
	err := executeCommand(cmd, secretYaml)
	if err != nil {
		return "", err
	}

	return "Secret created successfully", nil
}

// executeCommand executes a shell command with the provided input
func executeCommand(cmd string, input string) error {
	// Implement command execution (e.g., using os/exec or another method)
	// You can apply the YAML using `kubectl apply -f -` in this case
	return nil // Placeholder: Add actual command execution logic
}

func init() {
	rootCmd.AddCommand(createServiceAccount)

	// TODO: flag for namespace
	// TODO: flag for service account name
	createServiceAccount.PersistentFlags().StringVarP(&sourceKubeconfig, "kubeconfig", "i", "", "kubeconfig to start with")
	createServiceAccount.PersistentFlags().StringVarP(&destKubeconfig, "output", "o", "", "kubeconfig to output to")
	createServiceAccount.PersistentFlags().StringVarP(&context, "context", "c", "", "kubectl context to use")
	createServiceAccount.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "namespace to create service account in")
	createServiceAccount.PersistentFlags().StringVarP(&serviceAccountName, "service-account-name", "s", "", "service account name")
	// createServiceAccount.PersistentFlags().BoolVarP(&notAdmin, "select-namespaces", "T", false, "don't create service account as cluster-admin")
	createServiceAccount.PersistentFlags().StringVarP(&targetNamespaces, "target-namespaces", "t", "", "comma-separated list of namespaces to deploy to")
	createServiceAccount.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}
