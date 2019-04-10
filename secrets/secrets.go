package secrets

import (
	cloudkms "cloud.google.com/go/kms/apiv1"
	"context"
	"encoding/json"
	"github.com/cevaris/status/logging"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
	"io/ioutil"
	"os"
	"path/filepath"
)

var ReadOnlyApiKeys *ApiKeys

func init() {
	var err error
	ReadOnlyApiKeys, err = NewInstance(context.Background())
	if err != nil {
		panic("could not load api keys " + err.Error())
	}
}

var logger = logging.Logger()
var keyName = "projects/status-api-dev/locations/global/keyRings/status-api-key-ring/cryptoKeys/status-api-secrets"

type ApiKeys struct {
	AwsAccessKeyID     string `json:"aws_access_key_id"`
	AwsSecretAccessKey string `json:"aws_secret_access_key"`
}

func NewInstance(ctx context.Context) (*ApiKeys, error) {
	encBytes, err := readBytes()
	if err != nil {
		logger.Error(ctx, "failed to read encrypted bytes", err)
		return nil, err
	}
	plainBytes, err := decryptSymmetric(keyName, encBytes)
	if err != nil {
		logger.Error(ctx, "failed to decrypt bytes", err)
		return nil, err
	}

	var apiKeys ApiKeys
	if err := json.Unmarshal(plainBytes, &apiKeys); err != nil {
		logger.Error(ctx, "failed to unmarshal plain bytes", err, "\n", string(plainBytes))
		return nil, err
	}

	return &apiKeys, nil
}

func decryptSymmetric(keyName string, ciphertext []byte) ([]byte, error) {
	ctx := context.Background()
	client, err := cloudkms.NewKeyManagementClient(ctx)
	if err != nil {
		return nil, err
	}

	// Build the request.
	req := &kmspb.DecryptRequest{
		Name:       keyName,
		Ciphertext: ciphertext,
	}
	// Call the API.
	resp, err := client.Decrypt(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Plaintext, nil
}

func readBytes() ([]byte, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return ioutil.ReadFile(filepath.Join(pwd, "secrets/data.enc"))
}
