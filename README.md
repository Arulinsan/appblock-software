# APPBlock - Productivity Blocker

Aplikasi Windows yang auto-blokir aplikasi pengganggu (games, sosmed) selama jam produktif. Dilengkapi AI motivator dari Gemini.

---

## Quick Start

### 1. Setup API Key

Dapatkan API key gratis:

```
https://aistudio.google.com/app/apikey
```

Buat file `.env`:

```
GEMINI_API_KEY=AIzaSy_Your_API_Key_Here
```

### 2. Jalankan

```bash
appblock.exe
```

Icon muncul di **system tray** (pojok kanan bawah).

### 3. Buka Settings

```
Right-click icon → Settings
```

### 4. Setup Blocklist

```
Settings → Blocklist
→ Tambah dari Task Manager
→ Pilih apps (chrome, discord, games)
→ Add Selected
```

### 5. Atur Jam Produktif

```
Settings → Jam Produktif
→ Presets → "Student Schedule"
```

### 6. Pilih AI Personality

```
Settings → AI Personality
→ Pilih: Programmer/Student/Designer/Writer/Entrepreneur
```

### 7. Save!

```
Save → Right-click tray → Reload Config
```

**Done!**

---

## Penggunaan Sehari-hari

**Pagi:** APPBlock auto-start, check tray icon

**Saat Jam Produktif:** Buka blocked app → Tertutup otomatis → Popup AI muncul

**Butuh Break:** `Right-click tray → Disable Blocking`

**Update Settings:** `Settings → Edit → Save → Reload`

---

## Fitur

- **GUI Settings** - No manual config
- **Task Manager Integration** - Pilih apps live
- **Time Windows** - Blokir di jam tertentu
- **AI Motivator** - Pesan dari Gemini
- **Hot Reload** - Update tanpa restart
- **Autostart** - Jalan saat boot

---

## Troubleshooting

**"GEMINI_API_KEY not set"**

```bash
echo GEMINI_API_KEY=AIzaSy_Your_Key > .env
```

**Apps tidak diblokir?**

- Pastikan dalam jam produktif
- Verify process name di Task Manager
- Status harus "Active" di tray

**Settings tidak bisa dibuka?**

```bash
.\build.bat
```

**Icon tidak muncul?**

```bash
.\refresh_icon.bat
```

---

## Development

### Project Structure

```
APPBlock/
├── main.go              # Entry point
├── config/              # Config loader & hot reload
├── scheduler/           # Time windows logic
├── blocker/             # Process monitoring & killer
├── gemini/              # AI client (Gemini API)
├── popup/               # Windows notification
├── tray/                # System tray menu
├── gui/                 # Settings GUI (lxn/walk)
├── autostart/           # Registry manager
└── utils/               # Logger
```

### Tech Stack

- **Go 1.25+** - Main language
- **lxn/walk** - Windows native GUI
- **gopsutil** - Process management
- **Gemini AI** - AI motivator

### Build dari Source

```bash
# Install dependencies
go mod download

# Quick build
.\build.bat

# Manual build
rsrc -manifest appblock.manifest -ico icon.ico -o rsrc.syso
go build -ldflags "-s -w -H windowsgui" -o appblock.exe
```

### Key Components

**Config System (`config/config.go`):**

- JSON-based configuration
- Hot reload via file watcher
- Settings: blocklist, time windows, AI personality

**Blocker (`blocker/blocker.go`):**

- Scan interval: 2 seconds
- Process enumeration via gopsutil
- Kill matching processes instantly

**Scheduler (`scheduler/scheduler.go`):**

- Check current time vs time windows
- Return blocking status (enabled/disabled)

**GUI (`gui/settings.go`):**

- Task Manager integration for process picker
- Time windows visual editor
- AI personality presets

**Gemini Client (`gemini/client.go`):**

- Uses `GEMINI_API_KEY` from .env
- Sends context-aware prompts
- Handles API errors gracefully

### Adding Features

**Add new AI personality:**

```go
// In gui/settings.go
personalities["New Role"] = "Your prompt here..."
```

**Modify scan interval:**

```go
// In blocker/blocker.go
ticker := time.NewTicker(2 * time.Second) // Change duration
```

**Custom popup message:**

```go
// In popup/popup.go
// Modify ShowBlockPopup() function
```

### Debug Mode

```bash
# Run with console output
go run main.go

# Check logs
type app.log

# Monitor real-time
Get-Content app.log -Wait
```

---

## System Tray

```
APPBlock
├── Status: Active
├── Enable/Disable Blocking
├── Settings
├── Reload Config
├── Enable Autostart
└── Quit
```

---

## Tips

- **Start Simple** - Mulai 2-3 jam per hari
- **Right Personality** - Pilih AI yang cocok
- **Review Logs** - Check `app.log`
- **Use Presets** - Quick setup

---

## Credits

**lxn/walk** • **Google Gemini AI** • **gopsutil**

---

## License

Personal use project. Open source untuk learning & modification.

---

**Version 2.0.0** • Made with love for productivity
