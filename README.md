<!--
Copyright 2021-present Open Networking Foundation
SPDX-License-Identifier: Apache-2.0
-->
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/omec-project/simapp/badge)](https://scorecard.dev/viewer/?uri=github.com/omec-project/simapp)
[![GitHub latest release (SemVer)](https://img.shields.io/github/v/release/omec-project/simapp?sort=semver)](https://github.com/omec-project/simapp/releases)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://spdx.org/licenses/Apache-2.0.html)

# Simapp
## Sim subscription app for Aether
### Subscriber Configuration
- SIMApp can configure devices in 5G as well as in 4G.
- look for config/simapp.yaml file for all override options
- Update subscriber range for your usecase
- Make sure sub-provison-endpt points to webui (5g configpod service) or
config4g (4G configpod service)

### Optionaly simapp can be used to configure network slices in the 4G/5G network functions as well
- Update device-groups, network-slices as per your need

## Reach out to us through
1. #sdcore-dev channel in [ONF Community Slack](https://aether5g-project.slack.com/)
2. Extensive SD-Core documentation can be found at [SD-Core Documentation](https://docs.sd-core.aetherproject.org/index.html)
3. Raise Github issues
