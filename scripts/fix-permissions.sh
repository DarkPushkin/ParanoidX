#!/bin/bash
set -euo pipefail
echo "=== Fixing simplex-node permissions ==="
chown -R tomas:tomas /home/tomas/.local/share/simplex-node
chown -R tomas:tomas /home/tomas/simplex-node/docker
for d in /home/tomas/.local/share/simplex-node-A*; do
  [ -d "$d" ] && chown -R tomas:tomas "$d"
done
echo "Done. Now run: bash /home/tomas/simplex-node/scripts/launch-node.sh"
