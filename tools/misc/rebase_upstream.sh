#!/bin/bash
# Rebase from syndesisio/syndesis.git 1.10.x to jboss-fuse/syndesis 1.10.x
# if there is a conflict in assets_vfsdata.go use go to regenerate the file

git fetch https://github.com/syndesisio/syndesis.git 1.10.x:upstream
# if there is a previous rebase attempt, abort it
[[ -d .git/rebase-apply/ ]] && git rebase --abort
git rebase upstream &>/dev/null
if [ $? -ne 0 ]; then
    assets_vfsdata_conflict=$(git status --porcelain | grep -c 'UU install/operator/pkg/generator/assets_vfsdata.go')
    if [[ ${assets_vfsdata_conflict} == 1 ]] ; then
        if [[ ! -f "$GOROOT/bin/go" ]]; then
            echo "ERROR: go is required to regenerate the assets_vfsdata.go, but is not installed in \$GOROOT/bin/go. Env \$GOROOT: $GOROOT"
            exit 1
        fi
        echo "Regenerate install/operator/pkg/generator/assets_vfsdata.go"
        cd install/operator
        $GOROOT/bin/go generate -x ./pkg/generator/
        git add pkg/generator/assets_vfsdata.go
        cd ../..
        git rebase --continue &>/dev/null
        if [ $? -ne 0 ]; then
            echo "Could not rebase. The conflict must be manually resolved."
            git status
            exit 1
        fi
    else
        echo "Could not rebase. The conflict must be manually resolved."
        git status
        exit 1
    fi
fi
if [[ -d .git/rebase-apply/ ]]; then
    echo "ERROR: There are pending git conflicts to resolve."
    exit 1
else
    git push --force --set-upstream origin 1.10.x
fi
#git branch -D upstream
