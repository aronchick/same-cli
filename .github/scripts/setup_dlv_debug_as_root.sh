#!/bin/bash
echo "Probably don't execute. But if you want to, remove these two lines."
exit 0

# From here - https://github.com/microsoft/vscode-go/issues/2889#issuecomment-602122020

# Append delve to /etc/sudoers
export THIS_USER=userfoo
echo "$THIS_USER ALL=(root)NOPASSWD:/home/$THIS_USER/go/bin/dlv" >> /etc/sudoers

# Add to .vscode directory
cat << EOB
#!/bin/sh
if ! which dlv ; then
	PATH="${GOPATH}/bin:$PATH"
fi
if [ "$DEBUG_AS_ROOT" = "true" ]; then
	DLV=$(which dlv)
	exec sudo "$DLV" --only-same-user=false "$@"
else
	exec dlv "$@"
fi
EOB > .vscode/dlv-sudo.sh

# Add GOPATH to system
export GOPATH="$HOME/go"
export PATH="$PATH:$GOPATH/bin"
export PATH="$PATH:$HOME/.porter"
export GOBIN=$(go env GOPATH)/bin

# Add to settings.json
# "go.alternateTools": {
# 	"dlv": "${workspaceFolder}/.vscode/dlv-sudo.sh"
# }

# Add to launch.json
# {
#     "version": "0.2.0",
#     "configurations": [
#         {
#             "name": "Test as root",
#             "type": "go",
#             "request": "launch",
#             "mode": "test",
#             "program": "${fileDirname}",
#             "env": {
#                 "DEBUG_AS_ROOT": "true",
#             },
#         },
#     ],
#     "compounds": []
#  }
