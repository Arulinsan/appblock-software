//go:build windows
// +build windows

package gui

import (
	"appblock/config"
	"appblock/utils"
	"fmt"
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/shirou/gopsutil/v3/process"
)

// PersonalityPreset represents a personality template
type PersonalityPreset struct {
	Name        string
	Description string
	Personality string
}

var personalityPresets = []PersonalityPreset{
	{
		Name:        "Programmer/Developer",
		Description: "Fokus coding & problem solving",
		Personality: "Tegas, to-the-point, fokus produktivitas, bahasa Indonesia. Ingatkan untuk fokus coding dan selesaikan task programming.",
	},
	{
		Name:        "Designer/Creative",
		Description: "Fokus kreativitas & desain",
		Personality: "Inspiratif, mendukung kreativitas, fokus design thinking, bahasa Indonesia. Dorong untuk berkarya dan eksplorasi ide.",
	},
	{
		Name:        "Student/Pelajar",
		Description: "Fokus belajar & mengerjakan tugas",
		Personality: "Lembut, suportif, tegas, fokus belajar, bahasa Indonesia. Ingatkan pentingnya belajar dan menyelesaikan tugas.",
	},
	{
		Name:        "Writer/Content Creator",
		Description: "Fokus menulis & konten",
		Personality: "Inspiratif, mendorong produktivitas menulis, bahasa Indonesia. Motivasi untuk terus berkarya dan publish konten.",
	},
	{
		Name:        "Entrepreneur/Business",
		Description: "Fokus bisnis & produktivitas",
		Personality: "Motivatif, fokus goals, action-oriented, bahasa Indonesia. Dorong untuk eksekusi rencana dan capai target bisnis.",
	},
	{
		Name:        "Custom",
		Description: "Buat personality sendiri",
		Personality: "",
	},
}

