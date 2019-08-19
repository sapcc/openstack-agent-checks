openstack-agent-checks
======================

This package provides simple readiness and liveness checks for openstack agents. There can be used as appropriate checks in kubernetes pod checks.

Currently supported
-------------------

- Neutron agent liveness (neutron-agent-liveness)
- Neutron dhcp agent readiness (neutron-dhcp-readiness)
- Neutron linuxbridge readiness (neutron-linuxbridge-readiness)

Requirements
------------

- golang >=1.10
- GNU make

Build
-----

    make
