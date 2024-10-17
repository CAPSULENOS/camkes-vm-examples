#
# Copyright 2020, Data61, CSIRO (ABN 41 687 119 230)
#
# SPDX-License-Identifier: BSD-2-Clause
#

cmake_minimum_required(VERSION 3.8.2)

project(vm-multi C)

include(${CAMKES_ARM_VM_HELPERS_PATH})
find_package(camkes-vm-linux REQUIRED)
include(${CAMKES_VM_LINUX_HELPERS_PATH})

if("${PLATFORM}" STREQUAL "exynos5422")
# Create our CPP Flags based on ARM VM config variables
    set(cpp_flags "-DKERNELARMPLATFORM_EXYNOS5422")
elseif("${PLATFORM}" STREQUAL "qemu-arm-virt")
    set(cpp_flags "-DKERNELARMPLATFORM_QEMU-ARM-VIRT")
    set(SIMULATION ON CACHE BOOL "Generate simulation script to run qemu with the proper arguments")
    include(simulation)
    if(SIMULATION)
        GenerateSimulateScript()
    endif()
endif()

# Generate our overlayed rootfs images
# set(rootfs_file "${CAMKES_VM_IMAGES_DIR}/${KernelARMPlatform}/rootfs.cpio.gz")
set(vm0_rootfs_file "${CAMKES_VM_IMAGES_DIR}/${KernelARMPlatform}/vm0-rootfs.cpio.gz")
set(vm1_rootfs_file "${CAMKES_VM_IMAGES_DIR}/${KernelARMPlatform}/vm1-rootfs.cpio.gz")
set(vm2_rootfs_file "${CAMKES_VM_IMAGES_DIR}/${KernelARMPlatform}/vm2-rootfs.cpio.gz")
set(vm3_rootfs_file "${CAMKES_VM_IMAGES_DIR}/${KernelARMPlatform}/vm3-rootfs.cpio.gz")




AddFileToOverlayDir(
    "S90bridge_setup"
    ${CMAKE_CURRENT_SOURCE_DIR}/overlay_files/vm0_bridge_setup.sh
    "etc/init.d"
    overlay_vm0
)
# Add all the materials for aegis
AddFileToOverlayDir(
    "wireguard-interface.sh"
    ${CMAKE_CURRENT_SOURCE_DIR}/network/setup-scripts/wireguard-interface.sh
    "etc/aegis/scripts"
    overlay_vm0
)
AddFileToOverlayDir(
    "set-default-route"
    ${CMAKE_CURRENT_SOURCE_DIR}/network/setup-scripts/set-default-route
    "etc/aegis/scripts"
    overlay_vm0
)
AddFileToOverlayDir(
    "generate-client-vlan"
    ${CMAKE_CURRENT_SOURCE_DIR}/network/setup-scripts/generate-client-vlan
    "etc/aegis/scripts"
    overlay_vm0
)
AddFileToOverlayDir(
    "iptables-raw-forwarding"
    ${CMAKE_CURRENT_SOURCE_DIR}/network/setup-scripts/iptables-raw-forwarding
    "etc/aegis/scripts"
    overlay_vm0
)
AddFileToOverlayDir(
    "connect-vlans"
    ${CMAKE_CURRENT_SOURCE_DIR}/network/setup-scripts/connect-vlans
    "etc/aegis/scripts"
    overlay_vm0
)
AddFileToOverlayDir(
    "remote-etc-hosts"
    ${CMAKE_CURRENT_SOURCE_DIR}/network/setup-scripts/remote-etc-hosts
    "etc/aegis/scripts"
    overlay_vm0
)
AddFileToOverlayDir(
    "silent"
    ${CMAKE_CURRENT_SOURCE_DIR}/network/vnfs/silent
    "etc/aegis/scripts"
    overlay_vm0
)
AddFileToOverlayDir(
    "add-ip-forward"
    ${CMAKE_CURRENT_SOURCE_DIR}/network/vnfs/add-ip-forward
    "etc/aegis/scripts"
    overlay_vm0
)
AddFileToOverlayDir(
    "add-route"
    ${CMAKE_CURRENT_SOURCE_DIR}/network/vnfs/add-route
    "etc/aegis/scripts"
    overlay_vm0
)
AddFileToOverlayDir(
    "aegis"
    ${CMAKE_CURRENT_SOURCE_DIR}/network/aegis
    "root"
    overlay_vm0
)
AddFileToOverlayDir(
    "aegis-conf.yml"
    ${CMAKE_CURRENT_SOURCE_DIR}/network/aegis-conf.yml
    "etc/aegis"
    overlay_vm0
)
AddOverlayDirToRootfs(
    overlay_vm0
    ${vm0_rootfs_file}
    "buildroot"
    "rootfs_install"
    vm0_output_overlayed_rootfs_location
    rootfs_target_vm0
    GZIP
)





