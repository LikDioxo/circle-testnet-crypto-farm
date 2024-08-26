GO_UTIL := go
BUILD_OUT_DIR := bin
BUILD_OUT_PATH := $(shell pwd)/${BUILD_OUT_DIR}/circle-farm
MAIN_SCRIPT := $(shell pwd)/cmd/app/main.go

run:
	${GO_UTIL} run ${MAIN_SCR	IPT} -n 1 -blockchain MATIC-AMOY -dest 0xddDcBe5eb37Acc00f5Fb059C9e850c28FcF5333D

build:
	${GO_UTIL} build -o ${BUILD_OUT_PATH} ${MAIN_SCRIPT}

run-amoy: ${BUILD_OUT_PATH}
	${BUILD_OUT_PATH} -n 10 -blockchain MATIC-AMOY -dest 0x753b3D36Cb059AEA57CD6eA0dC2772AfD38c1b72
