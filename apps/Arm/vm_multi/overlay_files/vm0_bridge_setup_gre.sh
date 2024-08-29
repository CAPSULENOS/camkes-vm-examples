#!/bin/sh
#
# Copyright 2020, Data61, CSIRO (ABN 41 687 119 230)
#
# SPDX-License-Identifier: BSD-2-Clause
#

set -e

# Set up internal resolution
NETMASK=255.255.255.0
GATEWAY=10.192.10.1
DHCPRANGE=10.192.10.100,10.192.10.254

/usr/share/openvswitch/scripts/ovs-ctl start
ovs-vsctl add-br br0
ip link set up dev br0

ip link set up dev eth0
ip addr add $GATEWAY/24 dev eth0

dnsmasq \
    --strict-order \
    --except-interface=lo \
    --interface=eth0 \
    --listen-address=$GATEWAY \
    --bind-interfaces \
    --dhcp-authoritative  \
    --dhcp-range=$DHCPRANGE \
    --conf-file="" \
    --pid-file=/var/run/dnsmasq-eth0.pid \
    --dhcp-leasefile=/var/run/dnsmasq-eth0.leases \
    --dhcp-no-override