# Files needed for vm1 overlay
AddFileToOverlayDir(
    "arp-forwarding"
    ${CMAKE_CURRENT_SOURCE_DIR}/network/vnfs/arp-forwarding
    "etc/aegis/scripts"
    overlay_client_vm1
)
AddFileToOverlayDir(
    "clean"
    ${CMAKE_CURRENT_SOURCE_DIR}/network/vnfs/clean
    "etc/aegis/scripts"
    overlay_client_vm1
)
AddFileToOverlayDir(
    "S90client_net_vm"
    ${CMAKE_CURRENT_SOURCE_DIR}/overlay_files/vm_client_net_setup.sh
    "etc/init.d"
    overlay_client_vm1
)
AddOverlayDirToRootfs(
    overlay_client_vm1
    ${vm1_rootfs_file}
    "buildroot"
    "rootfs_install"
    vm1_output_overlayed_rootfs_location
    rootfs_target_client_vm1
    GZIP
)



# Files needed for vm2 overlay
AddFileToOverlayDir(
    "arp-forwarding"
    ${CMAKE_CURRENT_SOURCE_DIR}/network/vnfs/arp-forwarding
    "etc/aegis/scripts"
    overlay_client_vm2
)
AddFileToOverlayDir(
    "clean"
    ${CMAKE_CURRENT_SOURCE_DIR}/network/vnfs/clean
    "etc/aegis/scripts"
    overlay_client_vm2
)
AddFileToOverlayDir(
    "S90client_net_vm"
    ${CMAKE_CURRENT_SOURCE_DIR}/overlay_files/vm_client_net_setup.sh
    "etc/init.d"
    overlay_client_vm2
)
AddOverlayDirToRootfs(
    overlay_client_vm2
    ${vm2_rootfs_file}
    "buildroot"
    "rootfs_install"
    vm2_output_overlayed_rootfs_location
    rootfs_target_client_vm2
    GZIP
)



AddFileToOverlayDir(
    "arp-forwarding"
    ${CMAKE_CURRENT_SOURCE_DIR}/network/vnfs/arp-forwarding
    "etc/aegis/scripts"
    overlay_client_vm3
)
AddFileToOverlayDir(
    "clean"
    ${CMAKE_CURRENT_SOURCE_DIR}/network/vnfs/clean
    "etc/aegis/scripts"
    overlay_client_vm3
)
AddFileToOverlayDir(
    "ipsec.secrets"
    ${CMAKE_CURRENT_SOURCE_DIR}/network/ipsec.secrets
    "etc"
    overlay_client_vm3
)
AddFileToOverlayDir(
    "S90client_net_vm"
    ${CMAKE_CURRENT_SOURCE_DIR}/overlay_files/vm_client_net_setup.sh
    "etc/init.d"
    overlay_client_vm3
)
AddOverlayDirToRootfs(
    overlay_client_vm3
    ${vm3_rootfs_file}
    "buildroot"
    "rootfs_install"
    vm3_output_overlayed_rootfs_location
    rootfs_target_client_vm3
    GZIP
)





AddToFileServer(
    "linux-initrd-vm0"
    "${vm0_output_overlayed_rootfs_location}"
    DEPENDS
    rootfs_target_vm0
)



# AddToFileServer(
#     "linux" 
#     "${CAMKES_VM_IMAGES_DIR}/${KernelARMPlatform}/linux"
#     DEPENDS
#     rootfs_target_client_vm
# )

AddToFileServer(
    "linux-initrd-vm1-client"
    "${vm1_output_overlayed_rootfs_location}"
    DEPENDS
    rootfs_target_client_vm1
)

AddToFileServer(
    "linux-initrd-vm2-client"
    "${vm2_output_overlayed_rootfs_location}"
    DEPENDS
    rootfs_target_client_vm2
)

AddToFileServer(
    "linux-initrd-vm3-client"
    "${vm3_output_overlayed_rootfs_location}"
    DEPENDS
    rootfs_target_client_vm3
)

AddToFileServer("linux" "${CAMKES_VM_IMAGES_DIR}/${KernelARMPlatform}/linux")


AddCamkesCPPFlag(cpp_flags CONFIG_VARS VmEmmc2NoDMA VmVUSB VmVchan Tk1DeviceFwd Tk1Insecure)

DefineCAmkESVMFileServer()

CAmkESAddImportPath(${KernelARMPlatform})

# add CAPSULE_MAC_PREFIX
if (CAPSULE_MAC_PREFIX)
    list(APPEND cpp_flags "-DCAPSULE_MAC_PREFIX=${CAPSULE_MAC_PREFIX}")
endif()

# add CAPSULE_NODE
if (CAPSULE_NODE)
    list(APPEND cpp_flags "-DCAPSULE_NODE=${CAPSULE_NODE}")
endif()

# Declare root server
DeclareCAmkESRootserver(
    vm_multi.camkes
    CPP_FLAGS
    ${cpp_flags}
    CPP_INCLUDES
    ${CAMKES_VM_DIR}/components/VM_Arm
)
