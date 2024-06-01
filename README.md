<!--
Copyright 2021-present Open Networking Foundation
SPDX-License-Identifier: Apache-2.0
-->
[![Go Report Card](https://goreportcard.com/badge/github.com/omec-project/simapp)](https://goreportcard.com/report/github.com/omec-project/simapp)

# Simapp
## Sim subscription app for Aether
### Subscriber Configuration
- SIMApp can configure devices in 5G as well as in 4G.
- look for config/simapp.yaml file for all override options
- Update subscriber range for your usecase
- Make sure sub-provison-endpt points to webui (5g configpod service) or
config4g (4G configpod service)
- In case sub-proxy is running in the deployment then sub-proxy-endpt should
point to sub-proxy (ROC Component). Also port needs to be changed accordingly

### Optionaly simapp can be used to configure network slices in the 4G/5G network functions as well
- Update device-groups, network-slices as per your need
- provision-network-slice should be set to true if you need simapp to configure
slices as well
- In case ROC is running in the deployment then network slices can be configured
from ROC

## Reach out to us thorugh
1. #sdcore-dev channel in [ONF Community Slack](https://onf-community.slack.com/)
2. Extensive SD-Core documentation can be found at [SD-Core Documentation](https://docs.sd-core.opennetworking.org/master/index.html)
3. Raise Github issues
