// Code generated by "stringer -type=MessageType"; DO NOT EDIT.

package protocol

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[MessageTypePing-0]
	_ = x[MessageTypeLiveCmd-1]
	_ = x[MessageTypeDeviceData-2]
	_ = x[MessageTypePresetRecall-4]
}

const (
	_MessageType_name_0 = "MessageTypePingMessageTypeLiveCmdMessageTypeDeviceData"
	_MessageType_name_1 = "MessageTypePresetRecall"
)

var (
	_MessageType_index_0 = [...]uint8{0, 15, 33, 54}
)

func (i MessageType) String() string {
	switch {
	case i <= 2:
		return _MessageType_name_0[_MessageType_index_0[i]:_MessageType_index_0[i+1]]
	case i == 4:
		return _MessageType_name_1
	default:
		return "MessageType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
