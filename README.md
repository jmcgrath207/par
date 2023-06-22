# Par - Label Based DNS Operator


[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Actions Status](https://github.com/jmcgrath207/par/workflows/ci/badge.svg)](https://github.com/jmcgrath207/par/actions)

Par is a DNS operator that allows you to control deployment DNS queries by labels without cluster administrative changes (ex. [Istio sidecar](https://istio.io/latest/docs/setup/platform-setup/prerequisites/#:~:text=Istio%20proxy%20sidecar%20container) )

![plot](./asssets/par.drawio.png)

## Installation

Provide instructions on how to install and set up your project. Be sure to include any dependencies or prerequisites needed.


## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| image.repository | string | `"local.io/library/par"` |  |
| image.tag | string | `"debug-latest"` |  |
| kubernetesClusterDomain | string | `"cluster.local"` |  |
| metrics | bool | `false` |  |
| requests.cpu | string | `"256m"` |  |
| requests.memory | string | `"128Mi"` |  |
| resources.limits.cpu | string | `"1"` |  |
| resources.limits.memory | string | `"512Mi"` |  |

## License

This project is licensed under the [MIT License](https://opensource.org/licenses/MIT). See the `LICENSE` file for more details.