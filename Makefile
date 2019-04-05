TMP=./tmp
PROJECT_ID=dev-status-api
ZONE=us-west1-a

VERSION=$(shell date +%s)

SERVICE_DEPLOY_LOCATION=$(shell echo "gs://${PROJECT_ID}/service-${VERSION}.tar")

all: build-service

build-service:
	GOOS=linux GOARCH=amd64 go build -v -o ${TMP}/service ./service
	tar -c -f ${TMP}/service-bundle.tar -C ${TMP} service

deploy-service: build-service
	test "$(SERVICE_DEPLOY_LOCATION)"

	gsutil cp ${TMP}/service-bundle.tar ${SERVICE_DEPLOY_LOCATION}
	echo "uploaded to ${SERVICE_DEPLOY_LOCATION}"

	gcloud compute instances create ${PROJECT_ID} \
		--image-family=debian-9 \
		--image-project=debian-cloud \
		--machine-type=g1-small \
		--scopes datastore,cloud-platform \
		--metadata app-location=${SERVICE_DEPLOY_LOCATION} \
		--metadata-from-file startup-script=deploy/startup-service.sh \
		--zone ${ZONE}

teardown:
	gcloud config set compute/zone ${ZONE}
	gcloud compute instances delete ${PROJECT_ID}