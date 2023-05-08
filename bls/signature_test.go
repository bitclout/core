//go:build relic

package bls

import (
	"bytes"
	"crypto/rand"
	flowCrypto "github.com/onflow/flow-go/crypto"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestVerifyingBLSSignatures(t *testing.T) {
	// Generate two BLS public/private key pairs.
	blsPrivateKey1 := _generateRandomBLSPrivateKey(t)
	blsPublicKey1 := blsPrivateKey1.PublicKey()

	blsPrivateKey2 := _generateRandomBLSPrivateKey(t)
	blsPublicKey2 := blsPrivateKey2.PublicKey()

	// Test bls.PrivateKey.Sign() and bls.PublicKey.Verify().
	// 1. PrivateKey1 signs a random payload.
	randomPayload1 := _generateRandomBytes(t, 256)
	blsSignature1, err := blsPrivateKey1.Sign(randomPayload1)
	require.NoError(t, err)
	// 2. Verify bls.PublicKey1 is the signer.
	isVerified, err := blsPublicKey1.Verify(blsSignature1, randomPayload1)
	require.NoError(t, err)
	require.True(t, isVerified)
	// 3. Verify bls.PublicKey2 is not the signer.
	isVerified, err = blsPublicKey2.Verify(blsSignature1, randomPayload1)
	require.NoError(t, err)
	require.False(t, isVerified)

	// 4. PrivateKey2 signs a different random payload.
	randomPayload2 := _generateRandomBytes(t, 256)
	blsSignature2, err := blsPrivateKey2.Sign(randomPayload2)
	require.NoError(t, err)
	// 5. Verify bls.PublicKey1 is not the signer.
	isVerified, err = blsPublicKey1.Verify(blsSignature2, randomPayload2)
	require.NoError(t, err)
	require.False(t, isVerified)
	// 6. Verify bls.PublicKey2 is the signer.
	isVerified, err = blsPublicKey2.Verify(blsSignature2, randomPayload2)
	require.NoError(t, err)
	require.True(t, isVerified)

	// Test AggregateSignatures() and VerifyAggregateSignature().
	// 1. PrivateKey1 signs a random payload.
	randomPayload3 := _generateRandomBytes(t, 256)
	blsSignature1, err = blsPrivateKey1.Sign(randomPayload3)
	require.NoError(t, err)
	// 2. PrivateKey2 signs the same random payload.
	blsSignature2, err = blsPrivateKey2.Sign(randomPayload3)
	require.NoError(t, err)
	// 3. Aggregate their signatures.
	aggregateSignature, err := AggregateSignatures([]*Signature{blsSignature1, blsSignature2})
	require.NoError(t, err)
	// 4. Verify the AggregateSignature.
	isVerified, err = VerifyAggregateSignature(
		[]*PublicKey{blsPublicKey1, blsPublicKey2}, aggregateSignature, randomPayload3,
	)
	require.NoError(t, err)
	require.True(t, isVerified)
	// 5. Verify PrivateKey1's signature doesn't work on its own.
	isVerified, err = VerifyAggregateSignature([]*PublicKey{blsPublicKey1}, aggregateSignature, randomPayload3)
	require.NoError(t, err)
	require.False(t, isVerified)
	// 6. Verify PrivateKey2's signature doesn't work on its own.
	isVerified, err = VerifyAggregateSignature([]*PublicKey{blsPublicKey2}, aggregateSignature, randomPayload3)
	require.NoError(t, err)
	require.False(t, isVerified)
	// 7. Verify the AggregateSignature doesn't work on a different payload.
	isVerified, err = VerifyAggregateSignature(
		[]*PublicKey{blsPublicKey1, blsPublicKey2}, aggregateSignature, randomPayload1,
	)
	require.NoError(t, err)
	require.False(t, isVerified)

	// Test bls.PrivateKey.Eq().
	require.True(t, blsPrivateKey1.Eq(blsPrivateKey1))
	require.True(t, blsPrivateKey2.Eq(blsPrivateKey2))
	require.False(t, blsPrivateKey1.Eq(blsPrivateKey2))

	// Test bls.PrivateKey.ToString() and bls.PrivateKey.FromString().
	blsPrivateKeyString := blsPrivateKey1.ToString()
	copyBLSPrivateKey1, err := (&PrivateKey{}).FromString(blsPrivateKeyString)
	require.NoError(t, err)
	require.True(t, blsPrivateKey1.Eq(copyBLSPrivateKey1))

	// Test bls.PublicKey.Eq().
	require.True(t, blsPublicKey1.Eq(blsPublicKey1))
	require.True(t, blsPublicKey2.Eq(blsPublicKey2))
	require.False(t, blsPublicKey1.Eq(blsPublicKey2))

	// Test bls.PublicKey.ToBytes() and bls.PublicKey.FromBytes().
	blsPublicKeyBytes := blsPublicKey1.ToBytes()
	copyBLSPublicKey1, err := (&PublicKey{}).FromBytes(blsPublicKeyBytes)
	require.NoError(t, err)
	require.True(t, blsPublicKey1.Eq(copyBLSPublicKey1))

	// Test bls.PublicKey.ToString() and bls.PublicKey.FromString().
	blsPublicKeyString := blsPublicKey1.ToString()
	copyBLSPublicKey1, err = (&PublicKey{}).FromString(blsPublicKeyString)
	require.NoError(t, err)
	require.True(t, blsPublicKey1.Eq(copyBLSPublicKey1))

	// Test bls.Signature.Eq().
	require.True(t, blsSignature1.Eq(blsSignature1))
	require.True(t, blsSignature2.Eq(blsSignature2))
	require.False(t, blsSignature1.Eq(blsSignature2))

	// Test bls.Signature.ToBytes() and bls.Signature.FromBytes().
	blsSignatureBytes := blsSignature1.ToBytes()
	copyBLSSignature, err := (&Signature{}).FromBytes(blsSignatureBytes)
	require.NoError(t, err)
	require.True(t, blsSignature1.Eq(copyBLSSignature))

	// Test bls.Signature.ToString() and bls.Signature.FromString().
	blsSignatureString := blsSignature1.ToString()
	copyBLSSignature, err = (&Signature{}).FromString(blsSignatureString)
	require.NoError(t, err)
	require.True(t, blsSignature1.Eq(copyBLSSignature))

	// Test bls.PublicKey.Copy().
	blsPublicKey1Copy := blsPublicKey1.Copy()
	require.True(t, blsPublicKey1.Eq(blsPublicKey1Copy))

	// Test bls.Signature.Copy().
	blsSignature1Copy := blsSignature1.Copy()
	require.True(t, blsSignature1.Eq(blsSignature1Copy))

	// Test nil bls.PrivateKey edge cases.
	// Sign()
	_, err = (&PrivateKey{}).Sign(randomPayload1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "bls.PrivateKey is nil")
	// PublicKey()
	require.Nil(t, (&PrivateKey{}).PublicKey())
	// ToString()
	require.Equal(t, (&PrivateKey{}).ToString(), "")
	// FromString()
	_, err = (&PrivateKey{}).FromString("")
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty bls.PrivateKey string provided")
	// Eq()
	require.False(t, (&PrivateKey{}).Eq(nil))
	require.False(t, (&PrivateKey{}).Eq(&PrivateKey{}))
	require.False(t, (&PrivateKey{}).Eq(_generateRandomBLSPrivateKey(t)))
	require.False(t, _generateRandomBLSPrivateKey(t).Eq(nil))
	require.False(t, _generateRandomBLSPrivateKey(t).Eq(&PrivateKey{}))
	require.False(t, _generateRandomBLSPrivateKey(t).Eq(_generateRandomBLSPrivateKey(t)))

	// Test nil bls.PublicKey edge cases.
	// Verify()
	_, err = (&PublicKey{}).Verify(blsSignature1, randomPayload1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "bls.PublicKey is nil")
	// ToBytes()
	require.True(t, bytes.Equal((&PublicKey{}).ToBytes(), []byte{}))
	// FromBytes()
	_, err = (&PublicKey{}).FromBytes(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty bls.PublicKey bytes provided")
	_, err = (&PublicKey{}).FromBytes([]byte{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty bls.PublicKey bytes provided")
	// ToString()
	require.Equal(t, (&PublicKey{}).ToString(), "")
	// FromString()
	_, err = (&PublicKey{}).FromString("")
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty bls.PublicKey string provided")
	// Eq()
	require.False(t, (&PublicKey{}).Eq(nil))
	require.False(t, (&PublicKey{}).Eq(&PublicKey{}))
	require.False(t, (&PublicKey{}).Eq(_generateRandomBLSPrivateKey(t).PublicKey()))
	require.False(t, _generateRandomBLSPrivateKey(t).PublicKey().Eq(nil))
	require.False(t, _generateRandomBLSPrivateKey(t).PublicKey().Eq((&PrivateKey{}).PublicKey()))
	require.False(t, _generateRandomBLSPrivateKey(t).PublicKey().Eq(_generateRandomBLSPrivateKey(t).PublicKey()))
	// Copy()
	require.Nil(t, (&PublicKey{}).Copy().flowPublicKey)

	// Test nil bls.Signature edge cases.
	// ToBytes()
	require.True(t, bytes.Equal((&Signature{}).ToBytes(), []byte{}))
	// FromBytes()
	_, err = (&Signature{}).FromBytes(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty bls.Signature bytes provided")
	_, err = (&Signature{}).FromBytes([]byte{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty bls.Signature bytes provided")
	// ToString()
	require.Equal(t, (&Signature{}).ToString(), "")
	// FromString()
	_, err = (&Signature{}).FromString("")
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty bls.Signature string provided")
	// Eq()
	require.False(t, (&Signature{}).Eq(nil))
	require.False(t, (&Signature{}).Eq(&Signature{}))
	require.False(t, (&Signature{}).Eq(blsSignature1))
	require.False(t, blsSignature1.Eq(nil))
	require.False(t, blsSignature1.Eq(&Signature{}))
	// Copy()
	require.Nil(t, (&Signature{}).Copy().flowSignature)
}

func _generateRandomBLSPrivateKey(t *testing.T) *PrivateKey {
	flowPrivateKey, err := flowCrypto.GeneratePrivateKey(SigningAlgorithm, _generateRandomBytes(t, 64))
	require.NoError(t, err)
	return &PrivateKey{flowPrivateKey: flowPrivateKey}
}

func _generateRandomBytes(t *testing.T, numBytes int) []byte {
	randomBytes := make([]byte, 64)
	_, err := rand.Read(randomBytes)
	require.NoError(t, err)
	return randomBytes
}
