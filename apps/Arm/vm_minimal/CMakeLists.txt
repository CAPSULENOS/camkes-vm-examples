#
# Copyright 2018, Data61, CSIRO (ABN 41 687 119 230)
#
# SPDX-License-Identifier: BSD-2-Clause
#

cmake_minimum_required(VERSION 3.8.2)

project(camkes-arm-tk1-vm C)

include(${CAMKES_ARM_VM_HELPERS_PATH})

set(cpp_includes "${CAMKES_VM_DIR}/components/VM_Arm")

# Create our CPP Flags based on ARM VM config variables
if("${KernelARMPlatform}" STREQUAL "tk1")
    set(cpp_flags "-DKERNELARMPLATFORM_TK1")

    # Different device tree if compiling with Tk1Insecure
    if(Tk1Insecure)
        set(device_tree_src "linux-tk1-nonsecured.dts")
    else()
        set(device_tree_src "linux-tk1-secure.dts")
    endif()

    if(VmTk1InitrdRootfs)
        set(linux_binary_name "linux-tk1-initrd")
    else()
        set(linux_binary_name "linux-tk1-debian")
    endif()
    add_custom_command(
        OUTPUT linux/linux-dtb
        COMMAND
            dtc -I dts -O dtb -o linux/linux-dtb ${CAMKES_ARM_LINUX_DIR}/${device_tree_src}
        VERBATIM
        DEPENDS ${CAMKES_ARM_LINUX_DIR}/${device_tree_src}
    )
    # Create custom target for setting the dtb
    add_custom_target(set_dtb DEPENDS "${CMAKE_CURRENT_BINARY_DIR}/linux/linux-dtb")

    # Add Linux VM images into file server
    AddToFileServer("linux" "${CAMKES_ARM_LINUX_DIR}/${linux_binary_name}")
    AddToFileServer("linux-dtb" "${CMAKE_CURRENT_BINARY_DIR}/linux/linux-dtb" DEPENDS set_dtb)

elseif("${KernelARMPlatform}" STREQUAL "tx1")
    set(cpp_flags "-DKERNELARMPLATFORM_TX1")

    AddToFileServer("linux" "${CAMKES_VM_IMAGES_DIR}/tx1/linux")
    AddToFileServer("linux-dtb" "${CAMKES_VM_IMAGES_DIR}/tx1/linux-dtb")

elseif("${KernelARMPlatform}" STREQUAL "exynos5422")
    find_package(camkes-vm-linux REQUIRED)
    include(${CAMKES_VM_LINUX_HELPERS_PATH})
    set(cpp_flags "-DKERNELARMPLATFORM_EXYNOS5422")

    AddToFileServer("linux" "${CAMKES_VM_IMAGES_DIR}/exynos5422/linux")
    # Generate overlayed rootfs
    set(rootfs_file "${CAMKES_VM_IMAGES_DIR}/exynos5422/rootfs.cpio.gz")
    AddToFileServer("linux-initrd" ${rootfs_file})
    # Updated dtb based on initrd
    UpdateDtbFromInitrd(
        "${CAMKES_VM_IMAGES_DIR}/exynos5422/linux-dtb"
        ${rootfs_file}
        "0x4d700000"
        dtb_gen_target
        output_dtb_location
    )
    AddToFileServer("linux-dtb" "${output_dtb_location}" DEPENDS dtb_gen_target)

elseif("${KernelARMPlatform}" STREQUAL "tx2")
    find_package(camkes-vm-linux REQUIRED)
    include(${CAMKES_VM_LINUX_HELPERS_PATH})
    set(cpp_flags "-DKERNELARMPLATFORM_TX2")

    AddToFileServer("linux" "${CAMKES_VM_IMAGES_DIR}/tx2/linux")
    set(rootfs_file "${CAMKES_VM_IMAGES_DIR}/tx2/rootfs.cpio.gz")
    AddToFileServer("linux-initrd" ${rootfs_file})
    # Updated dtb based on initrd
    UpdateDtbFromInitrd(
        "${CAMKES_VM_IMAGES_DIR}/tx2/linux-dtb"
        ${rootfs_file}
        "0xf7000000"
        dtb_gen_target
        output_dtb_location
    )
    AddToFileServer("linux-dtb" "${output_dtb_location}" DEPENDS dtb_gen_target)

