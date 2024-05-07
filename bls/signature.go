package bls

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	flowCrypto "github.com/onflow/crypto"
	"github.com/onflow/crypto/hash"
)

// The signingAlgorithm for BLS keys is BLSBLS12381 which is BLS on the BLS 12-381 curve.
// This is the only supported BLS signing algorithm in the flowCrypto package.
// BLS is used such that we can aggregate signatures into one signature.
const signingAlgorithm = flowCrypto.BLSBLS12381

// The hashingAlgorithm for BLS keys is the following. This algorithm is used to hash input data onto the
// BLS 12-381 curve for generating signatures. The returned instance is a Hasher and can be used to
// generate BLS signatures with the Sign() method. This is the only supported BLS Hasher in the flowCrypto
// package. The input domainTag is a separation tag that defines the protocol and its subdomain. Such tag
// should be of the format: <protocol>-V<xx>-CS<yy>-with- where <protocol> is the name of the protocol,
// <xx> the protocol version number, and <yy> the index of the ciphersuite in the protocol.
var hashingAlgorithm = flowCrypto.NewExpandMsgXOFKMAC128("deso-V1-CS01-with-")

// AggregateSignatures takes in an input slice of bls.Signatures and aggregates them
// into a single bls.Signature. This signature aggregation supports signatures on a
// single or different payloads.
func AggregateSignatures(signatures []*Signature) (*Signature, error) {
	var flowSignatures []flowCrypto.Signature
	for _, signature := range signatures {
		flowSignatures = append(flowSignatures, signature.flowSignature)
	}
	aggregateFlowSignature, err := flowCrypto.AggregateBLSSignatures(flowSignatures)
	if err != nil {
		return nil, err
	}
	return &Signature{flowSignature: aggregateFlowSignature}, nil
}

// VerifyAggregateSignatureSinglePayload takes in a slice of bls.PublicKeys, a bls.Signature, and a single payload and returns
// true if every bls.PublicKey in the slice signed the payload. The input bls.Signature is the aggregate
// signature of each of their respective bls.Signatures for that payload.
func VerifyAggregateSignatureSinglePayload(publicKeys []*PublicKey, signature *Signature, payloadBytes []byte) (bool, error) {
	flowPublicKeys, err := extractFlowPublicKeys(publicKeys)
	if err != nil {
		return false, err
	}
	return flowCrypto.VerifyBLSSignatureOneMessage(flowPublicKeys, signature.flowSignature, payloadBytes, hashingAlgorithm)
}

// VerifyAggregateSignatureMultiplePayloads takes in a slice of bls.PublicKeys, a bls.Signature, and a slice of payloads.
// It returns true if each bls.PublicKey at index i has signed its respective payload at index i in the payloads slice.
// The input bls.Signature is the aggregate signature of each public key's partial bls.Signatures for its respective payload.
func VerifyAggregateSignatureMultiplePayloads(publicKeys []*PublicKey, signature *Signature, payloadsBytes [][]byte) (bool, error) {
	if len(publicKeys) != len(payloadsBytes) {
		return false, fmt.Errorf("number of public keys %d does not equal number of payloads %d", len(publicKeys), len(payloadsBytes))
	}

	flowPublicKeys, err := extractFlowPublicKeys(publicKeys)
	if err != nil {
		return false, err
	}

	var hashingAlgorithms []hash.Hasher
	for ii := 0; ii < len(publicKeys); ii++ {
		hashingAlgorithms = append(hashingAlgorithms, hashingAlgorithm)
	}

	return flowCrypto.VerifyBLSSignatureManyMessages(flowPublicKeys, signature.flowSignature, payloadsBytes, hashingAlgorithms)
}

//
// TYPES: PrivateKey
//

type PrivateKey struct {
	flowPrivateKey flowCrypto.PrivateKey
}

func NewPrivateKey() (*PrivateKey, error) {
	randomBytes := make([]byte, 64)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}
	flowPrivateKey, err := flowCrypto.GeneratePrivateKey(signingAlgorithm, randomBytes)
	if err != nil {
		return nil, err
	}
	return &PrivateKey{flowPrivateKey: flowPrivateKey}, nil
}

func (privateKey *PrivateKey) Sign(payloadBytes []byte) (*Signature, error) {
	if privateKey == nil || privateKey.flowPrivateKey == nil {
		return nil, errors.New("PrivateKey is nil")
	}
	flowSignature, err := privateKey.flowPrivateKey.Sign(payloadBytes, hashingAlgorithm)
	if err != nil {
		return nil, err
	}
	return &Signature{flowSignature: flowSignature}, nil
}

func (privateKey *PrivateKey) PublicKey() *PublicKey {
	if privateKey == nil || privateKey.flowPrivateKey == nil {
		return nil
	}
	publicKey := privateKey.flowPrivateKey.PublicKey()
	return &PublicKey{flowPublicKey: publicKey, flowPublicKeyBytes: publicKey.Encode()}
}

