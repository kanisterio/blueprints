#!/usr/bin/env bash
# Copyright 2025 The Kanister Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace


usage() {
    echo ./bump_version.sh previous_version release_version
    exit 1
}

main() {

    local prev=${1:?"$(usage)"}; shift
    local next=${1:?"$(usage)"}; shift
    if [ "$#" -eq 0 ]; then
            pkgs=( . )
        else
            pkgs=( "$@" )
    fi
    
    # -F matches for exact words, not regular expression (-E)
    local files_to_update
    files_to_update=$(grep -F "${prev}" -Ir "${pkgs[@]}" \
        --exclude-dir={mod,bin,html,frontend} \
        --exclude=\*.sum --exclude=\*.mod \
        | cut -d ':' -f 1 | sort | uniq)
    
    if [ -z "$files_to_update" ]; then
        echo "No files found containing version: ${prev}"
        return 0
    fi
    
    # Update files using while loop to handle filenames with special characters
    while IFS= read -r file; do
        sed -i.bak "s/${prev//./\\.}/${next//./\\.}/g" "$file"
        rm -f "$file.bak"
    done <<< "$files_to_update"
}

main $@

