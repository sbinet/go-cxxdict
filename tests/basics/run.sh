#!/bin/sh

readlink=`which greadlink`
sc=$?
if [ $sc -ne 0 ]; then
    readlink=`which readlink`
    sc=$?
fi

if [ $sc -ne 0 ]; then
    echo "** could not find 'readlink'"
    exit $sc
fi

# Absolute path to this script, e.g. /home/user/bin/foo.sh
SCRIPT=`$readlink -f $0`
# Absolute path this script is in, thus /home/user/bin
SCRIPTPATH=`dirname $SCRIPT`

export CXX=${CXX-g++}
export GCCXML=${GCCXML-gccxml}
export GOCXXPKGROOT=`$readlink -f $SCRIPTPATH/../../../go-cxxdict`
export GOCXXDICTROOT=${GOCXXPKGROOT}/tests/basics
export GOCXXDICTTESTROOT=${GOCXXDICTROOT}/test
export GOPATH=${GOCXXDICTTESTROOT}/go:${GOPATH}
export LD_LIBRARY_PATH=${GOCXXDICTTESTROOT}/lib:${LD_LIBRARY_PATH}
export LD_LIBRARY_PATH=`pwd`:$LD_LIBRARY_PATH
export CGO_LDFLAGS=-L`pwd`

function clean_up() {
    /bin/rm -rf ${GOCXXDICTTESTROOT}
    /bin/rm -rf ${GOCXXDICTROOT}/{log,mylib.s}
}

function setup_dirs() {
    #clean_up || return 1
    mkdir -p ${GOCXXDICTTESTROOT} || return 1
    mkdir -p ${GOCXXDICTTESTROOT}/include || return 1
    mkdir -p ${GOCXXDICTTESTROOT}/lib || return 1
    mkdir -p ${GOCXXDICTTESTROOT}/go/{src/mylib,pkg,bin} || return 1
    /bin/ln -sfn \
        ${GOCXXDICTROOT}/mylib.hh \
        ${GOCXXDICTTESTROOT}/include/. || return 1
}

function compile_lib() {
    $CXX -O2 -shared -fPIC \
        -I${GOCXXDICTTESTROOT}/include \
        -o ${GOCXXDICTTESTROOT}/lib/libmylib.so \
        ${GOCXXDICTROOT}/mylib.cxx || return 1
}

function gen_cxxdict() {
    pushd ${GOCXXDICTTESTROOT}

    ${GCCXML} -fxml=out.xml \
        ${GOCXXDICTROOT}/mylib.hh \
        || return 1


    export CGO_CFLAGS="-I${GOCXXDICTTESTROOT}/include"
    export CGO_LDFLAGS="-L${GOCXXDICTTESTROOT}/lib"

    echo ":: go-gencxxinfos..."
    go-gencxxinfos \
        -fname ./out.xml \
        -libname mylib \
        -hdrname mylib.hh \
        || return 1

    echo ":: go-gencxxwrapper..."
    go-gencxxwrapper -fname ./ids.db || return 1
    gofmt -w . || return 1

    /bin/cp mylib_cxxgo.plugin.h ${GOCXXDICTTESTROOT}/include/.

    popd
}

function compile_cxxdict() {
    $CXX -O2 -shared -fPIC \
        -I${GOCXXDICTTESTROOT}/include \
        -o ${GOCXXDICTTESTROOT}/lib/libmylib_cxxgo.plugin.so \
        ${GOCXXDICTTESTROOT}/mylib_cxxgo.plugin.cxx \
        -L${GOCXXDICTTESTROOT}/lib \
        -lmylib \
        || return 1

}

function make_cxxdict() {
    gen_cxxdict || return 1
    compile_cxxdict || return 1

    cd ${GOCXXDICTTESTROOT}
    export CGO_CFLAGS="-I${GOCXXDICTTESTROOT}/include"
    export CGO_LDFLAGS="-L${GOCXXDICTTESTROOT}/lib"

    echo ":: go install mylib_cxxgo.plugin..."
    # go install -compiler gc \
    #     -o ${GOCXXDICTTESTROOT}/go/pkg/gocxx_mylib.a \
    #     mylib_cxxgo.plugin.go \
    #     || return 1

    # install go-pkg files
    /bin/cp mylib_cxxgo.plugin.go \
        ${GOCXXDICTTESTROOT}/go/src/mylib/. || return 1
    # compile
    pushd ${GOCXXDICTTESTROOT}/go/src/mylib
    CGO_CFLAGS="-I${GOCXXDICTTESTROOT}/include" \
    CGO_LDFLAGS="-L${GOCXXDICTTESTROOT}/lib" \
        go install . || return 1
    popd

}


function test_go_pkg() {

    echo "####"
    cd $GOCXXDICTTESTROOT
    go run ../gorun-test.go 2>&1 | tee chk.log || return 1
    echo "####"

    echo ":: checking ref.log..."
    diff -urN ../ref.log chk.log || return 1
    echo ":: checking ref.log... [ok]"

    return 0
}

function update_gocxx_pkgs() {
    cd ${GOCXXPKGROOT}/pkg/cxxtypes/gccxml
    go install . || return 1
    go test . || return 1

    cd ${GOCXXPKGROOT}/cmd/go-gencxxinfos
    go install . || return 1


    cd ${GOCXXPKGROOT}/cmd/go-gencxxwrapper
    go install . || return 1

}

function run() {
    echo ":::::::::::::::::::::::::::::::"
    echo ":: CXX: [$CXX]"
    echo ":: GOCXXDICTTESTROOT: [$GOCXXDICTTESTROOT]"
    echo ":: GOPATH: [${GOPATH}]"

    clean_up || return 1
    setup_dirs || return 1

    echo ":: ...update gocxx pkgs and cmds..."
    update_gocxx_pkgs || return 1

    echo ":: ...compiling c++ library..."
    compile_lib || return 1

    echo ":: ...generating c++ dictionary..."
    make_cxxdict || return 1

    echo ":: ...running gorun-test..."
    test_go_pkg || return 1

    echo ":: done."
    echo ":::::::::::::::::::::::::::::::"
    #clean_up
    return 0
}

run || exit 1

