#!/bin/sh

readlink=`which greadlink`
sc=$?
if [ $sc -ne 0 ]; then
    readlink=`which readlink`
fi

if [ $sc -ne 0 ]; then
    echo "** could not find 'readlink'"
    exit $sc
fi

# Absolute path to this script, e.g. /home/user/bin/foo.sh
SCRIPT=`$readlink -f $0`
# Absolute path this script is in, thus /home/user/bin
SCRIPTPATH=`dirname $SCRIPT`

tests="basics cxx-classes cxx-stl cxx-oop"

function run() {
    cd $SCRIPTPATH
    echo ":: tests: $tests"
    for t in $tests
    do
	echo ":: running test [$t]..."
        (cd $t &&            
            (./run.sh >& /dev/null || exit 1) &&
            echo ":: running test [$t]...[ok]" &&
            cd ..) || \
            echo ":: running test [$t]...[err]"
    done;
}

run || exit 1

