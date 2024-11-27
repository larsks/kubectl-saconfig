package main

import (
	"context"
	"fmt"
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
		Help               bool
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
	flag.BoolVarP(&options.Help, "help", "h", false, "")
	flag.StringVar(&options.Impersonate, "as", "", "impersonate a user or serviceaccount")
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

	// Create the token request
	tokenRequest := &authv1.TokenRequest{
		Spec: authv1.TokenRequestSpec{
			Audiences: []string{"https://kubernetes.default.svc"},
		},
	}

	// Call the CreateToken method
	return clientset.CoreV1().ServiceAccounts(options.Namespace).CreateToken(
		context.TODO(), options.ServiceAccountName, tokenRequest, metav1.CreateOptions{})
}

func parseArgs() {
	flag.Parse()

	if options.Help {
		/* --help output should always go to stdout. I will die on this hill. */
		flag.CommandLine.SetOutput(os.Stdout)
		flag.Usage()
		return
	}

	if flag.NArg() != 1 {
		log.Fatalf("missing serviceaccount name\n")
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

	ctx := config.CurrentContext
	if options.Namespace == "" {
		options.Namespace = config.Contexts[ctx].Namespace
	}

	qualName := fmt.Sprintf("system:serviceaccount:%s:%s", options.Namespace, options.ServiceAccountName)

	tokenResponse, err := requestToken(config)
	must(err, "failed to acquire token for serviceaccount %s", options.ServiceAccountName)

	config.Contexts[qualName] = &clientcmdapi.Context{
		Cluster:   config.Contexts[ctx].Cluster,
		AuthInfo:  qualName,
		Namespace: options.Namespace,
	}
	delete(config.Contexts, ctx)
	config.CurrentContext = qualName

	config.AuthInfos = map[string]*clientcmdapi.AuthInfo{
		qualName: {
			Token: tokenResponse.Status.Token,
		},
	}

	content, err := clientcmd.Write(*config)
	must(err, "failed to marshal configuration")
	os.Stdout.Write(content)
}