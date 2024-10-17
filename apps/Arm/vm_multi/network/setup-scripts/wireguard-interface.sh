#!/bin/sh

# set -e

InterfaceName="wg0"
PrivateKey="/root/wg0.conf"
Address="10.0.0.1/24"
AddressNoSubnet=""
ListenPort="5810"
node=""


# Help menu
show_help() {
    echo "Set up basic of a wireguard server. You will still need to allocate allowable peers."
    echo
    echo "Usage: $0 [options] InterfaceName PrivateKey Address"
    echo
    echo "Options:"
    echo "  --interface,   Name of the wireguard interface you'd like to create"
    echo "  --pk,          Private Key of the new wireguard interface you're creating"
    echo "  --addr,        Address that this node can be addressed with inside wireguard"
    echo "  --port,        Set the port the wireguard server should be listening on"
    echo "  --node,        Set the node you'd like to set up the interface on"
    echo "  --help,        Print this help menu"
    echo
}


# Parse CLI flag arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --interface)
        shift
        InterfaceName="$1"
        shift
        ;;
    --pk)
        shift
        PrivateKey="$1"
        shift
        ;;
    --addr)
        shift
        Address="$1"
        AddressNoSubnet="${Address%%/*}"
        shift
        ;;
    --port)
        shift
        ListenPort="$1"
        shift
        ;;
    --node)
        shift
        node=$1
        shift
        ;;
    --help)
        show_help
        exit 0
        ;;
    *)
        echo "Unknown cli entered"
        exit 0
  esac
done


# Get the subnet
set -- $(echo "$Address" | sed 's/[./]/ /g')
i1=$1; i2=$2; i3=$3; i4=$4; cidr=$5
mask=$((0xFFFFFFFF << (32 - cidr) & 0xFFFFFFFF))
Subnet=$(printf "%d.%d.%d.%d/%d\n" "$((i1 & (mask >> 24)))" "$((i2 & (mask >> 16 & 0xFF)))" "$((i3 & (mask >> 8 & 0xFF)))" "$((i4 & (mask & 0xFF)))" "$cidr")


# Set pubkey in wireguard
echo sshpass -p "root" ssh -o StrictHostKeyChecking=no "$node" "printf '%s' \"$PrivateKey\" | wg pubkey"
sshpass -p "root" ssh -o StrictHostKeyChecking=no "$node" "printf '%s' $PrivateKey | wg pubkey"

# Add a wireguard interface
echo sshpass -p "root" ssh -o StrictHostKeyChecking=no "$node" "ip link add dev $InterfaceName type wireguard"
sshpass -p "root" ssh -o StrictHostKeyChecking=no "$node" "ip link add dev $InterfaceName type wireguard"

# Add address for wireguard (used internally in wireguard comms)
echo sshpass -p "root" ssh -o StrictHostKeyChecking=no "$node" "ip address add $Address dev $InterfaceName"
sshpass -p "root" ssh -o StrictHostKeyChecking=no "$node" "ip address add $Address dev $InterfaceName"

# Set interface UP
echo sshpass -p "root" ssh -o StrictHostKeyChecking=no "$node" "ip link set $InterfaceName up"
sshpass -p "root" ssh -o StrictHostKeyChecking=no "$node" "ip link set $InterfaceName up"

# Set up route for easy resolution
echo sshpass -p "root" ssh -o StrictHostKeyChecking=no "$node" "ip route add $Subnet dev wg0 proto kernel scope link src $AddressNoSubnet"
sshpass -p "root" ssh -o StrictHostKeyChecking=no "$node" "ip route add $Subnet dev wg0 proto kernel scope link src $AddressNoSubnet"

# Set the priv key from wireguard instance
echo sshpass -p "root" ssh -o StrictHostKeyChecking=no "$node" "printf '%s' \"$PrivateKey\" | wg set $InterfaceName private-key /dev/stdin"
sshpass -p "root" ssh -o StrictHostKeyChecking=no "$node" "printf '%s' \"$PrivateKey\" | wg set $InterfaceName private-key /dev/stdin"

# Specify port inbound communication needs to come from
echo sshpass -p "root" ssh -o StrictHostKeyChecking=no "$node" "wg set $InterfaceName listen-port $ListenPort"
sshpass -p "root" ssh -o StrictHostKeyChecking=no "$node" "wg set $InterfaceName listen-port $ListenPort"

echo "Wireface setup complete"

