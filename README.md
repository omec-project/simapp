<!--
Copyright 2021-present Open Networking Foundation
SPDX-License-Identifier: Apache-2.0
SPDX-License-Identifier: LicenseRef-ONF-Member-Only-1.0
-->

# simapp
## Sim subscription app for Aether
### Subscriber Configuration
        - SIMApp can configure devices in 5G as well as in 4G.
        - look for config/simapp.yaml file for all override options
        - Update subscriber range for your usecase
        - Make sure sub-provision-endpt points to webui (in case of 5G) and config4g (in case of 4G).
        - In case ROC is running in the deployment then sub-provison-endpt should point to ROC. Also port needs to be changed accordingly
### Optionaly simapp can be used to configure network slices in the 4G/5G network functions as well.
        - Update device-groups, network-slices as per your need
        - provision-network-slice should be set to true if you need simapp to configure slices as well.
        - In case ROC is running in the deployment then network slices can be configured from ROC.
