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
)

type (
	Options struct {
		Kubeconfig         string
		Namespace          string
		ServiceAccountName string
		Impersonate        string
		OutputFile         string
		Help               bool
		Version            bool
	}
)

var options Options

func must(err error, msg string, v ...any) {
	if err != nil {
		newmsg := fmt.Sprintf(msg, v...)
		log.Fatalf("%s: %s", newmsg, err)
	}
}

func init() {
	flag.StringVarP(&options.Kubeconfig, "kubeconfig", "k", "", "path to the kubeconfig file")
	flag.StringVarP(&options.Namespace, "namespace", "n", "", "namespace containing serviceaccount")
	flag.StringVar(&options.Impersonate, "as", "", "impersonate a user or serviceaccount")
	flag.StringVarP(&options.OutputFile, "output", "o", "", "write configuration to named file")
	flag.BoolVarP(&options.Help, "help", "h", false, "")
	flag.BoolVarP(&options.Version, "version", "v", false, "")

	flag.Usage = usage
}

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

func usage() {
	prg := os.Args[0]
	fmt.Fprintf(flag.CommandLine.Output(), "%s: usage: %s [options] serviceAccountName\n\nOptions:\n\n", prg, prg)
	flag.CommandLine.PrintDefaults()
}

func parseArgs() {
	flag.Parse()

	if options.Help {
		/* --help output should always go to stdout. I will die on this hill. */
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

	must(clientcmdapi.MinifyConfig(config), "failed to minify configuration")
	must(clientcmdapi.FlattenConfig(config), "failed to flatten configuration")

	if options.Namespace == "" {
		options.Namespace = config.Contexts[config.CurrentContext].Namespace
	}

	tokenResponse, err := requestToken(config)
	must(err, "failed to acquire token for serviceaccount %s", options.ServiceAccountName)
	addServiceAccountToken(config, tokenResponse)

	writeConfig(config)
}

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
