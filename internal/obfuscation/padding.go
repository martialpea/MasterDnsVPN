// ==============================================================================
// MasterDnsVPN — Innovation: Traffic Obfuscation
// Pads outbound DNS-tunneled packets to canonical DNS response sizes so that
// DPI classifiers cannot distinguish tunnel traffic from legitimate DNS.
// Standard DNS UDP sizes: 28, 56, 120, 248, 512, 1232 bytes.
// ==============================================================================
package obfuscation

import "crypto/rand"

// canonicalSizes are common DNS UDP payload sizes observed in the wild.
var canonicalSizes = []int{28, 56, 120, 248, 512, 1232}

// PadToCanonical pads data with random bytes to the next canonical DNS size.
// Returns the original slice unchanged if it already exceeds all sizes.
// The receiver must strip padding using the embedded length prefix.
func PadToCanonical(data []byte) []byte {
	target := 0
	for _, s := range canonicalSizes {
		if len(data) <= s {
			target = s
			break
		}
	}
	if target == 0 || target == len(data) {
		return data
	}

	padded := make([]byte, target)
	copy(padded, data)
	// Fill remainder with random bytes to prevent pattern matching on zeros.
	if _, err := rand.Read(padded[len(data):]); err != nil {
		// Fallback: fill with 0x00 — still better than no padding.
		for i := len(data); i < target; i++ {
			padded[i] = 0
		}
	}
	return padded
}

// StripPadding extracts the original payload using the length prefix written
// by the encoder (first 2 bytes = big-endian uint16 original length).
func StripPadding(data []byte) []byte {
	if len(data) < 2 {
		return data
	}
	origLen := int(data[0])<<8 | int(data[1])
	if origLen < 0 || origLen > len(data)-2 {
		return data // not padded or corrupt — return as-is
	}
	return data[2 : 2+origLen]
}

// WrapWithLength prepends a 2-byte big-endian length so StripPadding can work.
func WrapWithLength(data []byte) []byte {
	if len(data) > 0xFFFF {
		return data
	}
	out := make([]byte, 2+len(data))
	out[0] = byte(len(data) >> 8)
	out[1] = byte(len(data))
	copy(out[2:], data)
	return out
}
