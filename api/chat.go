package api

type ChatUD struct {
	Name          string   `structs:"Name"`
	Msg           []string `structs:"Msg"`
	Type          int      `structs:"Type"`
	RawMsg        string   `structs:"RawMsg"`
	RawName       string   `structs:"RawName"`
	RawParameters []string `structs:"RawParameters"`
	Aux           any      `structs:"Aux"`
	ParsedMsg     string   `structs:"ParsedMsg"`
}

type ChatMsg struct {
	Msg           []string `structs:"msg"`
	Name          string   `structs:"name"`
	ParsedMsg     string   `structs:"parsed_msg"`
	RawMsg        string   `structs:"raw_msg"`
	RawParameters []string `structs:"raw_parameters"`
	Type          byte     `structs:"type"`
	UD            ChatUD   `structs:"ud"`
}
