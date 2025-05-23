# kubectl-saconfig

This is a `kubectl` plugin that will generate a kubeconfig file for authenticating as a service account.

## Installation

To install into `/usr/local/bin`:

```
make install
```

To install into `~/bin`:

```
make install bindir=$HOME/bin
```

## Usage

```
kubectl-saconfig [options] serviceAccountName
```

Where `[options]` are:

```
    --as string                      Username to impersonate for the operation. User could be a regular user or a service account in a namespace.
    --as-group stringArray           Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
    --as-uid string                  UID to impersonate for the operation.
    --cache-dir string               Default cache directory (default "/home/lars/.kube/cache")
    --certificate-authority string   Path to a cert file for the certificate authority
    --client-certificate string      Path to a client certificate file for TLS
    --client-key string              Path to a client key file for TLS
    --cluster string                 The name of the kubeconfig cluster to use
    --context string                 The name of the kubeconfig context to use
    --disable-compression            If true, opt-out of response compression for all requests to the server
-h, --help
    --insecure-skip-tls-verify       If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
    --kubeconfig string              Path to the kubeconfig file to use for CLI requests.
-n, --namespace string               If present, the namespace scope for this CLI request
-o, --output string                  File to which to write configuration
    --request-timeout string         The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
-s, --server string                  The address and port of the Kubernetes API server
    --tls-server-name string         Server name to use for server certificate validation. If it is not provided, the hostname used to contact the server is used
    --token string                   Bearer token for authentication to the API server
    --user string                    The name of the kubeconfig user to use
-v, --version
```

## Example

To generate write configuration to file `kubeconfig` for service account `default` in the current namespace:

```
kubectl-saconfig -o kubeconfig default
```

To generate write a configuration to file `kubeconfig` for service account `vault-secret-reader` in the `acct-mgt` namespace, using impersonation ("sudo mode"):

```
kubectl-saconfig -o kubeconfig -n acct-mgt --as system:admin vault-secret-reader
```
