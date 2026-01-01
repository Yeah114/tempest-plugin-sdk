package api

// NameChatModule is the module name used by Frame.GetModule / Frame.ListModules.
// It must match the host module name (tempest frame/module/common/name.go).
const NameChatModule = "chat"

type ChatModule interface {
	Name() string
	RegisterWhenChatMsg(handler func(event *ChatMsg)) (string, error)
	UnregisterWhenChatMsg(listenerID string) bool
	RegisterWhenReceiveMsgFromSenderNamed(name string, handler func(event *ChatMsg)) (string, error)
	UnregisterWhenReceiveMsgFromSenderNamed(listenerID string) bool
	InterceptNextMessage(name string, handler func(*ChatMsg)) (func(), error)
}
