package integration_testing

import (
	"fmt"
	"github.com/deso-protocol/core/cmd"
	"github.com/deso-protocol/core/lib"
	"testing"
)

func waitForValidatorConnection(t *testing.T, node1 *cmd.Node, node2 *cmd.Node) {
	userAgentN1 := node1.Params.UserAgent
	userAgentN2 := node2.Params.UserAgent
	nmN1 := node1.Server.GetNetworkManager()
	n1ValidatedN2 := func() bool {
		if true != checkRemoteNodeIndexerUserAgent(nmN1, userAgentN2, true, false, false) {
			return false
		}
		rnFromN2 := getRemoteNodeWithUserAgent(node1, userAgentN2)
		if rnFromN2 == nil {
			return false
		}
		if !rnFromN2.IsHandshakeCompleted() {
			return false
		}
		return true
	}
	waitForCondition(t, fmt.Sprintf("Waiting for Node (%s) to connect to validator Node (%s)", userAgentN1, userAgentN2), n1ValidatedN2)
}

func waitForNonValidatorOutboundConnection(t *testing.T, node1 *cmd.Node, node2 *cmd.Node) {
	userAgentN1 := node1.Params.UserAgent
	userAgentN2 := node2.Params.UserAgent
	condition := conditionNonValidatorOutboundConnection(t, node1, node2)
	waitForCondition(t, fmt.Sprintf("Waiting for Node (%s) to connect to outbound non-validator Node (%s)", userAgentN1, userAgentN2), condition)
}

func conditionNonValidatorOutboundConnection(t *testing.T, node1 *cmd.Node, node2 *cmd.Node) func() bool {
	return conditionNonValidatorOutboundConnectionDynamic(t, node1, node2, false)
}

func conditionNonValidatorOutboundConnectionDynamic(t *testing.T, node1 *cmd.Node, node2 *cmd.Node, inactiveValidator bool) func() bool {
	userAgentN2 := node2.Params.UserAgent
	nmN1 := node1.Server.GetNetworkManager()
	return func() bool {
		if true != checkRemoteNodeIndexerUserAgent(nmN1, userAgentN2, false, true, false) {
			return false
		}
		rnFromN2 := getRemoteNodeWithUserAgent(node1, userAgentN2)
		if rnFromN2 == nil {
			return false
		}
		if !rnFromN2.IsHandshakeCompleted() {
			return false
		}
		// inactiveValidator should have the public key.
		if inactiveValidator {
			return rnFromN2.GetValidatorPublicKey() != nil
		}
		return rnFromN2.GetValidatorPublicKey() == nil
	}
}

func waitForNonValidatorInboundConnection(t *testing.T, node1 *cmd.Node, node2 *cmd.Node) {
	userAgentN1 := node1.Params.UserAgent
	userAgentN2 := node2.Params.UserAgent
	condition := conditionNonValidatorInboundConnection(t, node1, node2)
	waitForCondition(t, fmt.Sprintf("Waiting for Node (%s) to connect to inbound non-validator Node (%s)", userAgentN1, userAgentN2), condition)
}

func waitForNonValidatorInboundConnectionDynamic(t *testing.T, node1 *cmd.Node, node2 *cmd.Node, inactiveValidator bool) {
	userAgentN1 := node1.Params.UserAgent
	userAgentN2 := node2.Params.UserAgent
	condition := conditionNonValidatorInboundConnectionDynamic(t, node1, node2, inactiveValidator)
	waitForCondition(t, fmt.Sprintf("Waiting for Node (%s) to connect to inbound non-validator Node (%s), "+
		"inactiveValidator (%v)", userAgentN1, userAgentN2, inactiveValidator), condition)
}

func conditionNonValidatorInboundConnection(t *testing.T, node1 *cmd.Node, node2 *cmd.Node) func() bool {
	return conditionNonValidatorInboundConnectionDynamic(t, node1, node2, false)
}

func conditionNonValidatorInboundConnectionDynamic(t *testing.T, node1 *cmd.Node, node2 *cmd.Node, inactiveValidator bool) func() bool {
	userAgentN2 := node2.Params.UserAgent
	nmN1 := node1.Server.GetNetworkManager()
	return func() bool {
		if true != checkRemoteNodeIndexerUserAgent(nmN1, userAgentN2, false, false, true) {
			return false
		}
		rnFromN2 := getRemoteNodeWithUserAgent(node1, userAgentN2)
		if rnFromN2 == nil {
			return false
		}
		if !rnFromN2.IsHandshakeCompleted() {
			return false
		}
		// inactiveValidator should have the public key.
		if inactiveValidator {
			return rnFromN2.GetValidatorPublicKey() != nil
		}
		return rnFromN2.GetValidatorPublicKey() == nil
	}
}

func checkInactiveValidatorConnection(t *testing.T, node1 *cmd.Node, node2 *cmd.Node) bool {
	userAgentN2 := node2.Params.UserAgent
	nmN1 := node1.Server.GetNetworkManager()
	if true != checkRemoteNodeIndexerUserAgent(nmN1, userAgentN2, false, true, true) {
		return false
	}
	rnFromN2 := getRemoteNodeWithUserAgent(node1, userAgentN2)
	if rnFromN2 == nil {
		return false
	}
	if !rnFromN2.IsHandshakeCompleted() {
		return false
	}
	return rnFromN2.GetValidatorPublicKey() != nil
}

