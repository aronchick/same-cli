#!/bin/sh
export GOPATH="/home/daaronch/go"
export PATH=$PATH:$GOPATH/bin
if ! which dlv ; then
	PATH="${GOPATH}/bin:$PATH"
fi
if [ "$DEBUG_AS_ROOT" = "true" ]; then
	DLV=$(which dlv)
	exec sudo -E "$DLV" --only-same-user=false "$@"
else
	exec dlv "$@"
fi