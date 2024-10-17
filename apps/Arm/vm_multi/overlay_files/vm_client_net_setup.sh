#!/bin/sh
#
# Copyright 2020, Data61, CSIRO (ABN 41 687 119 230)
#
# SPDX-License-Identifier: BSD-2-Clause
#

set -e

# mkdir -p /etc/dhcp/
# touch /etc/dhcp/dhclient.conf

boot_args=$(cat /proc/cmdline)
# Extract the value of node=
vm_name=$(grep -o 'node=[^ ]*' /proc/cmdline | cut -d'=' -f2)

# echo "$vm_name" >> /etc/dhcp/dhclient.conf

# random_mac=$(printf '02:%02X:%02X:%02X:%02X:%02X\n' $((RANDOM%256)) $((RANDOM%256)) $((RANDOM%256)) $((RANDOM%256)) $((RANDOM%256)))
# ip link set dev eth0 address "02:00:00:00:AB:${vm_name:2}"

ip link set up dev eth0
ip link add link eth0 name eth0.4094 type vlan id 4094
ip link set up dev eth0.4094

udhcpc -i eth0.4094 -x hostname:$vm_name

# ip link set up dev eth0
# 
# node_str=$(cat /proc/cmdline | awk '{print $1}')
# node_num="${node_str: -1}"
# 
# ip link add link eth0 name "eth0.$node_num" type vlan id "$node_num"
# 
# ip link set up dev "eth0.$node_num"
# 
# ip addr add "192.168.10$node_num.1/24" dev "eth0.$node_num"
# 
# # Replace 'default_gateway' with your actual gateway IP
# ip route add default via 192.168.10"$node_num".9 dev "eth0.$node_num"
# 
# # Try and ping another nodes infra link
# ping 192.168.10$(($node_num + 1)).1 &


# #!/bin/bash
# 
# # Configuration
# SERVER_IP="central.node.ip.address"  # Replace with the central node's IP address
# SERVER_PORT=9999
# 
# # Function to request VLAN from the server
# request_vlan() {
#     echo "REQUEST_VLAN" | socat - UDP-DATAGRAM:255.255.255.255:$SERVER_PORT,broadcast
#     RESPONSE=$(timeout 5 socat - UDP-RECVFROM:$SERVER_PORT,broadcast)
#     echo $RESPONSE
# }
# 
# # Function to configure VLAN on the client
# configure_vlan() {
#     VLAN_ID=$1
#     INTERFACE="eth0.$VLAN_ID"
#     IP_ADDRESS="192.168.$VLAN_ID.9/24"
# 
#     # Create VLAN interface
#     ip link add link eth0 name $INTERFACE type vlan id $VLAN_ID
# 
#     # Assign IP address
#     ip addr add $IP_ADDRESS dev $INTERFACE
# 
#     # Bring VLAN interface up
#     ip link set $INTERFACE up
# }
# 
# # Main function
# main() {
#     while true; do
#         RESPONSE=$(request_vlan)
#         case "$RESPONSE" in
#             VLAN*)
#                 VLAN_ID=$(echo $RESPONSE | awk '{print $2}')
#                 configure_vlan $VLAN_ID
#                 echo "Configured VLAN $VLAN_ID"
#                 break
#                 ;;
#             *)
#                 echo "No VLAN available"
#                 sleep 1  # Adding a delay to avoid spamming the request
#                 ;;
#         esac
#     done
# }
# 
# 
# 
# main &
# 
