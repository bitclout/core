//go:build relic

package consensus

import (
	"testing"
	"time"

	"github.com/deso-protocol/core/bls"
	"github.com/deso-protocol/core/collections/bitset"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {

	// Test initial status for newly constructed instance
	{
		fc := NewFastHotStuffEventLoop()
		require.Equal(t, eventLoopStatusNotInitialized, fc.status)
		require.NotPanics(t, fc.Stop) // Calling Stop() on an uninitialized instance should be a no-op
	}

	// Test Init() function with invalid block construction interval
	{
		fc := NewFastHotStuffEventLoop()
		err := fc.Init(0, 1,
			BlockWithValidators{createDummyBlock(2), createDummyValidatorSet()},     // tip
			[]BlockWithValidators{{createDummyBlock(2), createDummyValidatorSet()}}, // safeBlocks
		)
		require.Error(t, err)
	}

	// Test Init() function with invalid timeout duration
	{
		fc := NewFastHotStuffEventLoop()
		err := fc.Init(1, 0,
			BlockWithValidators{createDummyBlock(2), createDummyValidatorSet()},     // tip
			[]BlockWithValidators{{createDummyBlock(2), createDummyValidatorSet()}}, // safeBlocks
		)
		require.Error(t, err)
	}

	// Test Init() function with malformed tip block
	{
		fc := NewFastHotStuffEventLoop()
		err := fc.Init(1, 1,
			BlockWithValidators{nil, createDummyValidatorSet()},                     // tip
			[]BlockWithValidators{{createDummyBlock(2), createDummyValidatorSet()}}, // safeBlocks
		)
		require.Error(t, err)
	}

	// Test Init() function with malformed validator set for tip block
	{
		fc := NewFastHotStuffEventLoop()
		err := fc.Init(1, 1,
			BlockWithValidators{createDummyBlock(2), nil},                           // tip
			[]BlockWithValidators{{createDummyBlock(2), createDummyValidatorSet()}}, // safeBlocks
		)
		require.Error(t, err)
	}

	// Test Init() function with malformed safe block
	{
		fc := NewFastHotStuffEventLoop()
		err := fc.Init(1, 1,
			BlockWithValidators{createDummyBlock(2), createDummyValidatorSet()}, // tip
			[]BlockWithValidators{{nil, createDummyValidatorSet()}},             // safeBlocks
		)
		require.Error(t, err)
	}

	// Test Init() function with malformed validator set for safe block
	{
		fc := NewFastHotStuffEventLoop()
		err := fc.Init(1, 1,
			BlockWithValidators{createDummyBlock(2), createDummyValidatorSet()}, // tip
			[]BlockWithValidators{{createDummyBlock(2), nil}},                   // safeBlocks
		)
		require.Error(t, err)
	}

	// Test Init() function with valid parameters
	{
		block := createDummyBlock(2)

		fc := NewFastHotStuffEventLoop()
		err := fc.Init(100, 101,
			BlockWithValidators{block, createDummyValidatorSet()},     // tip
			[]BlockWithValidators{{block, createDummyValidatorSet()}}, // safeBlocks
		)
		require.NoError(t, err)

		require.Equal(t, eventLoopStatusInitialized, fc.status)

		require.NotPanics(t, fc.Stop) // Calling Stop() on an initialized instance should be a no-op
		require.Equal(t, fc.status, eventLoopStatusInitialized)

		require.Equal(t, fc.tip.block.GetBlockHash().GetValue(), block.GetBlockHash().GetValue())
		require.Equal(t, fc.tip.block.GetView(), uint64(2))
		require.Equal(t, fc.tip.block.GetHeight(), uint64(1))

		require.Equal(t, fc.blockConstructionInterval, time.Duration(100))
		require.Equal(t, fc.timeoutBaseDuration, time.Duration(101))

		require.Equal(t, fc.currentView, uint64(3))
		require.Equal(t, len(fc.tip.validatorSet), 2)
		require.Equal(t, len(fc.tip.validatorLookup), 2)

		require.Equal(t, len(fc.safeBlocks), 1)
		require.Equal(t, fc.safeBlocks[0].block.GetBlockHash().GetValue(), block.GetBlockHash().GetValue())
		require.Equal(t, fc.safeBlocks[0].block.GetView(), uint64(2))
		require.Equal(t, fc.safeBlocks[0].block.GetHeight(), uint64(1))
		require.Equal(t, len(fc.safeBlocks[0].validatorSet), 2)
		require.Equal(t, len(fc.safeBlocks[0].validatorLookup), 2)
	}
}

func TestProcessTipBlock(t *testing.T) {
	oneHourInNanoSecs := time.Duration(3600000000000)

	fc := NewFastHotStuffEventLoop()
	err := fc.Init(oneHourInNanoSecs, oneHourInNanoSecs,
		BlockWithValidators{createDummyBlock(2), createDummyValidatorSet()},     // tip
		[]BlockWithValidators{{createDummyBlock(2), createDummyValidatorSet()}}, // safeBlocks
	)
	require.NoError(t, err)

	// Test ProcessTipBlock() function when event loop is not running
	{
		err := fc.ProcessTipBlock(
			BlockWithValidators{createDummyBlock(2), createDummyValidatorSet()},     // tip
			[]BlockWithValidators{{createDummyBlock(2), createDummyValidatorSet()}}, // safeBlocks
		)
		require.Error(t, err)
	}

	// Start the event loop
	fc.Start()

	// Test ProcessTipBlock() function with malformed tip block
	{
		err := fc.ProcessTipBlock(
			BlockWithValidators{nil, createDummyValidatorSet()},                     // tip
			[]BlockWithValidators{{createDummyBlock(2), createDummyValidatorSet()}}, // safeBlocks
		)
		require.Error(t, err)
	}

	// Test ProcessTipBlock() function with malformed tip validator set
	{
		err := fc.ProcessTipBlock(
			BlockWithValidators{createDummyBlock(2), nil},                           // tip
			[]BlockWithValidators{{createDummyBlock(2), createDummyValidatorSet()}}, // safeBlocks
		)
		require.Error(t, err)
	}

	// Test ProcessTipBlock() function with malformed safe block
	{
		err := fc.ProcessTipBlock(
			BlockWithValidators{createDummyBlock(2), createDummyValidatorSet()}, // tip
			[]BlockWithValidators{{nil, createDummyValidatorSet()}},             // safeBlocks
		)
		require.Error(t, err)
	}

	// Test ProcessTipBlock() function with malformed safe block's validator set
	{
		err := fc.ProcessTipBlock(
			BlockWithValidators{createDummyBlock(2), createDummyValidatorSet()}, // tip
			[]BlockWithValidators{{createDummyBlock(2), nil}},                   // safeBlocks
		)
		require.Error(t, err)
	}

	// Populate the votesSeen and timeoutsSeen maps with dummy data
	{
		fc.votesSeen = map[[32]byte]map[string]VoteMessage{
			{0}: { // blockHash = 0
				"pubKeyA": createDummyVoteMessage(0),
			},
			{1}: { // blockHash = 1
				"pubKeyB": createDummyVoteMessage(1),
			},
			{2}: { // blockHash = 2
				"pubKeyC": createDummyVoteMessage(2),
			},
			{3}: { // blockHash = 3
				"pubKeyD": createDummyVoteMessage(3),
			},
			{4}: { // blockHash = 4
				"pubKeyE": createDummyVoteMessage(4),
			},
		}

		fc.timeoutsSeen = map[uint64]map[string]TimeoutMessage{
			0: { // view = 0
				"pubKeyA": createDummyTimeoutMessage(0),
			},
			1: { // view = 1
				"pubKeyB": createDummyTimeoutMessage(1),
			},
			2: { // view = 2
				"pubKeyC": createDummyTimeoutMessage(2),
			},
			3: { // view = 3
				"pubKeyD": createDummyTimeoutMessage(3),
			},
			4: { // view = 4
				"pubKeyE": createDummyTimeoutMessage(4),
			},
		}
	}

	// Verify the sizes of the votesSeen and timeoutsSeen maps
	{
		require.Equal(t, len(fc.votesSeen), 5)
		require.Equal(t, len(fc.timeoutsSeen), 5)
	}

	// Test ProcessTipBlock() function with valid parameters
	{
		nextBlock := createDummyBlock(2)
		nextBlock.height = 2
		nextBlock.view = 3

		err := fc.ProcessTipBlock(
			BlockWithValidators{nextBlock, createDummyValidatorSet()},     // tip
			[]BlockWithValidators{{nextBlock, createDummyValidatorSet()}}, // safeBlocks
		)
		require.NoError(t, err)

		require.Equal(t, nextBlock.GetBlockHash().GetValue(), fc.tip.block.GetBlockHash().GetValue())
		require.Equal(t, uint64(3), fc.tip.block.GetView())
		require.Equal(t, uint64(2), fc.tip.block.GetHeight())

		require.Equal(t, uint64(4), fc.currentView)
		require.Equal(t, 2, len(fc.tip.validatorSet))
	}

	// Verify that stale votes and timeouts have been evicted
	{
		require.Equal(t, 2, len(fc.votesSeen))
		require.Equal(t, 2, len(fc.timeoutsSeen))
	}

	// Stop the event loop
	fc.Stop()
}

func TestAdvanceView(t *testing.T) {
	oneHourInNanoSecs := time.Duration(3600000000000)

	fc := NewFastHotStuffEventLoop()

	// BlockHeight = 1, Current View = 3
	err := fc.Init(oneHourInNanoSecs, oneHourInNanoSecs,
		BlockWithValidators{createDummyBlock(2), createDummyValidatorSet()},     // tip
		[]BlockWithValidators{{createDummyBlock(2), createDummyValidatorSet()}}, // safeBlocks
	)
	require.NoError(t, err)

	// Running AdvanceView() should fail because the event loop is not running
	{
		_, err := fc.AdvanceView()
		require.Error(t, err)
	}

	// Start the event loop
	fc.Start()

	// Populate the votesSeen and timeoutsSeen maps with dummy data
	{
		fc.votesSeen = map[[32]byte]map[string]VoteMessage{
			{1}: { // blockHash = 1
				"pubKeyA": createDummyVoteMessage(1),
			},
			{2}: { // blockHash = 2
				"pubKeyB": createDummyVoteMessage(2),
			},
			{3}: { // blockHash = 3
				"pubKeyC": createDummyVoteMessage(3),
			},
			{4}: { // blockHash = 4
				"pubKeyD": createDummyVoteMessage(4),
			},
			{5}: { // blockHash = 5
				"pubKeyE": createDummyVoteMessage(5),
			},
		}

		fc.timeoutsSeen = map[uint64]map[string]TimeoutMessage{
			1: { // view = 1
				"pubKeyA": createDummyTimeoutMessage(1),
			},
			2: { // view = 2
				"pubKeyB": createDummyTimeoutMessage(2),
			},
			3: { // view = 3
				"pubKeyC": createDummyTimeoutMessage(3),
			},
			4: { // view = 4
				"pubKeyD": createDummyTimeoutMessage(4),
			},
			5: { // view = 5
				"pubKeyE": createDummyTimeoutMessage(5),
			},
		}
	}

	// Run AdvanceView() to view 4
	{
		newView, err := fc.AdvanceView()
		require.NoError(t, err)
		require.Equal(t, uint64(4), newView)
	}

	// Verify that vote and timeout messages haven't changed
	{
		require.Equal(t, len(fc.votesSeen), 3)
		require.Equal(t, len(fc.timeoutsSeen), 3)
	}

	// Run AdvanceView() to view 5
	{
		newView, err := fc.AdvanceView()
		require.NoError(t, err)
		require.Equal(t, uint64(5), newView)
	}

	// Verify that stale votes and timeouts have been evicted
	{
		require.Equal(t, len(fc.votesSeen), 2)
		require.Equal(t, len(fc.timeoutsSeen), 2)
	}

	// Stop the event loop
	fc.Stop()
}

func TestProcessValidatorVote(t *testing.T) {
	oneHourInNanoSecs := time.Duration(3600000000000)

	fc := NewFastHotStuffEventLoop()

	// BlockHeight = 1, Current View = 3
	err := fc.Init(oneHourInNanoSecs, oneHourInNanoSecs,
		BlockWithValidators{createDummyBlock(2), createDummyValidatorSet()},     // tip
		[]BlockWithValidators{{createDummyBlock(2), createDummyValidatorSet()}}, // safeBlocks
	)
	require.NoError(t, err)

	// Start the event loop
	fc.Start()

	// Current View = 4
	{
		currentView, err := fc.AdvanceView()
		require.NoError(t, err)
		require.Equal(t, uint64(4), currentView)
	}

	// Test with malformed vote
	{
		err := fc.ProcessValidatorVote(nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "Malformed vote message")
	}

	// Test invalid signature
	{
		vote := createDummyVoteMessage(4)
		vote.signature = createDummyBLSSignature()
		err := fc.ProcessValidatorVote(vote)
		require.Error(t, err)
		require.Contains(t, err.Error(), "Invalid signature")
	}

	// Test with stale view
	{
		vote := createDummyVoteMessage(1)
		err := fc.ProcessValidatorVote(vote)
		require.Error(t, err)
		require.Contains(t, err.Error(), "Vote has a stale view")
	}

	// Test when we've already seen a vote from the validator for the same view
	{
		vote := createDummyVoteMessage(4)
		fc.votesSeen[GetVoteSignaturePayload(vote.GetView(), vote.GetBlockHash())] = map[string]VoteMessage{
			vote.publicKey.ToString(): vote,
		}

		err := fc.ProcessValidatorVote(vote)
		require.Error(t, err)
		require.Contains(t, err.Error(), "has already voted for view")
	}

	// Test when we've already seen a timeout from the validator for the same view
	{
		vote := createDummyVoteMessage(5)
		timeout := createDummyTimeoutMessage(5)
		timeout.publicKey = vote.publicKey

		fc.timeoutsSeen[timeout.GetView()] = map[string]TimeoutMessage{
			timeout.publicKey.ToString(): timeout,
		}

		err := fc.ProcessValidatorVote(vote)
		require.Error(t, err)
		require.Contains(t, err.Error(), "has already timed out for view")
	}

	// Test happy path
	{
		vote := createDummyVoteMessage(4)
		err := fc.ProcessValidatorVote(vote)
		require.NoError(t, err)
	}

	// Stop the event loop
	fc.Stop()
}

func TestProcessValidatorTimeout(t *testing.T) {
	oneHourInNanoSecs := time.Duration(3600000000000)

	fc := NewFastHotStuffEventLoop()

	// BlockHeight = 1, Current View = 3
	err := fc.Init(oneHourInNanoSecs, oneHourInNanoSecs,
		BlockWithValidators{createDummyBlock(2), createDummyValidatorSet()},     // tip
		[]BlockWithValidators{{createDummyBlock(2), createDummyValidatorSet()}}, // safeBlocks
	)
	require.NoError(t, err)

	// Start the event loop
	fc.Start()

	// Current View = 4
	{
		currentView, err := fc.AdvanceView()
		require.NoError(t, err)
		require.Equal(t, uint64(4), currentView)
	}

	// Test with malformed timeout
	{
		err := fc.ProcessValidatorTimeout(nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "Malformed timeout message")
	}

	// Test invalid signature
	{
		timeout := createDummyTimeoutMessage(4)
		timeout.signature = createDummyBLSSignature()
		err := fc.ProcessValidatorTimeout(timeout)
		require.Error(t, err)
		require.Contains(t, err.Error(), "Invalid signature")
	}

	// Test with stale view
	{
		timeout := createDummyTimeoutMessage(1)
		err := fc.ProcessValidatorTimeout(timeout)
		require.Error(t, err)
		require.Contains(t, err.Error(), "Timeout has a stale view")
	}

	// Test when we've already seen a vote from the validator for the same view
	{
		timeout := createDummyTimeoutMessage(4)
		vote := createDummyVoteMessage(4)

		fc.votesSeen[GetVoteSignaturePayload(vote.GetView(), vote.GetBlockHash())] = map[string]VoteMessage{
			timeout.publicKey.ToString(): vote,
		}

		err = fc.ProcessValidatorTimeout(timeout)
		require.Error(t, err)
		require.Contains(t, err.Error(), "has already voted for view")
	}

	// Test when we've already seen a timeout from the validator for the same view
	{
		timeout := createDummyTimeoutMessage(4)

		fc.timeoutsSeen[timeout.view] = map[string]TimeoutMessage{
			timeout.publicKey.ToString(): timeout,
		}

		err = fc.ProcessValidatorTimeout(timeout)
		require.Error(t, err)
		require.Contains(t, err.Error(), "has already timed out for view")
	}

	// Test happy path
	{
		timeout := createDummyTimeoutMessage(4)
		err := fc.ProcessValidatorTimeout(timeout)
		require.NoError(t, err)
	}

	// Stop the event loop
	fc.Stop()
}

func TestTimeoutScheduledTaskExecuted(t *testing.T) {
	oneHourInNanoSecs := time.Duration(3600000000000)
	oneMilliSecondInNanoSeconds := time.Duration(1000000)

	dummyBlock := createDummyBlock(2)

	fc := NewFastHotStuffEventLoop()
	err := fc.Init(oneHourInNanoSecs, oneMilliSecondInNanoSeconds,
		BlockWithValidators{dummyBlock, createDummyValidatorSet()},     // tip
		[]BlockWithValidators{{dummyBlock, createDummyValidatorSet()}}, // safeBlocks
	)
	require.NoError(t, err)

	// Start the event loop
	fc.Start()

	// Wait for the timeout signal to be sent
	timeoutSignal := <-fc.Events

	// Confirm that the timeout signal is for the the expected view
	require.Equal(t, timeoutSignal.EventType, FastHotStuffEventTypeTimeout)
	require.Equal(t, timeoutSignal.View, dummyBlock.GetView()+1)
	require.Equal(t, timeoutSignal.TipBlockHash.GetValue(), dummyBlock.GetBlockHash().GetValue())

	// Confirm that the timeout is no longer running
	require.False(t, fc.nextTimeoutTask.IsScheduled())

	// Advance the view, which should reset the timeout scheduled task
	fc.AdvanceView()

	// Wait for the timeout signal to be sent
	timeoutSignal = <-fc.Events

	// Confirm that the timeout signal is for the the expected view
	require.Equal(t, timeoutSignal.EventType, FastHotStuffEventTypeTimeout)
	require.Equal(t, timeoutSignal.View, dummyBlock.GetView()+2)
	require.Equal(t, timeoutSignal.TipBlockHash.GetValue(), dummyBlock.GetBlockHash().GetValue())

	// Confirm that the timeout is no longer running
	require.False(t, fc.nextTimeoutTask.IsScheduled())

	// Stop the event loop
	fc.Stop()
}

func TestResetEventLoopSignal(t *testing.T) {
	oneHourInNanoSecs := time.Duration(3600000000000)

	fc := NewFastHotStuffEventLoop()
	err := fc.Init(oneHourInNanoSecs, 2*oneHourInNanoSecs,
		BlockWithValidators{createDummyBlock(2), createDummyValidatorSet()},     // tip
		[]BlockWithValidators{{createDummyBlock(2), createDummyValidatorSet()}}, // safeBlocks
	)
	require.NoError(t, err)

	// Start the event loop
	fc.Start()

	// Confirm the ETAs for the block construction and timeout timers
	require.Equal(t, fc.nextBlockConstructionTask.GetDuration(), oneHourInNanoSecs) // 1 hour away
	require.Equal(t, fc.nextTimeoutTask.GetDuration(), 2*oneHourInNanoSecs)         // 2 hours away

	// Advance the view to simulate a timeout
	_, err = fc.AdvanceView()
	require.NoError(t, err)

	// Confirm the ETAs for the block construction and timeout timers
	require.Equal(t, fc.nextBlockConstructionTask.GetDuration(), oneHourInNanoSecs) // 1 hour away
	require.Equal(t, fc.nextTimeoutTask.GetDuration(), 4*oneHourInNanoSecs)         // 2 hours * 2 = 4 hours away

	// Advance the view to simulate a 2nd timeout
	_, err = fc.AdvanceView()
	require.NoError(t, err)

	// Confirm the ETAs for the block construction and timeout timers
	require.Equal(t, fc.nextBlockConstructionTask.GetDuration(), oneHourInNanoSecs) // 1 hour away
	require.Equal(t, fc.nextTimeoutTask.GetDuration(), 8*oneHourInNanoSecs)         // 2 hours * 2^2 = 8 hours away

	// Advance the view to simulate a 3nd timeout
	_, err = fc.AdvanceView()
	require.NoError(t, err)

	// Confirm the ETAs for the block construction and timeout timers
	require.Equal(t, fc.nextBlockConstructionTask.GetDuration(), oneHourInNanoSecs) // 1 hour away
	require.Equal(t, fc.nextTimeoutTask.GetDuration(), 16*oneHourInNanoSecs)        // 2 hours * 2^3 = 16 hours away

	// Stop the event loop
	fc.Stop()
}

func TestVoteQCConstructionSignal(t *testing.T) {

	// Create a valid dummy block at view 2
	block := createDummyBlock(2)

	// Create a valid validator set
	validatorPrivateKey1, _ := bls.NewPrivateKey()
	validatorPrivateKey2, _ := bls.NewPrivateKey()

	validatorSet := []Validator{
		&validator{
			publicKey:   validatorPrivateKey1.PublicKey(),
			stakeAmount: uint256.NewInt().SetUint64(70),
		},
		&validator{
			publicKey:   validatorPrivateKey2.PublicKey(),
			stakeAmount: uint256.NewInt().SetUint64(30),
		},
	}

	voteSignaturePayload := GetVoteSignaturePayload(2, block.GetBlockHash())

	validator1Vote, _ := validatorPrivateKey1.Sign(voteSignaturePayload[:])
	validator2Vote, _ := validatorPrivateKey2.Sign(voteSignaturePayload[:])

	// Sad path, not enough votes to construct a QC
	{
		fc := NewFastHotStuffEventLoop()
		err := fc.Init(time.Microsecond, time.Hour,
			BlockWithValidators{block, validatorSet},     // tip
			[]BlockWithValidators{{block, validatorSet}}, // safeBlocks
		)
		require.NoError(t, err)

		// Create a vote from validator 2
		vote := voteMessage{
			view:      2,                                // The view the block was proposed in
			blockHash: block.GetBlockHash(),             // Block hash is the same as the block hash of the dummy block
			publicKey: validatorPrivateKey2.PublicKey(), // Validator 2 with 30/100 stake votes
			signature: validator2Vote,                   // Validator 2's vote
		}

		// Store the vote in the event loop's votesSeen map
		fc.votesSeen[voteSignaturePayload] = map[string]VoteMessage{
			vote.publicKey.ToString(): &vote,
		}

		// Start the event loop
		fc.Start()

		// Wait up to 100 milliseconds for a block construction signal to be sent
		select {
		case <-fc.Events:
			require.Fail(t, "Received a block construction signal when there were not enough votes to construct a QC")
		case <-time.After(100 * time.Millisecond):
			// Do nothing
		}

		// Stop the event loop
		fc.Stop()
	}

	// Happy path, there are votes with a super-majority of stake to construct a QC
	{
		fc := NewFastHotStuffEventLoop()
		err := fc.Init(time.Microsecond, time.Hour,
			BlockWithValidators{block, validatorSet},     // tip
			[]BlockWithValidators{{block, validatorSet}}, // safeBlocks
		)
		require.NoError(t, err)

		// Create a vote from validator 1
		vote := voteMessage{
			view:      2,                                // The view the block was proposed in
			blockHash: block.GetBlockHash(),             // Block hash is the same as the block hash of the dummy block
			publicKey: validatorPrivateKey1.PublicKey(), // Validator 1 with 70/100 stake votes
			signature: validator1Vote,                   // Validator 1's vote
		}

		// Store the vote in the event loop's votesSeen map
		fc.votesSeen[voteSignaturePayload] = map[string]VoteMessage{
			vote.publicKey.ToString(): &vote,
		}

		// Start the event loop
		fc.Start()

		var blockConstructionSignal *FastHotStuffEvent

		// Wait up to 100 milliseconds for a block construction signal to be sent
		select {
		case blockConstructionSignal = <-fc.Events:
			// Do nothing
		case <-time.After(100 * time.Millisecond):
			require.Fail(t, "Did not receive a block construction signal when there were enough votes to construct a QC")
		}

		// Stop the event loop
		fc.Stop()

		// Confirm that the block construction signal has the expected parameters
		require.Equal(t, blockConstructionSignal.EventType, FastHotStuffEventTypeConstructVoteQC)
		require.Equal(t, blockConstructionSignal.View, block.GetView()+1)
		require.Equal(t, blockConstructionSignal.TipBlockHash.GetValue(), block.GetBlockHash().GetValue())
		require.Equal(t, blockConstructionSignal.TipBlockHeight, block.GetHeight())
		require.Equal(t, blockConstructionSignal.QC.GetView(), block.GetView())
		require.Equal(t, blockConstructionSignal.QC.GetBlockHash().GetValue(), block.GetBlockHash().GetValue())
		require.Equal(t, blockConstructionSignal.QC.GetAggregatedSignature().GetSignersList().ToBytes(), bitset.NewBitset().Set(0, true).ToBytes())
		require.Equal(t, blockConstructionSignal.QC.GetAggregatedSignature().GetSignature().ToBytes(), validator1Vote.ToBytes())
	}
}

func TestTimeoutQCConstructionSignal(t *testing.T) {
	// Create a valid dummy block at view 2
	block1 := createDummyBlock(2)

	// Create a valid dummy block that extends from the above block at view 3
	block2 := &block{
		blockHash: createDummyBlockHash(),
		view:      3,
		height:    2,
		qc:        createDummyQC(2, block1.GetBlockHash()),
	}

	// Create a valid validator set
	validatorPrivateKey1, _ := bls.NewPrivateKey()
	validatorPrivateKey2, _ := bls.NewPrivateKey()

	validatorSet := []Validator{
		&validator{
			publicKey:   validatorPrivateKey1.PublicKey(),
			stakeAmount: uint256.NewInt().SetUint64(70),
		},
		&validator{
			publicKey:   validatorPrivateKey2.PublicKey(),
			stakeAmount: uint256.NewInt().SetUint64(30),
		},
	}

	// Both validators will timeout for view 4. Validator 1 will timeout with a highQC from view 2, and
	// validator 2 will timeout with a highQC from view 3
	timeoutSignaturePayload1 := GetTimeoutSignaturePayload(4, 1)
	timeoutSignaturePayload2 := GetTimeoutSignaturePayload(4, 2)

	validator1TimeoutSignature, _ := validatorPrivateKey1.Sign(timeoutSignaturePayload1[:])
	validator2TimeoutSignature, _ := validatorPrivateKey2.Sign(timeoutSignaturePayload2[:])

	// Sad path, not enough timeouts to construct a timeout QC
	{
		fc := NewFastHotStuffEventLoop()
		err := fc.Init(time.Microsecond, time.Hour,
			BlockWithValidators{block2, validatorSet}, // tip
			[]BlockWithValidators{ // safeBlocks
				{block1, validatorSet},
				{block2, validatorSet},
			},
		)
		require.NoError(t, err)

		// Manually set the view to view 5 to simulate a timeout at view 4
		fc.currentView = 5

		// Create a timeout message from validator 2
		timeout := timeoutMessage{
			view:      4,                                // The view which the validator is timing out for
			highQC:    block2.GetQC(),                   // The highest QC the validator has seen
			publicKey: validatorPrivateKey2.PublicKey(), // Validator 2 with 30/100 stake
			signature: validator2TimeoutSignature,       // Validator 2's timeout signature on payload (view 4, highQCview 2)
		}

		// Store the timeout in the event loop's timeoutsSeen map
		fc.timeoutsSeen[4] = map[string]TimeoutMessage{
			timeout.publicKey.ToString(): &timeout,
		}

		// Start the event loop
		fc.Start()

		// Wait up to 100 milliseconds for a block construction signal to be sent
		select {
		case <-fc.Events:
			require.Fail(t, "Received a block construction signal when there were not enough timeouts to construct a timeout QC")
		case <-time.After(100 * time.Millisecond):
			// Do nothing
		}

		// Stop the event loop
		fc.Stop()
	}

	// Happy path, there are enough timeouts with a super-majority of stake to construct a timeout QC
	{
		fc := NewFastHotStuffEventLoop()
		err := fc.Init(time.Microsecond, time.Hour,
			BlockWithValidators{block2, validatorSet}, // tip
			[]BlockWithValidators{ // safeBlocks
				{block1, validatorSet},
				{block2, validatorSet},
			},
		)
		require.NoError(t, err)

		// Manually set the currentView to 5 to simulate a timeout on view 4
		fc.currentView = 5

		// Create a timeout message from validator 1 with highQC from block 1
		timeout1 := timeoutMessage{
			view:      4,                                // The view which the validator is timing out for
			highQC:    block1.GetQC(),                   // The highest QC this validator has seen
			publicKey: validatorPrivateKey1.PublicKey(), // Validator 1 with 70/100 stake
			signature: validator1TimeoutSignature,       // Validator 1's timeout signature on payload (view 4, highQCview 1)
		}

		// Create a timeout message from validator 2 with highQC from block 2
		timeout2 := timeoutMessage{
			view:      4,                                // The view which the validator is timing out for
			highQC:    block2.GetQC(),                   // The highest QC the validator has seen
			publicKey: validatorPrivateKey2.PublicKey(), // Validator 2 with 30/100 stake
			signature: validator2TimeoutSignature,       // Validator 2's timeout signature on payload (view 4, highQCview 2)
		}

		// Store the timeout in the event loop's timeoutsSeen map
		fc.timeoutsSeen[4] = map[string]TimeoutMessage{
			timeout1.publicKey.ToString(): &timeout1,
			timeout2.publicKey.ToString(): &timeout2,
		}

		// Start the event loop
		fc.Start()

		var signal *FastHotStuffEvent

		// Wait up to 100 milliseconds for a block construction signal to be sent
		select {
		case signal = <-fc.Events:
			// Do nothing
		case <-time.After(100 * time.Second):
			require.Fail(t, "Did not receive a block construction signal when there were enough timeouts to construct a timeout QC")
		}

		// Stop the event loop
		fc.Stop()

		// Confirm that the block construction signal has the expected parameters
		require.Equal(t, signal.EventType, FastHotStuffEventTypeConstructTimeoutQC)
		require.Equal(t, signal.View, uint64(5))                                           // The timeout QC will be proposed in view 5
		require.Equal(t, signal.TipBlockHash.GetValue(), block1.GetBlockHash().GetValue()) // The timeout QC will be proposed in a block that extends from block 1
		require.Equal(t, signal.TipBlockHeight, block1.GetHeight())                        // The timeout QC will be proposed at the block height after block 1
		require.Equal(t, signal.AggregateQC.GetView(), uint64(4))                          // The timed out view is 4
		require.Equal(t, signal.AggregateQC.GetHighQCViews(), []uint64{1, 2})              // The high QC view is 1 for validator 1 and 2 for validator 2
		require.Equal(t,
			signal.AggregateQC.GetAggregatedSignature().GetSignersList().ToBytes(),
			bitset.NewBitset().Set(0, true).Set(1, true).ToBytes(), // Both validators have timed out, so both validators are in the timeout QC
		)

		// Verify that the high QC is the QC from block 2. It should be unchanged.
		require.Equal(t, signal.AggregateQC.GetHighQC().GetBlockHash(), block2.GetQC().GetBlockHash())
		require.Equal(t, signal.AggregateQC.GetHighQC().GetView(), block2.GetQC().GetView())
		require.Equal(t,
			signal.AggregateQC.GetHighQC().GetAggregatedSignature().GetSignersList().ToBytes(),
			block2.GetQC().GetAggregatedSignature().GetSignersList().ToBytes(),
		)
		require.Equal(t,
			signal.AggregateQC.GetHighQC().GetAggregatedSignature().GetSignature().ToBytes(),
			block2.GetQC().GetAggregatedSignature().GetSignature().ToBytes(),
		)
	}
}

func TestFastHotStuffEventLoopStartStop(t *testing.T) {
	oneHourInNanoSecs := time.Duration(3600000000000)

	fc := NewFastHotStuffEventLoop()
	err := fc.Init(oneHourInNanoSecs, 2*oneHourInNanoSecs,
		BlockWithValidators{createDummyBlock(2), createDummyValidatorSet()},     // tip
		[]BlockWithValidators{{createDummyBlock(2), createDummyValidatorSet()}}, // safeBlocks
	)
	require.NoError(t, err)

	// Start the event loop
	fc.Start()

	// Confirm the event loop status has changed to running
	require.Equal(t, eventLoopStatusRunning, fc.status)

	// Confirm that the ETAs for the block construction and timeout timers have been set
	require.Equal(t, fc.nextBlockConstructionTask.GetDuration(), oneHourInNanoSecs)
	require.Equal(t, fc.nextTimeoutTask.GetDuration(), 2*oneHourInNanoSecs)

	// Stop the event loop
	fc.Stop()

	// Confirm the event loop status has reverted to initialized
	require.Equal(t, eventLoopStatusInitialized, fc.status)

	// Confirm that calling fc.Stop() again doesn't panic
	require.NotPanics(t, fc.Stop)
}
