export CGO_ENABLED=1

export WDIR=${PWD}



all: linux

linux:
	GOOS=linux CGO_CFLAGS="-I${WDIR}/include" CGO_LDFLAGS="-L${WDIR}/lib/Linux -Wl,-rpath=${WDIR}/lib/Linux -lhcnetsdk" go build -x -ldflags "-s -w" -o build/Linux/main main.go cfuncs.go
	cp -r lib/Linux/* build/Linux/

clean:
	rm -r build/
