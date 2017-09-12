#!/usr/bin/env bash

DBG_SHOW=1
# Debug-level for server
DBG_SRV=1
# Debug-level for client
DBG_CLIENT=1
# For easier debugging
BUILDDIR=$(pwd)
STATICDIR=test

. $GOPATH/src/gopkg.in/dedis/onet.v1/app/libtest.sh

main(){
    startTest
    test Build
    test ServerCfg
    test RunI2B2dc
    clearSrv
}


#------- BUILD --------#
testBuild(){
    testOK dbgRun ./i2b2dc --help
}

build(){
    if [ "$STATICDIR" ]; then
        DIR=$STATICDIR
    else
        DIR=$(mktemp -d)
    fi

    rm -f $DIR/i2b2dc #to force compilation

    mkdir -p $DIR
    cd $DIR
    echo "Building in $DIR"

    if [ ! -x i2b2dc ]; then
        go build -o i2b2dc -a $BUILDDIR/*go
    fi

}


#------- SERVER CONFIGURATION --------#
testServerCfg(){
    for ((n=1; n <= 2*2; n+=2)) do
        runSrvCfg $n
        pkill -9 i2b2dc
        testFile srv$n/private.toml
    done
}

runSrvCfg(){
    echo -e "127.0.0.1:200$1\ni2b2dc $1\n$(pwd)/srv$1\n" | ./i2b2dc server setup > $OUT
}


#------- CLIENT CONFIGURATION --------#
testRunI2B2dc(){
    setupServers
    echo "Running i2b2dc APP"
    runCl 1 run
}

setupServers(){
    rm -f group.toml
    for ((n=1; n <= 2*2; n+=2)) do
        srv=srv$n
        rm -f $srv/*
        runSrvCfg $n
        tail -n 4 $srv/public.toml >> group.toml
        cp ../db.toml $srv/

        runSrv $n &
    done


}

runSrv(){
    cd srv$1
    ../i2b2dc -d $DBG_SRV server -c private.toml
    cd ..
}

runCl(){
    G=group.toml
    shift
    echo "Running Client with $G $@"
    ./i2b2dc -d $DBG_CLIENT $@ -f $G -c "ICD10:E08.52,ICD10:E08.59" -t "2015"  -g "location_cd"
}


#------- CLEAR SERVERS --------#
clearSrv(){
    pkill -9 i2b2dc
}


#------- OTHER STUFF --------#
if [ "$1" == "-q" ]; then
  DBG_RUN=
  STATICDIR=
fi

main
