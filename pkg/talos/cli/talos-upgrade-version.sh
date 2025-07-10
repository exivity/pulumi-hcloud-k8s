#!/bin/bash

# if ARM_IMAGE or X86_IMAGE are NODE_IMAGE, then exit 0
if [ "$ARM_IMAGE" == "$NODE_IMAGE" ] || [ "$X86_IMAGE" == "$NODE_IMAGE" ]; then
  exit 0
fi

# Write Talos configuration to a file
echo $TALOSCONFIG_VALUE > $TALOSCONFIG

# The --preserve flag is important as it ensures that ephemeral data on the node is kept intact during the upgrade process.
if ! talosctl upgrade --nodes $NODE_IP --image factory.talos.dev/installer/$TALOS_IMAGE:$TALOS_VERSION --preserve --timeout 10m; then
  echo "ERROR: Talos upgrade failed for node $NODE_IP" >&2
  exit 1
fi