// ShowSettings shows the settings window
func ShowSettings(onSave func()) error {
	cfg := config.Get()
	
	var mainWindow *walk.MainWindow
	var blocklistBox *walk.ListBox
	var blocklistModel *BlocklistModel
	var timeWindowBox *walk.ListBox
	var timeWindowModel *TimeWindowModel
	var personalityCombo *walk.ComboBox
	var personalityEdit *walk.TextEdit
	var enabledCheck *walk.CheckBox
	var autostartCheck *walk.CheckBox
	var scanIntervalEdit *walk.NumberEdit
	var popupCooldownEdit *walk.NumberEdit
	var aiEnabledCheck *walk.CheckBox
	
	// Blocklist model
	blocklistModel = NewBlocklistModel(cfg.Blocklist)
	
	// TimeWindow model
	timeWindowModel = NewTimeWindowModel(cfg.TimeWindows)
	
	// Create and show the window
	err := MainWindow{
		AssignTo: &mainWindow,
		Title:    "APPBlock - Settings",
		MinSize:  Size{Width: 600, Height: 700},
		Size:     Size{Width: 650, Height: 750},
		Layout:   VBox{},
		ToolTipText: "",
		Children: []Widget{
			Composite{
				Layout: VBox{},
				Children: []Widget{
					// Header
					Label{
						Text: "⚙️ Pengaturan APPBlock",
						Font: Font{PointSize: 14, Bold: true},
					},
					Label{
						Text: "Atur aplikasi yang akan diblokir dan jam produktif Anda",
						Font: Font{PointSize: 9},
					},
					
					// General Settings
					GroupBox{
						Title:  "General",
						Layout: Grid{Columns: 2},
						ToolTipText: "",
						Children: []Widget{
							CheckBox{
								AssignTo: &enabledCheck,
								Text:     "Enable Blocking",
								Checked:  cfg.Enabled,
								ToolTipText: "",
							},
							CheckBox{
								AssignTo: &autostartCheck,
								Text:     "Start with Windows",
								Checked:  cfg.Autostart,
								ToolTipText: "",
							},
							Label{Text: "Scan Interval (detik):", ToolTipText: ""},
							NumberEdit{
								AssignTo: &scanIntervalEdit,
								Value:    float64(cfg.ScanIntervalSeconds),
								MinValue: 1,
								MaxValue: 60,
								ToolTipText: "",
							},
							Label{Text: "Popup Cooldown (detik):", ToolTipText: ""},
							NumberEdit{
								AssignTo: &popupCooldownEdit,
								Value:    float64(cfg.PopupCooldownSeconds),
								MinValue: 10,
								MaxValue: 300,
								ToolTipText: "",
							},
						},
					},
					
					// Blocklist
					GroupBox{
						Title:  "📋 Aplikasi yang Diblokir",
						Layout: VBox{},
						Children: []Widget{
							Composite{
								Layout: HBox{},
								Children: []Widget{
									PushButton{
										Text: "➕ Tambah dari Task Manager",
										OnClicked: func() {
											addFromTaskManager(blocklistModel, mainWindow)
										},
									},
									PushButton{
										Text: "📁 Browse File .exe",
										OnClicked: func() {
											addFromFileBrowser(blocklistModel, mainWindow)
										},
									},
									PushButton{
										Text: "✏️ Input Manual",
										OnClicked: func() {
											addManual(blocklistModel, mainWindow)
										},
									},
								},
							},
							ListBox{
								AssignTo: &blocklistBox,
								Model:    blocklistModel,
								MinSize:  Size{Height: 150},
							},
							Composite{
								Layout: HBox{},
								Children: []Widget{
									PushButton{
										Text: "🗑️ Hapus",
										OnClicked: func() {
											idx := blocklistBox.CurrentIndex()
											if idx >= 0 {
												blocklistModel.Remove(idx)
											}
										},
									},
									PushButton{
										Text: "🔄 Refresh dari Task Manager",
										OnClicked: func() {
											showRunningProcesses(mainWindow)
										},
									},
								},
							},
						},
					},
					
					// AI Personality
					GroupBox{
						Title:  "🤖 AI Personality",
						Layout: VBox{},
						Children: []Widget{						Label{
							Text: "ℹ️ Untuk AI: Buat file .env dengan GEMINI_API_KEY=your_key\nTanpa API key = pesan default.",
							Font: Font{PointSize: 8},
							TextColor: walk.RGB(100, 100, 100),
						},							CheckBox{
								AssignTo: &aiEnabledCheck,
								Text:     "Enable AI Motivational Messages",
								Checked:  cfg.AI.Enabled,
							},
							Label{Text: "Pilih Personality Preset:"},
							ComboBox{
								AssignTo: &personalityCombo,
								Model:    getPersonalityNames(),
								OnCurrentIndexChanged: func() {
									idx := personalityCombo.CurrentIndex()
									if idx >= 0 && idx < len(personalityPresets) {
										preset := personalityPresets[idx]
										personalityEdit.SetText(preset.Personality)
										if preset.Name == "Custom" {
											personalityEdit.SetReadOnly(false)
										} else {
											personalityEdit.SetReadOnly(true)
										}
									}
								},
							},
							Label{Text: "Personality Detail:"},
							TextEdit{
								AssignTo: &personalityEdit,
								Text:     cfg.AI.Personality,
								MinSize:  Size{Height: 80},
								ReadOnly: true,
							},
						},
					},
					
					// Time Windows
					GroupBox{
						Title:  "⏰ Jam Produktif",
						Layout: VBox{},
						ToolTipText: "",
						Children: []Widget{
							Label{
								Text: "Atur kapan blocking aktif (format 24 jam)",
								Font: Font{PointSize: 9},
							},
							Composite{
								Layout: HBox{},
								Children: []Widget{
									PushButton{
										Text: "➕ Tambah Window",
										OnClicked: func() {
											addTimeWindow(timeWindowModel, mainWindow)
										},
									},
									PushButton{
										Text: "✏️ Edit",
										OnClicked: func() {
											idx := timeWindowBox.CurrentIndex()
											if idx >= 0 {
												editTimeWindow(timeWindowModel, idx, mainWindow)
											}
										},
									},
									PushButton{
										Text: "🗑️ Hapus",
										OnClicked: func() {
											idx := timeWindowBox.CurrentIndex()
											if idx >= 0 {
												timeWindowModel.Remove(idx)
											}
										},
									},
									PushButton{
										Text: "📋 Presets",
										OnClicked: func() {
											showTimePresets(timeWindowModel, mainWindow)
										},
									},
								},
							},
							ListBox{
								AssignTo: &timeWindowBox,
								Model:    timeWindowModel,
								MinSize:  Size{Height: 100},
							},
						},
					},
				},
			},
			
			// Buttons
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{
						Text: "💾 Save",
						OnClicked: func() {
							// Save config
							cfg := config.Get()
							cfg.Enabled = enabledCheck.Checked()
							cfg.Autostart = autostartCheck.Checked()
							cfg.ScanIntervalSeconds = int(scanIntervalEdit.Value())
							cfg.PopupCooldownSeconds = int(popupCooldownEdit.Value())
							cfg.Blocklist = blocklistModel.GetItems()
							cfg.TimeWindows = timeWindowModel.GetItems()
							cfg.AI.Enabled = aiEnabledCheck.Checked()
							cfg.AI.Personality = personalityEdit.Text()
							
							if err := config.Save(); err != nil {
								walk.MsgBox(mainWindow, "Error", "Failed to save config: "+err.Error(), walk.MsgBoxIconError)
								return
							}
							
							// Mark first run as completed
							cfg.MarkFirstRunCompleted()
							
							utils.LogInfo("Settings saved via GUI")
							
							// Trigger reload callback
							if onSave != nil {
								onSave()
							}
							
							walk.MsgBox(mainWindow, "Success", "Settings saved successfully!\n\nClick Reload Config from tray menu to apply changes.", walk.MsgBoxIconInformation)
							mainWindow.Close()
						},
					},
					PushButton{
						Text: "❌ Cancel",
						OnClicked: func() {
							mainWindow.Close()
						},
					},
				},
			},
		},
	}.Create()
	
	if err != nil {
		return err
	}
	
	// Bring window to front and activate
	mainWindow.Show()
	mainWindow.BringToTop()
	mainWindow.Activate()
	
	// Run the message loop
	mainWindow.Run()
	
	return nil
}

