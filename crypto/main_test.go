package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testSigner struct {
	algorithm Algorithm
}

func (s testSigner) GenerateKeyPair() ([]byte, []byte, error) {
	return []byte("priv"), []byte("pub"), nil
}

func (s testSigner) Sign(privateKey, data []byte) ([]byte, error) {
	return append(privateKey, data...), nil
}

func (s testSigner) Algorithm() Algorithm {
	return s.algorithm
}

// Указанные методы подписи работают, а на неизвестный выдаёт ошибку
func Test_NewSignerFactorySupportsOnlyRSAAndECCByDefault(t *testing.T) {
	factory := NewSignerFactory()

	assert.Len(t, factory.signers, 2)

	_, err := factory.GetSigner(AlgorithmRSA)
	require.NoError(t, err)

	_, err = factory.GetSigner(AlgorithmECC)
	require.NoError(t, err)

	_, err = factory.GetSigner("DSA")
	assert.Error(t, err)
}

// Алгоритмы подписи свободно добавляются
func Test_SignersMapCanBeExtendedWithCustomAlgorithm(t *testing.T) {
	const customAlgorithm Algorithm = "TEST_ALGO"

	factory := &SignersMap{
		signers: map[Algorithm]Signer{
			AlgorithmRSA: &RSASigner{},
			AlgorithmECC: &ECCSigner{},
			customAlgorithm: testSigner{
				algorithm: customAlgorithm,
			},
		},
	}

	signer, err := factory.GetSigner(customAlgorithm)
	require.NoError(t, err)

	assert.Equal(t, customAlgorithm, signer.Algorithm())
}
