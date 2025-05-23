#! /bin/bash


dnf install - y \
  vim \
  qemu-kvm \
  libvirt \
  virt-manager \
  virt-install \
  bridge-utils \
  cloud-utils \
  genisoimage \
  libguestfs-tools \
  virt-top \
  virt-viewer \
  cockpit


mkdir ~/vm-images/