// TimeWindowModel for ListBox
type TimeWindowModel struct {
	walk.ListModelBase
	items []config.TimeWindow
}

func NewTimeWindowModel(items []config.TimeWindow) *TimeWindowModel {
	m := &TimeWindowModel{items: make([]config.TimeWindow, len(items))}
	copy(m.items, items)
	return m
}

func (m *TimeWindowModel) ItemCount() int {
	return len(m.items)
}

func (m *TimeWindowModel) Value(index int) interface{} {
	tw := m.items[index]
	return fmt.Sprintf("%s - %s", tw.Start, tw.End)
}

func (m *TimeWindowModel) Add(item config.TimeWindow) {
	m.items = append(m.items, item)
	m.PublishItemsReset()
}

func (m *TimeWindowModel) Update(index int, item config.TimeWindow) {
	if index >= 0 && index < len(m.items) {
		m.items[index] = item
		m.PublishItemsReset()
	}
}

func (m *TimeWindowModel) Remove(index int) {
	m.items = append(m.items[:index], m.items[index+1:]...)
	m.PublishItemsReset()
}

func (m *TimeWindowModel) GetItems() []config.TimeWindow {
	return m.items
}

// BlocklistModel for ListBox
type BlocklistModel struct {
	walk.ListModelBase
	items []string
}

func NewBlocklistModel(items []string) *BlocklistModel {
	m := &BlocklistModel{items: make([]string, len(items))}
	copy(m.items, items)
	return m
}

func (m *BlocklistModel) ItemCount() int {
	return len(m.items)
}

func (m *BlocklistModel) Value(index int) interface{} {
	return m.items[index]
}

func (m *BlocklistModel) Add(item string) {
	// Check for duplicates
	for _, existing := range m.items {
		if strings.EqualFold(existing, item) {
			return
		}
	}
	m.items = append(m.items, item)
	m.PublishItemsReset()
}

func (m *BlocklistModel) Remove(index int) {
	m.items = append(m.items[:index], m.items[index+1:]...)
	m.PublishItemsReset()
}

func (m *BlocklistModel) GetItems() []string {
	return m.items
}

// Helper functions
func getPersonalityNames() []string {
	names := make([]string, len(personalityPresets))
	for i, preset := range personalityPresets {
		names[i] = preset.Name + " - " + preset.Description
	}
	return names
}

func addFromTaskManager(model *BlocklistModel, owner walk.Form) {
	processes, err := process.Processes()
	if err != nil {
		walk.MsgBox(owner, "Error", "Failed to get processes: "+err.Error(), walk.MsgBoxIconError)
		return
	}
	
	var processList []string
	processMap := make(map[string]bool)
	
	for _, proc := range processes {
		name, err := proc.Name()
		if err != nil {
			continue
		}
		
		// Filter out system processes
		lowerName := strings.ToLower(name)
		if strings.Contains(lowerName, "system") || 
		   strings.Contains(lowerName, "svchost") ||
		   strings.Contains(lowerName, "conhost") {
			continue
		}
		
		if !processMap[name] {
			processMap[name] = true
			processList = append(processList, name)
		}
	}
	
	if len(processList) == 0 {
		walk.MsgBox(owner, "Info", "No user processes found", walk.MsgBoxIconInformation)
		return
	}
	
	// Show selection dialog
	var dlg *walk.Dialog
	var listBox *walk.ListBox
	
	listModel := &BlocklistModel{items: processList}
	
	Dialog{
		AssignTo: &dlg,
		Title:    "Select Process",
		MinSize:  Size{Width: 400, Height: 500},
		Layout:   VBox{},
		Children: []Widget{
			Label{Text: "Pilih aplikasi yang ingin diblokir:"},
			ListBox{
				AssignTo:      &listBox,
				Model:         listModel,
				MinSize:       Size{Height: 400},
				MultiSelection: true,
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{
						Text: "Add Selected",
						OnClicked: func() {
							indices := listBox.SelectedIndexes()
							for _, idx := range indices {
								if idx < len(processList) {
									model.Add(processList[idx])
								}
							}
							dlg.Accept()
						},
					},
					PushButton{
						Text: "Cancel",
						OnClicked: func() {
							dlg.Cancel()
						},
					},
				},
			},
		},
	}.Run(owner)
}

