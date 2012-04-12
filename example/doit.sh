#!/bin/sh

export CXX=${CXX-g++}
export PYTHON=${PYTHON-python2}
export GOCXXDICTTESTROOT=${TMPDIR-/tmp}/go-cxxdict-test
export GOPATH=${GOCXXDICTTESTROOT}/go:${GOPATH}
export LD_LIBRARY_PATH=${GOCXXDICTTESTROOT}/lib:${LD_LIBRARY_PATH}

function clean_up() {
    /bin/rm -rf ${GOCXXDICTTESTROOT}
    /bin/rm -rf mylib_cxxdict.cxx
    /bin/rm -rf mylib_gccxmlout.xml
    /bin/rm -rf dump.xml
    /bin/rm -rf mylibpkg.cxx
    /bin/rm -rf mylibpkg.h
    /bin/rm -rf mylibpkg.go
    /bin/rm -rf libmylib_cxxdict.dsomap
}

function setup_dirs() {
    #clean_up || return 1
    mkdir -p ${GOCXXDICTTESTROOT}/include || return 1
    mkdir -p ${GOCXXDICTTESTROOT}/lib || return 1
    mkdir -p ${GOCXXDICTTESTROOT}/go/{src/mylibpkg,pkg,bin} || return 1
    /bin/ln -sfn `pwd`/mylib.h ${GOCXXDICTTESTROOT}/include/. || return 1
}

function compile_lib() {
    $CXX -O2 -shared -fPIC \
        -I${GOCXXDICTTESTROOT}/include \
        -o ${GOCXXDICTTESTROOT}/lib/libmylib.so \
        mylib.cxx || return 1
}

function gen_cxxdict() {
    export PYTHONPATH=`pwd`/../genreflex:${PYTHONPATH}
    GENREFLEX="${PYTHON} ../genreflex/genreflex.py"
    ${GENREFLEX} mylib.h -s ./sel.xml \
        --output mylib_cxxdict.cxx \
        --package mylibpkg \
        --debug \
        --fail_on_warnings \
        --gccxmlopt=-m64 \
        -I${GOCXXDICTTESTROOT}/include \
        -DNDEBUG -D__REFLEX__ \
        --rootmap=libmylib_cxxdict.dsomap \
        --rootmap-lib=libmylib_cxxdict.so \
        || return 1

    # install files...
    /bin/ln -sfn `pwd`/mylibpkg.h ${GOCXXDICTTESTROOT}/include/. || return 1
    /bin/ln -sfn `pwd`/libmylib_cxxdict.dsomap ${GOCXXDICTTESTROOT}/lib/. || return 1
}

function compile_cxxdict() {
    $CXX -O2 -shared -fPIC \
        -I${GOCXXDICTTESTROOT}/include \
        -o ${GOCXXDICTTESTROOT}/lib/libmylibpkg.so \
        mylibpkg.cxx || return 1

}

function make_cxxdict() {
    gen_cxxdict || return 1
    compile_cxxdict || return 1
}


function test_go_pkg() {
    # install go-pkg files
    /bin/ln -sfn `pwd`/mylibpkg.go ${GOCXXDICTTESTROOT}/go/src/mylibpkg/. || return 1
    # compile
    pushd ${GOCXXDICTTESTROOT}/go/src/mylibpkg
    CGO_CFLAGS="-I${GOCXXDICTTESTROOT}/include" \
    CGO_LDFLAGS="-L${GOCXXDICTTESTROOT}/lib" \
        go install . || return 1
    popd || return 1

    echo "####"
    go run ./gorun-test.go |& tee chk.log || return 1
    echo "####"

    echo ":: checking ref.log..."
    diff -urN ref.log chk.log || return 1
    echo ":: checking ref.log... [ok]"

    return 0
}

function run() {
    echo ":::::::::::::::::::::::::::::::"
    echo ":: CXX: [$CXX]"
    echo ":: PYTHON: [$PYTHON]"
    echo ":: GOCXXDICTTESTROOT: [$GOCXXDICTTESTROOT]"
    
    clean_up || return 1
    setup_dirs || return 1

    echo ":: ...compiling c++ library..."
    compile_lib || return 1

    echo ":: ...generating c++ dictionary..."
    make_cxxdict || return 1

    echo ":: ...running gorun-test..."
    test_go_pkg

    echo ":: done."
    echo ":::::::::::::::::::::::::::::::"
    clean_up
    return 0
}

run

