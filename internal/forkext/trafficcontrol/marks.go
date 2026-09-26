package trafficcontrol

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

const (
	markMin uint32 = 1
	markMax uint32 = 65534
)

// AllocateMark is stable for a node/client pair and refuses a collision rather
// than silently assigning a second identity to an existing kernel mark.
func AllocateMark(nodeKey, clientKey string, occupied map[uint32]string) (uint32, error) {
	if nodeKey == "" || clientKey == "" {
		return 0, fmt.Errorf("node and client keys are required")
	}
	h := sha256.Sum256([]byte(nodeKey + "\x00" + clientKey))
	span := markMax - markMin + 1
	mark := markMin + binary.BigEndian.Uint32(h[:4])%span
	if owner, ok := occupied[mark]; ok && owner != nodeKey+"\x00"+clientKey {
		return 0, fmt.Errorf("mark collision: %d belongs to %q", mark, owner)
	}
	return mark, nil
}

func ownedName(prefix, nodeKey, clientKey string) string {
	h := sha256.Sum256([]byte(nodeKey + "\x00" + clientKey))
	return fmt.Sprintf("%s-%x", prefix, h[:6])
}

func RuleName(nodeKey, clientKey string) string { return ownedName("catx-rule", nodeKey, clientKey) }