func (privateKey *PrivateKey) ToString() string {
	if privateKey == nil || privateKey.flowPrivateKey == nil {
		return ""
	}
	return privateKey.flowPrivateKey.String()
}

func (privateKey *PrivateKey) FromSeed(seed []byte) (*PrivateKey, error) {
	var err error

	if privateKey == nil {
		return nil, nil
	}

	// Generate a new private key from the seed.
	privateKey.flowPrivateKey, err = flowCrypto.GeneratePrivateKey(signingAlgorithm, seed)
	return privateKey, err
}

func (privateKey *PrivateKey) FromString(privateKeyString string) (*PrivateKey, error) {
	if privateKey == nil || privateKeyString == "" {
		return nil, nil
	}
	// Chop off leading 0x, if exists. Otherwise, does nothing.
	privateKeyStringCopy, _ := strings.CutPrefix(privateKeyString, "0x")
	// Convert from hex string to byte slice.
	privateKeyBytes, err := hex.DecodeString(privateKeyStringCopy)
	if err != nil {
		return nil, err
	}
	// Convert from byte slice to bls.PrivateKey.
	privateKey.flowPrivateKey, err = flowCrypto.DecodePrivateKey(signingAlgorithm, privateKeyBytes)
	return privateKey, err
}

func (privateKey *PrivateKey) MarshalJSON() ([]byte, error) {
	// This is called automatically by the JSON library when converting a
	// bls.PrivateKey to JSON. This is currently not used, since the client
	// never shares their bls.PrivateKey over the network, but is included
	// here are as a nicety utility for completeness.
	return json.Marshal(privateKey.ToString())
}

func (privateKey *PrivateKey) UnmarshalJSON(data []byte) error {
	// This is called automatically by the JSON library when converting a
	// bls.PrivateKey from JSON. This is currently not used, since the client
	// never shares their bls.PrivateKey over the network, but is included
	// here are as a nicety utility for completeness.
	privateKeyString := ""
	err := json.Unmarshal(data, &privateKeyString)
	if err != nil {
		return err
	}
	_, err = privateKey.FromString(privateKeyString)
	return err
}

func (privateKey *PrivateKey) Eq(other *PrivateKey) bool {
	if privateKey == nil || privateKey.flowPrivateKey == nil || other == nil {
		return false
	}
	return privateKey.flowPrivateKey.Equals(other.flowPrivateKey)
}

//
// TYPES: PublicKey
//

type PublicKey struct {
	flowPublicKeyBytes []byte
	flowPublicKey      flowCrypto.PublicKey
}

func (publicKey *PublicKey) loadFlowPublicKey() error {
	if publicKey != nil &&
		publicKey.flowPublicKey == nil &&
		len(publicKey.flowPublicKeyBytes) > 0 {
		var err error
		publicKey.flowPublicKey, err = flowCrypto.DecodePublicKey(signingAlgorithm, publicKey.flowPublicKeyBytes)
		return err
	}
	return nil
}

func (publicKey *PublicKey) Verify(signature *Signature, input []byte) (bool, error) {
	if publicKey == nil || len(publicKey.flowPublicKeyBytes) == 0 {
		return false, errors.New("bls.PublicKey is nil")
	}
	if publicKey.loadFlowPublicKey() != nil {
		return false, errors.New("failed to load flowPublicKey")
	}
	return publicKey.flowPublicKey.Verify(signature.flowSignature, input, hashingAlgorithm)
}

func (publicKey *PublicKey) ToBytes() []byte {
	return publicKey.flowPublicKeyBytes
}

func (publicKey *PublicKey) FromBytes(publicKeyBytes []byte) (*PublicKey, error) {
	if publicKey == nil || len(publicKeyBytes) == 0 {
		return nil, nil
	}
	publicKey.flowPublicKeyBytes = publicKeyBytes
	return publicKey, nil
}

func (publicKey *PublicKey) ToString() string {
	if publicKey == nil || len(publicKey.flowPublicKeyBytes) == 0 {
		return ""
	}
	return "0x" + hex.EncodeToString(publicKey.flowPublicKeyBytes)
}

func (publicKey *PublicKey) FromString(publicKeyString string) (*PublicKey, error) {
	if publicKey == nil || publicKeyString == "" {
		return nil, nil
	}
	// Chop off leading 0x, if exists. Otherwise, does nothing.
	publicKeyStringCopy, _ := strings.CutPrefix(publicKeyString, "0x")
	// Convert from hex string to byte slice.
	publicKeyBytes, err := hex.DecodeString(publicKeyStringCopy)
	if err != nil {
		return nil, err
	}
	publicKey.flowPublicKeyBytes = publicKeyBytes
	return publicKey, err
}

func (publicKey *PublicKey) ToAbbreviatedString() string {
	str := publicKey.ToString()
	if len(str) <= 8 {
		return str
	}
	return str[:8] + "..." + str[len(str)-8:]
}

