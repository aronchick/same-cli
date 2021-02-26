# Copyright © 2021 The SAME Authors

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

#    http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Convenient script to handle dependencies.
#
# Kudo to https://github.com/knative/hack/blob/master/library.sh


# Remove symlinks in a path that are broken or lead outside the repo.
# Parameters: $1 - path name, e.g. vendor
function remove_broken_symlinks() {
  for link in $(find $1 -type l); do
    # Remove broken symlinks
    if [[ ! -e ${link} ]]; then
      unlink ${link}
      continue
    fi
    # Get canonical path to target, remove if outside the repo
    local target="$(ls -l ${link})"
    target="${target##* -> }"
    [[ ${target} == /* ]] || target="./${target}"
    target="$(cd `dirname "${link}"` && cd "${target%/*}" && echo "$PWD"/"${target##*/}")"
    if [[ ${target} != *github.com/knative/* && ${target} != *knative.dev/* ]]; then
      unlink "${link}"
      continue
    fi
  done
}

# Run a go tool, installing it first if necessary.
# Parameters: $1 - tool package/dir for go get/install.
#             $2 - tool to run.
#             $3..$n - parameters passed to the tool.
function run_go_tool() {
  local tool=$2
  local install_failed=0
  if [[ -z "$(which ${tool})" ]]; then
    local action=get
    [[ $1 =~ ^[\./].* ]] && action=install
    # Avoid running `go get` from root dir of the repository, as it can change go.sum and go.mod files.
    # See discussions in https://github.com/golang/go/issues/27643.
    if [[ ${action} == "get" && $(pwd) == "${REPO_ROOT_DIR}" ]]; then
      local temp_dir="$(mktemp -d)"
      # Swallow the output as we are returning the stdout in the end.
      pushd "${temp_dir}" > /dev/null 2>&1
      GOFLAGS="" go ${action} "$1" || install_failed=1
      popd > /dev/null 2>&1
    else
      GOFLAGS="" go ${action} "$1" || install_failed=1
    fi
  fi
  (( install_failed )) && return ${install_failed}
  shift 2
  ${tool} "$@"
}

# Run go-licenses to update licenses.
# Parameters: $1 - output file, relative to repo root dir.
#             $2 - directory to inspect.
function update_licenses() {
  cd "${REPO_ROOT_DIR}" || return 1
  local dst=$1
  local dir=$2
  shift
  run_go_tool github.com/google/go-licenses go-licenses save "${dir}" --save_path="${dst}" --force || \
    { echo "--- FAIL: go-licenses failed to update licenses"; return 1; }
  # Hack to make sure directories retain write permissions after save. This
  # can happen if the directory being copied is a Go module.
  # See https://github.com/google/go-licenses/issues/11
  chmod -R +w "${dst}"
}

export GO111MODULE=on
export GOFLAGS=""

# Go mod tidy and vendor
orig_pipefail_opt=$(shopt -p -o pipefail)
set -o pipefail
go mod tidy 2>&1 | grep -v "ignoring symlink" || true
go mod vendor 2>&1 |  grep -v "ignoring symlink" || true
eval "$orig_pipefail_opt"


# Updating licenses
update_licenses third_party/VENDOR-LICENSE "./..."

# Removing broken symlinks
remove_broken_symlinks ./vendor
