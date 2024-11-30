// kubectl-saconfig generates a kubeconfig file for authenticating as service account.
package main

import (
	"context"
	"fmt"
	"io"
	"kubectl-saconfig/version"
	"log"
	"os"

	flag "github.com/spf13/pflag"

	authv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	// per https://krew.sigs.k8s.io/docs/developer-guide/develop/best-practices/
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

type (
	// Holds values of command line options.
	Options struct {
		Kubeconfig         string // Path to kubeconfig file (optional)
		Context            string // Kubeconfig context to use
		Namespace          string // Target namespace (defaults to namespace of current context)
		ServiceAccountName string // Target service account name
		Impersonate        string // Name of user to impersonate (optional)
		OutputFile         string // Output kubeconfig to this file when specified (default to stdout)
		Help               bool   // --help was requested
		Version            bool   // --verbose was requested
	}
)

var options Options

// If err is not nil, log a failure message and exit.
func must(err error, msg string, v ...any) {
	if err != nil {
		newmsg := fmt.Sprintf(msg, v...)
		log.Fatalf("%s: %s", newmsg, err)
	}
}

// Set up command line options processing
func init() {
	flag.StringVarP(&options.Kubeconfig, "kubeconfig", "k", "", "Path to the kubeconfig file")
	flag.StringVarP(&options.Context, "context", "", "", "The name of the kubeconfig context to use")
	flag.StringVarP(&options.Namespace, "namespace", "n", "", "The namespace scope for this CLI request")
	flag.StringVar(&options.Impersonate, "as", "", "Username to impersonate for the operation")
	flag.StringVarP(&options.OutputFile, "output", "o", "", "File to which to write configuration")
	flag.BoolVarP(&options.Help, "help", "h", false, "")
	flag.BoolVarP(&options.Version, "version", "v", false, "")

	flag.Usage = usage
}

// Request a token for the target service account
func requestToken(config *clientcmdapi.Config) (*authv1.TokenRequest, error) {
	clientConfig := clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{})
	restConfig, err := clientConfig.ClientConfig()
	must(err, "failed to extract client configuration")

	if options.Impersonate != "" {
		restConfig.Impersonate = rest.ImpersonationConfig{
			UserName: options.Impersonate,
		}
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	must(err, "failed to create kubernetes client")

	tokenRequest := &authv1.TokenRequest{
		Spec: authv1.TokenRequestSpec{
			Audiences: []string{"https://kubernetes.default.svc"},
		},
	}

	return clientset.CoreV1().ServiceAccounts(options.Namespace).CreateToken(
		context.TODO(), options.ServiceAccountName, tokenRequest, metav1.CreateOptions{})
}

// Print a usage message. Outputs to flag.CommandLine.Output(), which requires
// pflag 1.0.6 or any commit later than 81378bbcd8a.
func usage() {
	prg := os.Args[0]
	fmt.Fprintf(flag.CommandLine.Output(), "%s: usage: %s [options] serviceAccountName\n\nOptions:\n\n", prg, prg)
	flag.CommandLine.PrintDefaults()
}

// Parse command line arguments.
func parseArgs() {
	flag.Parse()

	if options.Help {
		// --help output should always go to stdout. I will die on this hill.
		flag.CommandLine.SetOutput(os.Stdout)
		flag.Usage()
		os.Exit(0)
	}

	if options.Version {
		version.ShowVersion()
		os.Exit(0)
	}

	if flag.NArg() != 1 {
		log.Printf("missing serviceaccount name\n")
		flag.Usage()
		os.Exit(2)
	}

	options.ServiceAccountName = flag.Arg(0)
}

func main() {
	parseArgs()

	pathopts := clientcmd.NewDefaultPathOptions()
	config, err := pathopts.GetStartingConfig()
	must(err, "failed to get kubernetes configuration")

	if options.Context != "" {
		config.CurrentContext = options.Context
	}

	// Minify and flatten the configuration: this gets us a config that
	// contains only the current context, and any external resources
	// have been embedded.
	must(clientcmdapi.MinifyConfig(config), "failed to minify configuration")
	must(clientcmdapi.FlattenConfig(config), "failed to flatten configuration")

	// Default to namespace of current context if not provided
	// explicitly on command line.
	if options.Namespace == "" {
		options.Namespace = config.Contexts[config.CurrentContext].Namespace
	}

	tokenResponse, err := requestToken(config)
	must(err, "failed to acquire token for serviceaccount %s", options.ServiceAccountName)
	addServiceAccountToken(config, tokenResponse)

	writeConfig(config)
}

// Add the service account token to the configuration. This adds a user users section,
// adds a new context, deletes the previous current context, and updates the current context
// to point at the one we just added.
func addServiceAccountToken(config *clientcmdapi.Config, tokenResponse *authv1.TokenRequest) {
	qualName := fmt.Sprintf("system:serviceaccount:%s:%s", options.Namespace, options.ServiceAccountName)

	config.Contexts[qualName] = &clientcmdapi.Context{
		Cluster:   config.Contexts[config.CurrentContext].Cluster,
		AuthInfo:  qualName,
		Namespace: options.Namespace,
	}
	delete(config.Contexts, config.CurrentContext)
	config.CurrentContext = qualName

	config.AuthInfos = map[string]*clientcmdapi.AuthInfo{
		qualName: {
			Token: tokenResponse.Status.Token,
		},
	}
}

// Write configuration to stdout (default) or to the file specified
// in the --output option.
func writeConfig(config *clientcmdapi.Config) {
	var out io.Writer
	var err error

	content, err := clientcmd.Write(*config)
	must(err, "failed to marshal configuration")

	if options.OutputFile != "" {
		out, err = os.Create(options.OutputFile)
		must(err, "failed to open %s for writing", options.OutputFile)
	} else {
		out = os.Stdout
	}
	_, err = out.Write(content)
	must(err, "failed to write configuration file")
}