elseif("${KernelARMPlatform}" STREQUAL "qemu-arm-virt")

    # QEMU supports ARMv7/aarch32 and ARMv8/aarch32, but currently there are
    # Linux images for aarch64 only. However, it is unclear if that is the only
    # blocker to get this example run in QEMU on aarch32.
    if(NOT KernelSel4ArchAarch64)
        message(FATAL_ERROR "Only AARCH64 is supported")
    endif()

    find_package(camkes-vm-linux REQUIRED)
    include(${CAMKES_VM_LINUX_HELPERS_PATH})
    set(cpp_flags "-DKERNELARMPLATFORM_QEMU-ARM-VIRT")

    AddToFileServer("linux" "${CAMKES_VM_IMAGES_DIR}/qemu-arm-virt/linux")
    set(rootfs_file "${CAMKES_VM_IMAGES_DIR}/qemu-arm-virt/rootfs.cpio.gz")
    AddToFileServer("linux-initrd" ${rootfs_file})
    # Updated dtb based on initrd
    UpdateDtbFromInitrd(
        "${CAMKES_VM_IMAGES_DIR}/qemu-arm-virt/linux-dtb"
        ${rootfs_file}
        "0x4d700000"
        dtb_gen_target
        output_dtb_location
    )
    AddToFileServer("linux-dtb" "${output_dtb_location}" DEPENDS dtb_gen_target)
    include(simulation)
    set(SIMULATION ON CACHE BOOL "Generate simulation script to run qemu with the proper arguments")
    if(SIMULATION)
        GenerateSimulateScript()
    endif()

elseif("${KernelARMPlatform}" STREQUAL "odroidc2")
    find_package(camkes-vm-linux REQUIRED)
    include(${CAMKES_VM_LINUX_HELPERS_PATH})
    set(cpp_flags "-DKERNELARMPLATFORM_ODROIDC2")

    set(
        CAMKES_ROOT_DTB_FILE_PATH "${CAMKES_VM_IMAGES_DIR}/odroidc2/camkes-linux-dtb"
        CACHE STRING ""
    )
    AddToFileServer("linux" "${CAMKES_VM_IMAGES_DIR}/odroidc2/linux")
    set(rootfs_file "${CAMKES_VM_IMAGES_DIR}/odroidc2/rootfs.cpio.gz")
    AddToFileServer("linux-initrd" ${rootfs_file})
    # Updated dtb based on initrd
    UpdateDtbFromInitrd(
        "${CAMKES_VM_IMAGES_DIR}/odroidc2/linux-dtb"
        ${rootfs_file}
        "0x2d700000"
        dtb_gen_target
        output_dtb_location
    )
    AddToFileServer("linux-dtb" "${output_dtb_location}" DEPENDS dtb_gen_target)

elseif("${KernelARMPlatform}" STREQUAL "zcu102")

    if("${VmZynqmpPetalinuxVersion}" STREQUAL "")
        message(
            FATAL_ERROR
                "VmZynqmpPetalinuxVersion is not set. Please set to include proper VM artifacts"
        )
    endif()

    AddToFileServer("linux" "${CAMKES_VM_IMAGES_DIR}/zynqmp/${VmZynqmpPetalinuxVersion}/linux")
    AddToFileServer(
        "linux-initrd"
        "${CAMKES_VM_IMAGES_DIR}/zynqmp/${VmZynqmpPetalinuxVersion}/rootfs-minimal.cpio.gz"
    )

    set(dts_file "${CAMKES_VM_IMAGES_DIR}/zynqmp/${VmZynqmpPetalinuxVersion}/linux.dts")

    add_custom_command(
        OUTPUT linux/linux-dtb
        COMMAND
            dtc -q -I dts -O dtb -o linux/linux-dtb ${dts_file}
        VERBATIM
        DEPENDS ${dts_file}
    )

    # Create custom target for setting the dtb
    add_custom_target(set_dtb DEPENDS "${CMAKE_CURRENT_BINARY_DIR}/linux/linux-dtb")
    AddToFileServer("linux-dtb" "${CMAKE_CURRENT_BINARY_DIR}/linux/linux-dtb" DEPENDS set_dtb)

    list(APPEND cpp_includes "${CAMKES_VM_DIR}/components/VM_Arm/plat_include/zynqmp")

    CAmkESAddImportPath(${KernelARMPlatform}/${VmZynqmpPetalinuxVersion})
endif()

AddCamkesCPPFlag(cpp_flags CONFIG_VARS VmEmmc2NoDMA VmVUSB Tk1DeviceFwd Tk1Insecure)

DefineCAmkESVMFileServer()

CAmkESAddImportPath(${KernelARMPlatform})

# Declare root server
DeclareCAmkESRootserver(vm_minimal.camkes CPP_FLAGS ${cpp_flags} CPP_INCLUDES ${cpp_includes})
