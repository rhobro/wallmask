# env
export DB2HOME=/Users/mavicmaverick/go/src/github.com/ibmdb/go_ibm_db/installer/clidriver
export CGO_CFLAGS=-I$DB2HOME/include
export CGO_LDFLAGS=-L$DB2HOME/lib
export DYLD_LIBRARY_PATH=$DYLD_LIBRARY_PATH:/Users/mavicmaverick/go/src/github.com/ibmdb/go_ibm_db/installer/clidriver/lib

# build
go build -o ./bin ./cmd/wmindex
./bin/wmindex
