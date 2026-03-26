BINARYNAME=ynab-metrics
BINARYPATH=target
BINARY = ${BINARYPATH}/${BINARYNAME}
CFGFILE=config.json

build:
	go build -o ${BINARY} -v

arm:
	GOOS=linux GOARCH=arm64 go build -o ${BINARY}.arm64 -v

run:
	./${BINARY} --config ${CFGFILE}