func (publicKey *PublicKey) MarshalJSON() ([]byte, error) {
	// This is called automatically by the JSON library when converting a
	// bls.PublicKey to JSON. This is useful when passing a bls.PublicKey
	// back and forth from the backend to the frontend as JSON.
	return json.Marshal(publicKey.ToString())
}

func (publicKey *PublicKey) UnmarshalJSON(data []byte) error {
	// This is called automatically by the JSON library when converting a
	// bls.PublicKey from JSON. This is useful when passing a bls.PublicKey
	// back and forth from the frontend to the backend as JSON.
	publicKeyString := ""
	err := json.Unmarshal(data, &publicKeyString)
	if err != nil {
		return err
	}
	_, err = publicKey.FromString(publicKeyString)
	return err
}

func (publicKey *PublicKey) Eq(other *PublicKey) bool {
	if publicKey == nil || publicKey.flowPublicKeyBytes == nil || other == nil {
		return false
	}
	return bytes.Equal(publicKey.flowPublicKeyBytes, other.flowPublicKeyBytes)
}

func (publicKey *PublicKey) Copy() *PublicKey {
	if publicKey == nil {
		return nil
	}
	return &PublicKey{
		flowPublicKeyBytes: publicKey.flowPublicKeyBytes,
		flowPublicKey:      publicKey.flowPublicKey,
	}
}

func (publicKey *PublicKey) IsEmpty() bool {
	return publicKey == nil || publicKey.flowPublicKeyBytes == nil
}

type SerializedPublicKey string

func (publicKey *PublicKey) Serialize() SerializedPublicKey {
	return SerializedPublicKey(publicKey.ToString())
}

func (serializedPublicKey SerializedPublicKey) Deserialize() (*PublicKey, error) {
	return new(PublicKey).FromString(string(serializedPublicKey))
}

//
// TYPES: Signature
//

type Signature struct {
	flowSignature flowCrypto.Signature
}

func (signature *Signature) ToBytes() []byte {
	var signatureBytes []byte
	if signature != nil && signature.flowSignature != nil {
		signatureBytes = signature.flowSignature.Bytes()
	}
	return signatureBytes
}

func (signature *Signature) FromBytes(signatureBytes []byte) (*Signature, error) {
	if signature == nil || len(signatureBytes) == 0 {
		return nil, nil
	}
	signature.flowSignature = signatureBytes
	return signature, nil
}

func (signature *Signature) ToString() string {
	if signature == nil || signature.flowSignature == nil {
		return ""
	}
	return signature.flowSignature.String()
}

func (signature *Signature) FromString(signatureString string) (*Signature, error) {
	if signature == nil || signatureString == "" {
		return nil, nil
	}
	// Chop off leading 0x, if exists. Otherwise, does nothing.
	signatureStringCopy, _ := strings.CutPrefix(signatureString, "0x")
	// Convert from hex string to byte slice.
	signatureBytes, err := hex.DecodeString(signatureStringCopy)
	if err != nil {
		return nil, err
	}
	// Convert from byte slice to bls.Signature.
	signature.flowSignature = signatureBytes
	return signature, nil
}

func (signature *Signature) ToAbbreviatedString() string {
	str := signature.ToString()
	if len(str) <= 8 {
		return str
	}
	return str[:8] + "..." + str[len(str)-8:]
}

func (signature *Signature) MarshalJSON() ([]byte, error) {
	// This is called automatically by the JSON library when converting a
	// bls.Signature to JSON. This is useful when passing a bls.Signature
	// back and forth from the backend to the frontend as JSON.
	return json.Marshal(signature.ToString())
}

func (signature *Signature) UnmarshalJSON(data []byte) error {
	// This is called automatically by the JSON library when converting a
	// bls.Signature from JSON. This is useful when passing a bls.Signature
	// back and forth from the frontend to the backend as JSON.
	signatureString := ""
	err := json.Unmarshal(data, &signatureString)
	if err != nil {
		return err
	}
	_, err = signature.FromString(signatureString)
	return err
}

func (signature *Signature) Eq(other *Signature) bool {
	if signature == nil || signature.flowSignature == nil || other == nil {
		return false
	}
	return bytes.Equal(signature.ToBytes(), other.ToBytes())
}

func (signature *Signature) Copy() *Signature {
	if signature == nil {
		return nil
	}
	if signature.flowSignature == nil {
		return &Signature{}
	}
	return &Signature{
		flowSignature: append([]byte{}, signature.flowSignature.Bytes()...),
	}
}

func (signature *Signature) IsEmpty() bool {
	return signature == nil || signature.flowSignature == nil
}

func extractFlowPublicKeys(publicKeys []*PublicKey) ([]flowCrypto.PublicKey, error) {
	flowPublicKeys := make([]flowCrypto.PublicKey, len(publicKeys))
	for i, publicKey := range publicKeys {
		if err := publicKey.loadFlowPublicKey(); err != nil {
			return nil, err
		}
		flowPublicKeys[i] = publicKey.flowPublicKey
	}
	return flowPublicKeys, nil
}
