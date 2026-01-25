package api

import "github.com/Yeah114/EmptyDea-plugin-sdk/define"

type PluginTool struct {
	frame          define.Frame
	pluginID       string
	config         define.PluginConfig
	terminalModule TerminalModule
	loggerModule   LoggerModule
}

func NewPluginTool(plugin Plugin) PluginTool {
	t := PluginTool{}
	frame := plugin.Frame()
	id := plugin.ID()
	t.frame = frame
	t.pluginID = id
	t.config, _ = frame.GetPluginConfig(id)
	t.terminalModule, _ = GetModule[TerminalModule](frame, NameTerminalModule)
	t.loggerModule, _ = GetModule[LoggerModule](frame, NameLoggerModule)
	return t
}

func (t *PluginTool) Name() string {
	return t.config.Name
}

func (t *PluginTool) Print(level Level, msg string) {
	t.terminalModule.Print(level, t.Name(), msg)
}

func (t *PluginTool) ColorTransANSI(msg string) string {
	return t.terminalModule.ColorTransANSI(msg)
}

func (t *PluginTool) Info(msg string) {
	t.terminalModule.Info(t.Name(), msg)
}

func (t *PluginTool) Warn(msg string) {
	t.terminalModule.Warn(t.Name(), msg)
}

func (t *PluginTool) Error(msg string) {
	t.terminalModule.Error(t.Name(), msg)
}

func (t *PluginTool) Success(msg string) {
	t.terminalModule.Success(t.Name(), msg)
}

func (t *PluginTool) Log(level Level, msg string) {
	t.loggerModule.Log(t.Name(), level, msg)
}

func (t *PluginTool) LogInfo(msg string) {
	t.loggerModule.Info(t.Name(), msg)
}

func (t *PluginTool) LogWarn(msg string) {
	t.loggerModule.Warn(t.Name(), msg)
}

func (t *PluginTool) LogError(msg string) {
	t.loggerModule.Error(t.Name(), msg)
}

func (t *PluginTool) LogSuccess(msg string) {
	t.loggerModule.Success(t.Name(), msg)
}

func (t *PluginTool) UpgradePluginConfig(config map[string]interface{}) error {
	if t == nil || t.frame == nil || t.pluginID == "" {
		return nil
	}
	return t.frame.UpgradePluginConfig(t.pluginID, config)
}

func (t *PluginTool) UpgradePluginFullConfig(config define.PluginConfig) error {
	if t == nil || t.frame == nil || t.pluginID == "" {
		return nil
	}
	return t.frame.UpgradePluginFullConfig(t.pluginID, config)
}

type PluginModuleTool struct {
	ChatModule
	CommandsModule
	UQHolderModule
	PlayersModule
	TerminalMenuModule
	GameMenuModule
	FlexModule
	StoragePathModule
}

func NewPluginModuleTool(frame define.Frame) PluginModuleTool {
	t := PluginModuleTool{}
	t.ChatModule, _ = GetModule[ChatModule](frame, NameChatModule)
	t.CommandsModule, _ = GetModule[CommandsModule](frame, NameCommandsModule)
	t.UQHolderModule, _ = GetModule[UQHolderModule](frame, NameUQHolderModule)
	t.PlayersModule, _ = GetModule[PlayersModule](frame, NamePlayersModule)
	t.TerminalMenuModule, _ = GetModule[TerminalMenuModule](frame, NameTerminalMenuModule)
	t.GameMenuModule, _ = GetModule[GameMenuModule](frame, NameGameMenuModule)
	t.FlexModule, _ = GetModule[FlexModule](frame, NameFlexModule)
	t.StoragePathModule, _ = GetModule[StoragePathModule](frame, NameStoragePathModule)
	return t
}
