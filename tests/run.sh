#!/bin/sh

# Absolute path to this script, e.g. /home/user/bin/foo.sh
SCRIPT=`readlink -f $0`
# Absolute path this script is in, thus /home/user/bin
SCRIPTPATH=`dirname $SCRIPT`

tests="basics cxx-oop"

function run() {
    cd $SCRIPTPATH
    echo ":: tests: $tests"
    for t in $tests
    do
        cd $t
        echo ":: running test [$t]..."
        ./run.sh >& /dev/null || exit 1
        echo ":: running test [$t]...[ok]"
        cd ..
    done;
}

run || exit 1

