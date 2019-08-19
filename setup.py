from setuptools import setup

setup(
    name="openstack-agent-checks",
    version="0.1",
    scripts=[
        'cmd/neutron-agent-liveness/neutron-agent-liveness', 
        'cmd/neutron-dhcp-readiness/neutron-dhcp-readiness',
        'cmd/neutron-linuxbridge-readiness/neutron-linuxbridge-readiness'
    ],
    exclude_package_data={'': ['tests']},
)
