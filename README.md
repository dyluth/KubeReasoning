# KubeReasoning
A reasoning engine for kubernetes clusters using JSONata to manage the data

It is designed to pull out all the config from a cluster - eg with `kubectl get pods -A -o=json`
and then provide an interface to query and work with that data.

for example to get all daemonsets in the `cam` with a name containing `kube`, we would run:
```
    podSet.With(
		"metadata.name~>/kube/i").With(
		"metadata.namespace='cam'").With(
		"metadata.ownerReferences.kind='DaemonSet'").Evaluate()
```

For details of what the queries can look like, see: http://docs.jsonata.org/overview


This is early, but the extension points are:
1) adding more convenience methods for filters (eg WithDaemonset(true) to get all daemonsets)
2) handle more types - deploys, services, secrets, etc
3) integrate more closely with kubernetes to get the info out - eg:
    - on the CLI using kubectl 
    - using the kube client
    - connecting to the cluster this is running as a pod within
