build() {
	rm -rf dist && export GOOS="$1" GOARCH="$2" && printf '==========    OS: %s    ARCH: %s\n' "$GOOS" "$GOARCH" && make -j12 && cd dist && tar cvf ../cca-0.1.12-$GOOS-$GOARCH.tar cca docs iadocs && cd ..
}

build aix ppc64
build android 386
build android amd64
build android arm
build android arm64
build darwin amd64
build darwin arm64
build dragonfly amd64
build freebsd 386
build freebsd amd64
build freebsd arm
build illumos amd64
build ios arm64
build js wasm
build linux 386
build linux amd64
build linux arm
build linux arm64
build linux loong64
build linux mips
build linux mipsle
build linux mips64
build linux mips64le
build linux ppc64
build linux ppc64le
build linux riscv64
build linux s390x
build netbsd 386
build netbsd amd64
build netbsd arm
build openbsd 386
build openbsd amd64
build openbsd arm
build openbsd arm64
build plan9 386
build plan9 amd64
build plan9 arm
build solaris amd64
build wasip1 wasm
build windows 386
build windows amd64
build windows arm
build windows arm64
