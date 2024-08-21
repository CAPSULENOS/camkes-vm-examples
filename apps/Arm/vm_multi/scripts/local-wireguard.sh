#!/bin/bash

# set -e

# echo "here1"
wg pubkey < ~/capsule-dev/camkes-vm-examples-manifest/projects/vm-examples/apps/Arm/vm_multi/network/example-wg-priv.txt

# echo "here2"
sudo ip link add dev wg0 type wireguard

# echo "here3"
sudo ip address add 10.198.10.1/32 dev wg0

# echo "here7"
sudo ip link set wg0 up

# echo "here4"
# sudo ip route add 10.198.10.113 dev wg0 proto kernel scope link src 10.198.10.1
sudo ip route add 10.198.10.0/24 dev wg0 proto kernel scope link src 10.198.10.1

# echo "here5"
sudo wg set wg0 listen-port 5861

# echo "here6"
sudo wg set wg0 private-key ~/capsule-dev/camkes-vm-examples-manifest/projects/vm-examples/apps/Arm/vm_multi/network/example-wg-priv.txt peer RSg/jyvPHmGtmt+8cHT/XwQfJbfYQydRCNJb/x+KjRY= allowed-ips 10.198.10.113 endpoint 10.10.10.113:5891

echo "Completed local wireguard setup"
