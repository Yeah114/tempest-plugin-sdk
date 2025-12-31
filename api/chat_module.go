package api

type ChatModule interface {
	Name() string
	RegisterWhenChatMsg(handler func(event *ChatMsg)) (string, error)
	UnregisterWhenChatMsg(listenerID string)
	RegisterWhenReceiveMsgFromSenderNamed(name string, handler func(event *ChatMsg))
	UnregisterWhenReceiveMsgFromSenderNamed(listenerID string)
	InterceptNextMessage(name string, handler func(*ChatMsg))
}
