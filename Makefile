BINARY_NAME=helm-api
INSTALL_DIR=/opt/helm-api/

build:
	go get
	go build -o ${BINARY_NAME} 
install:
	install -D ${BINARY_NAME} ${INSTALL_DIR}/${BINARY_NAME}
	cp conf /opt/helm-api/ -R
