package kubeloader

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetClusters(t *testing.T) {

	runGetStdoutReplace = func() (*bytes.Buffer, error) {
		return bytes.NewBufferString("  * banana1\nbanana2\napple1\nanotherbanana\nnotAFruit\n"), nil
	}
	clusters, _ := GetClusters("banana")
	require.Len(t, clusters, 3)
	require.Equal(t, "banana1", clusters[0])
	require.Equal(t, "banana2", clusters[1])
	require.Equal(t, "anotherbanana", clusters[2])

}
