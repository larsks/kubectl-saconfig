# kubectl-saconfig

This is a `kubectl` plugin that will generate a kubeconfig file for authenticating as a service account.

## Usage

```
kubectl-saconfig [options] serviceAccountName
```

Where `[options]` are:

```
      --as string           impersonate a user or serviceaccount
  -h, --help
  -k, --kubeconfig string   path to the kubeconfig file
  -n, --namespace string    namespace containing serviceaccount
  -o, --output string       write configuration to named file
  -v, --version
```

## Example

To generate a configuration for service account `default` in the current namespace:

```
kubectl-saconfig -o kubeconfig default
```

To generate a configuration for service account `vault-secret-reader` in the `acct-mgt` namespace, using "sudo mode":

```
kubectl-saconfig -o kubeconfig -n acct-mgt --as system:admin vault-secret-reader
```