func waitForEmptyRemoteNodeIndexer(t *testing.T, node1 *cmd.Node) {
	userAgentN1 := node1.Params.UserAgent
	nmN1 := node1.Server.GetNetworkManager()
	n1ValidatedN2 := func() bool {
		if true != checkRemoteNodeIndexerEmpty(nmN1) {
			return false
		}
		return true
	}
	waitForCondition(t, fmt.Sprintf("Waiting for Node (%s) to disconnect from all RemoteNodes", userAgentN1), n1ValidatedN2)
}

func waitForCountRemoteNodeIndexer(t *testing.T, node1 *cmd.Node, allCount int, validatorCount int,
	nonValidatorOutboundCount int, nonValidatorInboundCount int) {

	userAgent := node1.Params.UserAgent
	nm := node1.Server.GetNetworkManager()
	condition := func() bool {
		if true != checkRemoteNodeIndexerCount(nm, allCount, validatorCount, nonValidatorOutboundCount, nonValidatorInboundCount) {
			return false
		}
		return true
	}
	waitForCondition(t, fmt.Sprintf("Waiting for Node (%s) to have appropriate RemoteNodes counts", userAgent), condition)
}

func waitForCountRemoteNodeIndexerHandshakeCompleted(t *testing.T, node1 *cmd.Node, allCount, validatorCount int,
	nonValidatorOutboundCount int, nonValidatorInboundCount int) {

	userAgent := node1.Params.UserAgent
	nm := node1.Server.GetNetworkManager()
	condition := func() bool {
		return checkRemoteNodeIndexerCountHandshakeCompleted(nm, allCount, validatorCount,
			nonValidatorOutboundCount, nonValidatorInboundCount)
	}
	waitForCondition(t, fmt.Sprintf("Waiting for Node (%s) to have appropriate RemoteNodes counts", userAgent), condition)
}

func checkRemoteNodeIndexerUserAgent(manager *lib.NetworkManager, userAgent string, validator bool,
	nonValidatorOutbound bool, nonValidatorInbound bool) bool {

	if true != checkUserAgentInRemoteNodeList(userAgent, manager.GetAllRemoteNodes().GetAll()) {
		return false
	}
	if validator != checkUserAgentInRemoteNodeList(userAgent, manager.GetAllValidators().GetAll()) {
		return false
	}
	if nonValidatorOutbound != checkUserAgentInRemoteNodeList(userAgent, manager.GetNonValidatorOutboundIndex().GetAll()) {
		return false
	}
	if nonValidatorInbound != checkUserAgentInRemoteNodeList(userAgent, manager.GetNonValidatorInboundIndex().GetAll()) {
		return false
	}

	return true
}

func checkRemoteNodeIndexerCount(manager *lib.NetworkManager, allCount int, validatorCount int,
	nonValidatorOutboundCount int, nonValidatorInboundCount int) bool {

	if allCount != manager.GetAllRemoteNodes().Count() {
		return false
	}
	if validatorCount != manager.GetAllValidators().Count() {
		return false
	}
	if nonValidatorOutboundCount != manager.GetNonValidatorOutboundIndex().Count() {
		return false
	}
	if nonValidatorInboundCount != manager.GetNonValidatorInboundIndex().Count() {
		return false
	}

	return true
}

func checkRemoteNodeIndexerCountHandshakeCompleted(manager *lib.NetworkManager, allCount int, validatorCount int,
	nonValidatorOutboundCount int, nonValidatorInboundCount int) bool {

	if allCount != manager.GetAllRemoteNodes().Count() {
		return false
	}
	allValidators := manager.GetAllValidators()
	if validatorCount != allValidators.Count() {
		return false
	}
	for _, rn := range allValidators.GetAll() {
		if !rn.IsHandshakeCompleted() {
			return false
		}
	}

	if nonValidatorOutboundCount != manager.GetNonValidatorOutboundIndex().Count() {
		return false
	}
	for _, rn := range manager.GetNonValidatorOutboundIndex().GetAll() {
		if !rn.IsHandshakeCompleted() {
			return false
		}
	}

	if nonValidatorInboundCount != manager.GetNonValidatorInboundIndex().Count() {
		return false
	}
	for _, rn := range manager.GetNonValidatorInboundIndex().GetAll() {
		if !rn.IsHandshakeCompleted() {
			return false
		}
	}

	return true
}

func checkRemoteNodeIndexerEmpty(manager *lib.NetworkManager) bool {
	if manager.GetAllRemoteNodes().Count() != 0 {
		return false
	}
	if manager.GetAllValidators().Count() != 0 {
		return false
	}
	if manager.GetNonValidatorOutboundIndex().Count() != 0 {
		return false
	}
	if manager.GetNonValidatorInboundIndex().Count() != 0 {
		return false
	}
	return true
}

func checkUserAgentInRemoteNodeList(userAgent string, rnList []*lib.RemoteNode) bool {
	for _, rn := range rnList {
		if rn == nil {
			continue
		}
		if rn.GetUserAgent() == userAgent {
			return true
		}
	}
	return false
}

func getRemoteNodeWithUserAgent(node *cmd.Node, userAgent string) *lib.RemoteNode {
	nm := node.Server.GetNetworkManager()
	rnList := nm.GetAllRemoteNodes().GetAll()
	for _, rn := range rnList {
		if rn.GetUserAgent() == userAgent {
			return rn
		}
	}
	return nil
}
