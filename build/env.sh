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

venachaindir="$workspace/src/github.com/Venachain"
if [ ! -L "$venachaindir/Venachain" ]; then
    mkdir -p "$venachaindir"
    cd "$venachaindir"
    ln -s ../../../../../. Venachain
    echo "ln -s success."
    cd "$root"
fi

# Set up the environment to use the workspace.
GOPATH="$workspace"
export GOPATH
export PATH=$PATH:$GOPATH/bin

# Run the command inside the workspace.
cd "$venachaindir/Venachain"
PWD="$venachaindir/Venachain"

# Launch the arguments with the configured environment.
exec "$@"
