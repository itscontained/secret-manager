# Troubleshooting

When troubleshooting secret-manager your best friend is kubectl describe, this will give you information on the resources as well as recent events. Secret-manager logs may be quite verbose and should be used if following steps do not provide help.

## Troubleshooting a failed external secret

1. Check the status of the `ExternalSecret` resource

```
$ kubectl get externalsecret
NAME            LAST SYNC   SYNCED   AGE
example-secret              False    1h
```

If you see one with Synced status False you can get more info using kubectl describe externalscret, if the status is True that means that secret-manager has successfully created the secret.

```
$ kubectl describe externalsecret <es-name>
[...]
Status:
  Conditions:
    Last Transition Time:        2020-10-17T21:45:22Z
    Message:                     Cannot setup store client: unable to authenticate to Vault store
    Reason:                      Resource is not available for use
    Status:                      False
    Type:                        Ready
Events:
  Type    Reason     Age   From          Message
  ----    ------     ----  ----          -------
```

Here you will find more info about the external secret status under Status field. Future event information is planned to be added to provide more detail about the controllers actions ([#68](https://github.com/itscontained/secret-manager/issues/68)).

## Troubleshooting a crashing secret-mananger

The logs of secret-manager should help describe the issue which is causing secret-manager to crash.

`"no matches for kind "ExternalSecret" in version "secret-manager.itscontained.io/..."`

* This likely means that the CustomResourceDefinitions for secret-manager are not installed. If using the official helm chart to install, ensure that `installCRDs` is set to true. Check if the crds are installed with `kubectl get crd | grep secret-manager`.
