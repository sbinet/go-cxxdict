#!/bin/sh

readlink=$((type -p greadlink readlink) | grep -v "not found" | head -1 | cut -d ' ' -f 3 )
if [ -z $readlink ]; then
    echo "** could not find 'readlink'"
    exit 1
fi

# Absolute path to this script, e.g. /home/user/bin/foo.sh
SCRIPT=`$readlink -f $0`
# Absolute path this script is in, thus /home/user/bin
SCRIPTPATH=`dirname $SCRIPT`

tests="basics cxx-classes cxx-stl cxx-oop"

function run() {
    cd $SCRIPTPATH
    sc=0
    echo ":: tests: $tests"
    for t in $tests
    do	
        (echo "   running test [$t]..." && 
	    cd $t &&         
	    /bin/rm -f test.log &&
            (./run.sh >& test.log || exit 1) &&
            echo "   running test [$t]... [OK]" &&
            cd ..)
	test_sc=$?
	if [ $test_sc -ne 0 ]; then
	    echo "** running test [$t]... [ERROR] **"
	    sc=1
	fi
    done;
    echo ":: tests: $tests [DONE]"
    return $sc
}

run || exit 1

