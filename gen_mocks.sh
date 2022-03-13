#!/usr/bin/env bash
set -ex

for i in `find . -not -path '*/.*' -name mocks -type d`; do rm -rf $i; done
for d in `find . -not -path '*/.*' -type d` ; do mockery --output $d/mocks --all --case snake --dir "$d"; done
