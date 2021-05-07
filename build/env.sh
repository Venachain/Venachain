#!/bin/sh

set -e

if [ ! -f "build/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi

# Create fake Go workspace if it doesn't exist yet.
workspace="$PWD/build/_workspace"
root="$PWD"

echo "$root" "$workspace"

platonedir="$workspace/src/github.com/PlatONEnetwork"
if [ ! -L "$platonedir/PlatONE-Go" ]; then
    mkdir -p "$platonedir"
    cd "$platonedir"
    ln -s ../../../../../. PlatONE-Go
    echo "ln -s success."
    cd "$root"
fi

# Set up the environment to use the workspace.
GOPATH="$workspace"
export GOPATH
export PATH=$PATH:$GOPATH/bin

# Run the command inside the workspace.
cd "$platonedir/PlatONE-Go"
PWD="$platonedir/PlatONE-Go"

# Launch the arguments with the configured environment.
exec "$@"
