#!/usr/bin/env bash

root=`pwd`

mkdir -p $root/release/linux/bin
cp -r $root/build/bin/* $root/release/linux/bin
