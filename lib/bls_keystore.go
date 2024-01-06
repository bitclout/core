package lib

import (
	"github.com/deso-protocol/core/bls"
	"github.com/deso-protocol/core/consensus"
	"github.com/pkg/errors"
)

// BLSSigner is a wrapper for the bls.PrivateKey type, which abstracts away the private key
// and only exposes protected methods for signing a select set of message types needed for
// Proof of Stake. It allows signing for:
// - PoS Validator Votes Messages
// - PoS Validator Timeout Messages
// - PoS Block Proposals
// - PoS Validator Connection Handshakes
// - PoS Random Seed Signature
//
// TODO: We will likely need to associate individual op-codes for each message type that can be signed,
// so that there is no risk of signature collisions between different message types. Ex: the payload
// signed per message type must be made up of the following tuples:
// - Validator Vote:            (0x01, view uint64, blockHash consensus.BlockHash)
// - Validator Timeout:         (0x02, view uint64, highQCView uint64)
// - PoS Block Proposal:        (0x03, view uint64, blockHash consensus.BlockHash)
// - PoS Validator Handshake:   (0x04, peer's random nonce, our node's random nonce)
// - PoS Random Seed Signature: (previous block's random seed hash)

type BLSSignatureOpCode byte

const (
	BLSSignatureOpCodeValidatorVote         BLSSignatureOpCode = 0
	BLSSignatureOpCodeValidatorTimeout      BLSSignatureOpCode = 1
	BLSSignatureOpCodePoSBlockProposal      BLSSignatureOpCode = 2
	BLSSignatureOpCodePoSValidatorHandshake BLSSignatureOpCode = 3
)

func (opCode BLSSignatureOpCode) Bytes() []byte {
	return []byte{byte(opCode)}
}

//////////////////////////////////////////////////////////
// BLSKeystore
//////////////////////////////////////////////////////////

type BLSKeystore struct {
	signer *BLSSigner
}

func NewBLSKeystore(seed string) (*BLSKeystore, error) {
	privateKey, err := bls.NewPrivateKey()
	if err != nil {
		return nil, errors.Wrapf(err, "NewBLSKeystore: Problem generating private key from seed: %s", seed)
	}
	if _, err = privateKey.FromString(seed); err != nil {
		return nil, errors.Wrapf(err, "NewBLSKeystore: Problem retrieving private key from seed: %s", seed)
	}

	signer, err := NewBLSSigner(privateKey)
	if err != nil {
		return nil, err
	}
	return &BLSKeystore{signer: signer}, nil
}

func (keystore *BLSKeystore) GetSigner() *BLSSigner {
	return keystore.signer
}

//////////////////////////////////////////////////////////
// BLSSigner
//////////////////////////////////////////////////////////

type BLSSigner struct {
	privateKey *bls.PrivateKey
}

func NewBLSSigner(privateKey *bls.PrivateKey) (*BLSSigner, error) {
	if privateKey == nil {
		return nil, errors.New("NewBLSSigner: privateKey cannot be nil")
	}
	return &BLSSigner{privateKey: privateKey}, nil
}

func (signer *BLSSigner) sign(opCode BLSSignatureOpCode, payload []byte) (*bls.Signature, error) {
	newPayload := append(opCode.Bytes(), payload...)
	return signer.privateKey.Sign(newPayload)
}

func (signer *BLSSigner) GetPublicKey() *bls.PublicKey {
	return signer.privateKey.PublicKey()
}

func (signer *BLSSigner) SignBlockProposal(view uint64, blockHash consensus.BlockHash) (*bls.Signature, error) {
	// A block proposer's signature on a block is just its partial vote signature. This allows us to aggregate
	// signatures from the proposer and validators into a single aggregated signature to build a QC.
	return signer.SignValidatorVote(view, blockHash)
}

func (signer *BLSSigner) SignValidatorVote(view uint64, blockHash consensus.BlockHash) (*bls.Signature, error) {
	payload := consensus.GetVoteSignaturePayload(view, blockHash)
	return signer.sign(BLSSignatureOpCodeValidatorVote, payload[:])
}

func (signer *BLSSigner) SignValidatorTimeout(view uint64, highQCView uint64) (*bls.Signature, error) {
	payload := consensus.GetTimeoutSignaturePayload(view, highQCView)
	return signer.sign(BLSSignatureOpCodeValidatorTimeout, payload[:])
}

func (signer *BLSSigner) SignRandomSeedHash(randomSeedHash *RandomSeedHash) (*bls.Signature, error) {
	return SignRandomSeedHash(signer.privateKey, randomSeedHash)
}

// TODO: Add signing function for PoS blocks

func (signer *BLSSigner) SignPoSValidatorHandshake(nonceSent uint64, nonceReceived uint64, tstampMicro uint64) (*bls.Signature, error) {
	payload := GetVerackHandshakePayload(nonceSent, nonceReceived, tstampMicro)
	return signer.sign(BLSSignatureOpCodePoSValidatorHandshake, payload[:])
}

//////////////////////////////////////////////////////////
// BLS Verification
//////////////////////////////////////////////////////////

func _blsVerify(opCode BLSSignatureOpCode, payload []byte, signature *bls.Signature, publicKey *bls.PublicKey) (bool, error) {
	newPayload := append(opCode.Bytes(), payload...)
	return publicKey.Verify(signature, newPayload)
}

func BLSVerifyValidatorVote(view uint64, blockHash consensus.BlockHash, signature *bls.Signature, publicKey *bls.PublicKey) (bool, error) {
	payload := consensus.GetVoteSignaturePayload(view, blockHash)
	return _blsVerify(BLSSignatureOpCodeValidatorVote, payload[:], signature, publicKey)
}

func BLSVerifyValidatorTimeout(view uint64, highQCView uint64, signature *bls.Signature, publicKey *bls.PublicKey) (bool, error) {
	payload := consensus.GetTimeoutSignaturePayload(view, highQCView)
	return _blsVerify(BLSSignatureOpCodeValidatorTimeout, payload[:], signature, publicKey)
}

// TODO: Add Verifier function for PoS blocks

func BLSVerifyPoSValidatorHandshake(nonceSent uint64, nonceReceived uint64, tstampMicro uint64,
	signature *bls.Signature, publicKey *bls.PublicKey) (bool, error) {

	payload := GetVerackHandshakePayload(nonceSent, nonceReceived, tstampMicro)
	return _blsVerify(BLSSignatureOpCodePoSValidatorHandshake, payload[:], signature, publicKey)
}
