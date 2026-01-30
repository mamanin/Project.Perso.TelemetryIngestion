package credential

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

// AzureCredentials provides Azure identity credentials.
type AzureCredentials azcore.TokenCredential

// NewAzureDefault retrieves Azure identity credentials using DefaultAzureCredential.
func NewAzureDefault() (AzureCredentials, error) {
	credentials, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create azure credentials: %w", err)
	}

	return credentials, nil
}