func addFromFileBrowser(model *BlocklistModel, owner walk.Form) {
	dlg := walk.FileDialog{
		Title:  "Select Executable",
		Filter: "Executable Files (*.exe)|*.exe|All Files (*.*)|*.*",
	}
	
	if ok, _ := dlg.ShowOpen(owner); ok {
		// Extract filename from path
		parts := strings.Split(dlg.FilePath, "\\")
		if len(parts) > 0 {
			filename := parts[len(parts)-1]
			model.Add(filename)
		}
	}
}

func addManual(model *BlocklistModel, owner walk.Form) {
	var dlg *walk.Dialog
	var edit *walk.LineEdit
	
	Dialog{
		AssignTo: &dlg,
		Title:    "Add Process Name",
		MinSize:  Size{Width: 400, Height: 120},
		Layout:   VBox{},
		Children: []Widget{
			Label{Text: "Masukkan nama process (contoh: chrome.exe):"},
			LineEdit{
				AssignTo: &edit,
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{
						Text: "Add",
						OnClicked: func() {
							text := strings.TrimSpace(edit.Text())
							if text != "" {
								if !strings.HasSuffix(strings.ToLower(text), ".exe") {
									text += ".exe"
								}
								model.Add(text)
								dlg.Accept()
							}
						},
					},
					PushButton{
						Text: "Cancel",
						OnClicked: func() {
							dlg.Cancel()
						},
					},
				},
			},
		},
	}.Run(owner)
}

func showRunningProcesses(owner walk.Form) {
	processes, err := process.Processes()
	if err != nil {
		walk.MsgBox(owner, "Error", "Failed to get processes: "+err.Error(), walk.MsgBoxIconError)
		return
	}
	
	var info strings.Builder
	info.WriteString("Running User Processes:\n\n")
	
	processMap := make(map[string]bool)
	count := 0
	
	for _, proc := range processes {
		name, err := proc.Name()
		if err != nil {
			continue
		}
		
		lowerName := strings.ToLower(name)
		if strings.Contains(lowerName, "system") || 
		   strings.Contains(lowerName, "svchost") {
			continue
		}
		
		if !processMap[name] {
			processMap[name] = true
			info.WriteString(fmt.Sprintf("• %s\n", name))
			count++
		}
	}
	
	info.WriteString(fmt.Sprintf("\nTotal: %d processes", count))
	
	walk.MsgBox(owner, "Running Processes", info.String(), walk.MsgBoxIconInformation)
}

// Time Window Functions

