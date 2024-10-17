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
ovs-vsctl add-port br0 eth0
ovs-vsctl add-port br0 eth1
# ovs-vsctl add-port br0 eth0.100


ip link set up dev br0
ip link set up dev eth0
ip link set up dev eth1


ip link add link eth0 name eth0.4094 type vlan id 4094
ip link set up dev eth0.4094
ip addr add $GATEWAY/24 dev eth0.4094


dnsmasq \
    --strict-order \
    --except-interface=lo \
    --interface="eth0.4094" \
    --listen-address=$GATEWAY \
    --bind-interfaces \
    --dhcp-authoritative  \
    --dhcp-range=$DHCPRANGE \
    --conf-file="" \
    --pid-file=/var/run/dnsmasq-br0.pid \
    --dhcp-leasefile=/var/run/dnsmasq-br0.leases \
    --dhcp-no-override

# Forward data to the bridge and from the bridge. Going to have flows happen in bridge space
ovs-ofctl del-flows br0
ovs-ofctl add-flow br0 "priority=100,in_port=eth0 actions=output:br0"
ovs-ofctl add-flow br0 "priority=0,actions=output:eth0"
# ovs-ofctl add-flow br0 "priority=100,in_port=br0 actions=output:eth0"

# ETH1_IP=$(cat /tmp/dhcp | grep 'obtained from' | awk -F 'obtained from ' '{print $2}' | awk -F ',' '{print $1}')
# 
# rm /tmp/dhcp

# Get the IP address assigned to eth1
# ETH1_IP=$(ip -4 addr show dev eth1 | head -2 | tail -1 | awk '{print $2}' | sed 's/...$//')

# Add a default route via eth1 with a lower priority (higher metric)
# ip route add default via $ETH1_IP dev eth1 metric 200

