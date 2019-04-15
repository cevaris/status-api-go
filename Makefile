TMP=./tmp
ZONE=us-west1-a

VERSION=$(shell date +%s)

PROJECT_ID=status-api-dev
APPNAME_RUNNER=api-status-runner
BUNDLE_LOCATION_RUNNER=$(shell echo "gs://${PROJECT_ID}/${APPNAME_RUNNER}/${VERSION}.tar")

all: build-runner

run-runner:
	PROJECT_ID=${PROJECT_ID} go run runner/main.go

build-runner:
	GOOS=linux GOARCH=amd64 go build -v -o ${TMP}/runner ./runner
	tar -c -f ${TMP}/runner-bundle.tar -C ${TMP} runner

deploy-runner: build-runner
	test "$(BUNDLE_LOCATION_RUNNER)"

	gsutil cp ${TMP}/runner-bundle.tar ${BUNDLE_LOCATION_RUNNER}
	echo "uploaded to ${BUNDLE_LOCATION_RUNNER}"

	gcloud compute instances create ${APPNAME_RUNNER} \
		--image-family=debian-9 \
		--image-project=debian-cloud \
		--machine-type=g1-small \
		--scopes datastore,cloud-platform \
		--metadata app-location=${BUNDLE_LOCATION_RUNNER} \
		--metadata-from-file startup-script=deploy/startup-runner.sh \
		--preemptible \
		--zone ${ZONE}

ssh-runner:
	gcloud compute --project "${PROJECT_ID}" ssh --zone "${ZONE}" "${APPNAME_RUNNER}"

teardown:
	gcloud config set compute/zone ${ZONE}
	gcloud compute instances delete ${APPNAME_RUNNER}
	gcloud beta emulators datastore env-unset


encrypt-secrets:
	gcloud kms encrypt \
	  --location=global  \
	  --keyring=status-api-key-ring \
	  --key=status-api-secrets \
	  --plaintext-file=/git/secrets/status-api/secrets.json \
	  --ciphertext-file=secrets/data.enc