func addTimeWindow(model *TimeWindowModel, owner walk.Form) {
	var dlg *walk.Dialog
	var startEdit, endEdit *walk.LineEdit
	
	Dialog{
		AssignTo: &dlg,
		Title:    "Tambah Time Window",
		MinSize:  Size{Width: 400, Height: 180},
		Layout:   VBox{},
		Children: []Widget{
			Label{Text: "Atur jam produktif (format 24 jam: HH:MM)"},
			Composite{
				Layout: Grid{Columns: 2},
				Children: []Widget{
					Label{Text: "Start Time:"},
					LineEdit{
						AssignTo: &startEdit,
						Text:     "09:00",
					},
					Label{Text: "End Time:"},
					LineEdit{
						AssignTo: &endEdit,
						Text:     "17:00",
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{
						Text: "Add",
						OnClicked: func() {
							start := strings.TrimSpace(startEdit.Text())
							end := strings.TrimSpace(endEdit.Text())
							
							if !isValidTime(start) || !isValidTime(end) {
								walk.MsgBox(dlg, "Error", "Invalid time format!\nUse HH:MM (24-hour format)", walk.MsgBoxIconError)
								return
							}
							
							model.Add(config.TimeWindow{
								Start: start,
								End:   end,
							})
							dlg.Accept()
						},
					},
					PushButton{
						Text: "Cancel",
						OnClicked: func() {
							dlg.Cancel()
						},
					},
				},
			},
		},
	}.Run(owner)
}

func editTimeWindow(model *TimeWindowModel, index int, owner walk.Form) {
	if index < 0 || index >= model.ItemCount() {
		return
	}
	
	tw := model.items[index]
	var dlg *walk.Dialog
	var startEdit, endEdit *walk.LineEdit
	
	Dialog{
		AssignTo: &dlg,
		Title:    "Edit Time Window",
		MinSize:  Size{Width: 400, Height: 180},
		Layout:   VBox{},
		Children: []Widget{
			Label{Text: "Edit jam produktif (format 24 jam: HH:MM)"},
			Composite{
				Layout: Grid{Columns: 2},
				Children: []Widget{
					Label{Text: "Start Time:"},
					LineEdit{
						AssignTo: &startEdit,
						Text:     tw.Start,
					},
					Label{Text: "End Time:"},
					LineEdit{
						AssignTo: &endEdit,
						Text:     tw.End,
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{
						Text: "Save",
						OnClicked: func() {
							start := strings.TrimSpace(startEdit.Text())
							end := strings.TrimSpace(endEdit.Text())
							
							if !isValidTime(start) || !isValidTime(end) {
								walk.MsgBox(dlg, "Error", "Invalid time format!\nUse HH:MM (24-hour format)", walk.MsgBoxIconError)
								return
							}
							
							model.Update(index, config.TimeWindow{
								Start: start,
								End:   end,
							})
							dlg.Accept()
						},
					},
					PushButton{
						Text: "Cancel",
						OnClicked: func() {
							dlg.Cancel()
						},
					},
				},
			},
		},
	}.Run(owner)
}

func showTimePresets(model *TimeWindowModel, owner walk.Form) {
	presets := map[string][]config.TimeWindow{
		"Full Day Worker (9-5)": {
			{Start: "09:00", End: "12:00"},
			{Start: "13:00", End: "17:00"},
		},
		"Student Schedule": {
			{Start: "06:00", End: "08:00"},
			{Start: "12:00", End: "14:00"},
			{Start: "19:00", End: "22:00"},
		},
		"Freelancer/Night Owl": {
			{Start: "14:00", End: "18:00"},
			{Start: "20:00", End: "23:30"},
		},
		"Early Bird": {
			{Start: "05:00", End: "08:00"},
			{Start: "09:00", End: "12:00"},
		},
		"Lunch & Evening": {
			{Start: "12:00", End: "14:00"},
			{Start: "19:00", End: "22:00"},
		},
	}
	
	var dlg *walk.Dialog
	var listBox *walk.ListBox
	
	presetNames := []string{
		"Full Day Worker (9-5)",
		"Student Schedule",
		"Freelancer/Night Owl",
		"Early Bird",
		"Lunch & Evening",
	}
	
	listModel := &BlocklistModel{items: presetNames}
	
	Dialog{
		AssignTo: &dlg,
		Title:    "Time Window Presets",
		MinSize:  Size{Width: 400, Height: 400},
		Layout:   VBox{},
		Children: []Widget{
			Label{Text: "Pilih preset atau buat custom:"},
			ListBox{
				AssignTo: &listBox,
				Model:    listModel,
				MinSize:  Size{Height: 250},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{
						Text: "Apply Preset",
						OnClicked: func() {
							idx := listBox.CurrentIndex()
							if idx >= 0 && idx < len(presetNames) {
								// Clear existing
								for model.ItemCount() > 0 {
									model.Remove(0)
								}
								
								// Add preset windows
								for _, tw := range presets[presetNames[idx]] {
									model.Add(tw)
								}
								
								walk.MsgBox(dlg, "Success", "Preset applied!\nDon't forget to Save settings.", walk.MsgBoxIconInformation)
								dlg.Accept()
							}
						},
					},
					PushButton{
						Text: "Cancel",
						OnClicked: func() {
							dlg.Cancel()
						},
					},
				},
			},
		},
	}.Run(owner)
}

func isValidTime(timeStr string) bool {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return false
	}
	
	var hour, minute int
	_, err := fmt.Sscanf(timeStr, "%d:%d", &hour, &minute)
	if err != nil {
		return false
	}
	
	return hour >= 0 && hour < 24 && minute >= 0 && minute < 60
}
