# Secret Manager
Secret Manager is a Kubernetes add-on to automate the creation and renewal of
secrets from various external secret sources.

Secret Manager can also reformat the sourced secrets to fit the configuration
expected by the workloads using the created secrets.

Based on the work from [godaddy/kubernetes-external-secrets](https://github.com/godaddy/kubernetes-external-secrets)
and with borrowed wisdom from [jetstack/cert-manager](https://github.com/jetstack/cert-manager).

## Installation

Helm installation steps can be found on the chart readme at [artifacthub.io](https://artifacthub.io/packages/helm/itscontained/secret-manager)

## Documentation

Documentation and examples for supported external secret sources can be found in
the [docs directory](docs/README.md) of this project.

## Support

If you encounter any issues whilst using secret-manager, we have a number of places you 
can use to try and get help.

First of all we recommend looking at the [troubleshooting guide](docs/troubleshooting.md) of our documentation.

The quickest way to ask a question is to first post on [#external-secrets](https://kubernetes.slack.com/archives/C017BF84G2Y)
channel on the Kubernetes Slack. There are a lot of community members in this channel, and
you can often get an answer to your question straight away!

You can also try [searching for an existing issue](https://github.com/itscontained/secret-manager/issues). Properly searching
for an existing issue will help reduce the number of duplicates, and help you find the answer you are looking for quicker.

If you believe you have encountered a bug, and cannot find an existing issue similar to your own,
you may open a new issue. Please be sure to include as much information as possible about your environment.

## Contributing

We welcome pull requests with open arms! There's a lot of work to do here, and we're especially concerned with
ensuring the longevity and reliability of the project.

Please take a look at our [issue tracker](https://github.com/itscontained/secret-manager/issues) if you are
unsure where to start with getting involved!

Developer documentation is available in the [official documentation](docs/contributing).
