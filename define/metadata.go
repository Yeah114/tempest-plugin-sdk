package define

type Metadata struct {
	Name        string `json:"名称"`
	Description string `json:"描述"`
	Author      string `json:"作者,omitempty"`
}
