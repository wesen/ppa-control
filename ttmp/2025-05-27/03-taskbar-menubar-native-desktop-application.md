# Native Desktop Taskbar/Menubar Application Design for DSP Speaker Control

## 1. Purpose

Design a native desktop application for macOS and Windows that provides seamless control of DSP speaker systems through a taskbar/menubar interface. The application will use Wails v2 to combine Go backend capabilities with a modern web frontend, offering quick access to volume control, preset management, and device monitoring without requiring a full window interface.

## 2. System Architecture Overview

### Technology Stack
- **Backend**: Go (reusing existing PPA control system codebase)
- **Frontend**: React/TypeScript with modern UI components
- **Framework**: Wails v2 for native desktop integration
- **Communication**: Direct UDP to PPA devices (no server required)
- **UI Framework**: Tailwind CSS + Headless UI for modern, accessible components

### Architecture Diagram
```
┌─────────────────────────────────────────────────────────────┐
│                    Native Desktop App                       │
├─────────────────────────────────────────────────────────────┤
│  Taskbar/Menubar Interface                                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │   Volume    │  │   Presets   │  │   Devices   │        │
│  │   Control   │  │   Manager   │  │   Status    │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
├─────────────────────────────────────────────────────────────┤
│  Wails Runtime Bridge                                       │
├─────────────────────────────────────────────────────────────┤
│  Go Backend (PPA Control System)                           │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │ MultiClient │  │  Discovery  │  │  Protocol   │        │
│  │   Manager   │  │   Service   │  │   Handler   │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
├─────────────────────────────────────────────────────────────┤
│  UDP Network Layer                                         │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
                    ┌─────────────────┐
                    │   PPA Devices   │
                    │   (Speakers)    │
                    └─────────────────┘
```

## 3. User Experience Design

### 3.1. Menubar/Taskbar Icon States
- **Disconnected**: Gray speaker icon with red dot
- **Connected**: Blue speaker icon with device count badge
- **Active Control**: Green speaker icon with volume level indicator
- **Error State**: Orange speaker icon with warning indicator

### 3.2. Quick Access Menu (Primary Interface)
When clicking the menubar icon, show a compact dropdown menu:

```
┌─────────────────────────────────────┐
│ 🔊 PPA Speaker Control              │
├─────────────────────────────────────┤
│ 📡 Devices (2 connected)            │
│   • Living Room Speaker    [●]      │
│   • Kitchen Speaker        [●]      │
├─────────────────────────────────────┤
│ 🔊 Master Volume           [▓▓▓░░]  │
│    ├─ Living Room         [▓▓▓▓░]  │
│    └─ Kitchen             [▓▓░░░]  │
├─────────────────────────────────────┤
│ 🎵 Quick Presets                    │
│   [1] Normal  [2] Party  [3] Night  │
│   [4] Cinema  [5] Music  [6] Voice  │
├─────────────────────────────────────┤
│ ⚙️  Settings                        │
│ 🔍 Discover Devices                 │
│ ❌ Quit                             │
└─────────────────────────────────────┘
```

### 3.3. Detailed Control Window (Secondary Interface)
For advanced controls, open a dedicated window:

#### Main Control Panel
- **Device Grid**: Visual representation of all discovered devices
- **Individual Controls**: Per-device volume, mute, preset selection
- **Group Controls**: Select multiple devices for synchronized control
- **Real-time Feedback**: Live status updates and visual indicators

#### Advanced Features Panel
- **EQ Controls**: Graphical equalizer for each device
- **Preset Management**: Create, edit, and organize custom presets
- **Network Settings**: Interface selection, discovery configuration
- **Logging**: Real-time command/response monitoring

## 4. Core Features & Implementation

### 4.1. Device Discovery & Management

#### Auto-Discovery Service
```go
type DiscoveryService struct {
    devices     map[string]*DeviceInfo
    subscribers []chan DeviceEvent
    interval    time.Duration
    ctx         context.Context
}

type DeviceInfo struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Address     string    `json:"address"`
    Status      string    `json:"status"` // "online", "offline", "connecting"
    LastSeen    time.Time `json:"lastSeen"`
    Volume      float32   `json:"volume"`
    CurrentPreset int     `json:"currentPreset"`
    Capabilities []string `json:"capabilities"`
}

type DeviceEvent struct {
    Type   string      `json:"type"` // "discovered", "lost", "updated"
    Device *DeviceInfo `json:"device"`
}
```

#### Wails Backend Methods
```go
// App struct for Wails
type App struct {
    ctx           context.Context
    cmdCtx        *lib.CommandContext
    discovery     *DiscoveryService
    deviceManager *DeviceManager
}

// Wails exported methods
func (a *App) GetDevices() []*DeviceInfo {
    return a.discovery.GetAllDevices()
}

func (a *App) ConnectToDevice(address string) error {
    return a.deviceManager.Connect(address)
}

func (a *App) SetMasterVolume(volume float32) error {
    return a.deviceManager.SetMasterVolume(volume)
}

func (a *App) SetDeviceVolume(deviceID string, volume float32) error {
    return a.deviceManager.SetDeviceVolume(deviceID, volume)
}

func (a *App) RecallPreset(deviceID string, presetIndex int) error {
    return a.deviceManager.RecallPreset(deviceID, presetIndex)
}

func (a *App) StartDiscovery() error {
    return a.discovery.Start()
}

func (a *App) StopDiscovery() error {
    return a.discovery.Stop()
}
```

### 4.2. Volume Control Implementation

#### Real-time Volume Control
```typescript
// Frontend React component
interface VolumeControlProps {
  deviceId?: string; // undefined for master volume
  currentVolume: number;
  onVolumeChange: (volume: number) => void;
}

const VolumeControl: React.FC<VolumeControlProps> = ({
  deviceId,
  currentVolume,
  onVolumeChange
}) => {
  const [localVolume, setLocalVolume] = useState(currentVolume);
  const [isAdjusting, setIsAdjusting] = useState(false);

  const handleVolumeChange = useCallback(
    debounce(async (volume: number) => {
      try {
        if (deviceId) {
          await SetDeviceVolume(deviceId, volume / 100);
        } else {
          await SetMasterVolume(volume / 100);
        }
        setIsAdjusting(false);
      } catch (error) {
        console.error('Failed to set volume:', error);
        // Revert to previous volume
        setLocalVolume(currentVolume);
        setIsAdjusting(false);
      }
    }, 150),
    [deviceId, currentVolume]
  );

  return (
    <div className="flex items-center space-x-2">
      <VolumeIcon className="w-4 h-4" />
      <input
        type="range"
        min="0"
        max="100"
        value={localVolume}
        onChange={(e) => {
          const newVolume = parseInt(e.target.value);
          setLocalVolume(newVolume);
          setIsAdjusting(true);
          handleVolumeChange(newVolume);
        }}
        className="flex-1 h-2 bg-gray-200 rounded-lg appearance-none cursor-pointer"
      />
      <span className="text-sm font-mono w-8">
        {Math.round(localVolume)}
      </span>
    </div>
  );
};
```

#### Backend Volume Management
```go
type DeviceManager struct {
    clients    map[string]*client.SingleDevice
    multiClient *client.MultiClient
    mu         sync.RWMutex
}

func (dm *DeviceManager) SetMasterVolume(volume float32) error {
    dm.mu.RLock()
    defer dm.mu.RUnlock()
    
    if dm.multiClient == nil {
        return errors.New("no devices connected")
    }
    
    // Send to all connected devices
    dm.multiClient.SendMasterVolume(volume)
    
    // Update local state
    for _, device := range dm.clients {
        // Update device volume state
    }
    
    return nil
}

func (dm *DeviceManager) SetDeviceVolume(deviceID string, volume float32) error {
    dm.mu.RLock()
    device, exists := dm.clients[deviceID]
    dm.mu.RUnlock()
    
    if !exists {
        return fmt.Errorf("device %s not found", deviceID)
    }
    
    return device.SendMasterVolume(volume)
}
```

### 4.3. Preset Management

#### Preset System
```go
type PresetManager struct {
    presets map[int]*Preset
    custom  map[string]*CustomPreset
}

type Preset struct {
    Index       int     `json:"index"`
    Name        string  `json:"name"`
    Description string  `json:"description"`
    Icon        string  `json:"icon"`
}

type CustomPreset struct {
    ID          string             `json:"id"`
    Name        string             `json:"name"`
    DeviceSettings map[string]*DeviceSettings `json:"deviceSettings"`
    CreatedAt   time.Time          `json:"createdAt"`
}

type DeviceSettings struct {
    Volume    float32 `json:"volume"`
    Muted     bool    `json:"muted"`
    EQSettings map[string]float32 `json:"eqSettings"`
}
```

#### Frontend Preset Interface
```typescript
const PresetGrid: React.FC = () => {
  const [presets] = usePresets();
  const [selectedDevices] = useSelectedDevices();

  const handlePresetRecall = async (presetIndex: number) => {
    try {
      if (selectedDevices.length === 0) {
        // Apply to all devices
        await RecallPresetAll(presetIndex);
      } else {
        // Apply to selected devices only
        await Promise.all(
          selectedDevices.map(deviceId => 
            RecallPreset(deviceId, presetIndex)
          )
        );
      }
      
      showNotification(`Preset ${presetIndex} applied`, 'success');
    } catch (error) {
      showNotification(`Failed to apply preset: ${error}`, 'error');
    }
  };

  return (
    <div className="grid grid-cols-3 gap-2 p-4">
      {presets.map((preset) => (
        <button
          key={preset.index}
          onClick={() => handlePresetRecall(preset.index)}
          className="flex flex-col items-center p-3 bg-white rounded-lg shadow hover:shadow-md transition-shadow"
        >
          <span className="text-2xl mb-1">{preset.icon}</span>
          <span className="text-sm font-medium">{preset.name}</span>
        </button>
      ))}
    </div>
  );
};
```

### 4.4. System Integration

#### macOS Menubar Integration
```go
// Wails app configuration
func main() {
    app := NewApp()

    err := wails.Run(&options.App{
        Title:            "PPA Speaker Control",
        Width:            400,
        Height:           600,
        MinWidth:         350,
        MinHeight:        500,
        DisableResize:    false,
        Fullscreen:       false,
        Frameless:        false,
        StartHidden:      true, // Start in system tray
        HideWindowOnClose: true, // Don't quit on window close
        RGBA:             &options.RGBA{R: 27, G: 38, B: 54, A: 1},
        Menu:             app.applicationMenu(),
        Logger:           nil,
        OnStartup:        app.startup,
        OnDomReady:       app.domReady,
        OnBeforeClose:    app.beforeClose,
        OnShutdown:       app.shutdown,
        WindowStartState: options.Minimised,
        Assets:           assets,
        BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
        OnSecondInstanceLaunch: app.onSecondInstanceLaunch,
    })

    if err != nil {
        println("Error:", err.Error())
    }
}

// System tray menu
func (a *App) applicationMenu() *menu.Menu {
    AppMenu := menu.NewMenu()
    
    // Quick volume control
    volumeSubmenu := AppMenu.AddSubmenu("Volume Control")
    volumeSubmenu.AddText("Master Volume", keys.CmdOrCtrl("m"), func(cd *menu.CallbackData) {
        a.ShowVolumeControl()
    })
    
    // Device management
    devicesSubmenu := AppMenu.AddSubmenu("Devices")
    devicesSubmenu.AddText("Discover Devices", keys.CmdOrCtrl("d"), func(cd *menu.CallbackData) {
        a.StartDiscovery()
    })
    
    // Presets
    presetsSubmenu := AppMenu.AddSubmenu("Quick Presets")
    for i := 1; i <= 6; i++ {
        presetIndex := i
        presetsSubmenu.AddText(fmt.Sprintf("Preset %d", i), nil, func(cd *menu.CallbackData) {
            a.RecallPresetAll(presetIndex)
        })
    }
    
    AppMenu.AddSeparator()
    AppMenu.AddText("Show Main Window", keys.CmdOrCtrl("w"), func(cd *menu.CallbackData) {
        runtime.WindowShow(a.ctx)
    })
    
    AppMenu.AddText("Quit", keys.CmdOrCtrl("q"), func(cd *menu.CallbackData) {
        runtime.Quit(a.ctx)
    })

    return AppMenu
}
```

#### Windows System Tray Integration
```go
// Windows-specific system tray handling
func (a *App) setupSystemTray() {
    // Create system tray icon
    icon := systray.NewIcon()
    icon.SetIcon(iconData) // Embedded icon data
    icon.SetTooltip("PPA Speaker Control")
    
    // Add menu items
    mVolumeUp := icon.AddMenuItem("Volume Up", "Increase master volume")
    mVolumeDown := icon.AddMenuItem("Volume Down", "Decrease master volume")
    icon.AddSeparator()
    
    mPreset1 := icon.AddMenuItem("Preset 1", "Apply preset 1 to all devices")
    mPreset2 := icon.AddMenuItem("Preset 2", "Apply preset 2 to all devices")
    icon.AddSeparator()
    
    mShow := icon.AddMenuItem("Show Window", "Show main control window")
    mQuit := icon.AddMenuItem("Quit", "Exit application")
    
    // Handle menu clicks
    go func() {
        for {
            select {
            case <-mVolumeUp.ClickedCh:
                a.AdjustMasterVolume(0.1) // +10%
            case <-mVolumeDown.ClickedCh:
                a.AdjustMasterVolume(-0.1) // -10%
            case <-mPreset1.ClickedCh:
                a.RecallPresetAll(1)
            case <-mPreset2.ClickedCh:
                a.RecallPresetAll(2)
            case <-mShow.ClickedCh:
                runtime.WindowShow(a.ctx)
            case <-mQuit.ClickedCh:
                runtime.Quit(a.ctx)
            }
        }
    }()
}
```

## 5. User Interface Design

### 5.1. Menubar Dropdown (Primary Interface)

#### Compact Device List
```typescript
const DeviceList: React.FC = () => {
  const [devices] = useDevices();
  
  return (
    <div className="space-y-1">
      <div className="flex items-center justify-between px-3 py-2 text-sm font-medium text-gray-700">
        <span>Devices ({devices.filter(d => d.status === 'online').length} connected)</span>
        <button onClick={startDiscovery} className="text-blue-600 hover:text-blue-800">
          <RefreshIcon className="w-4 h-4" />
        </button>
      </div>
      
      {devices.map(device => (
        <div key={device.id} className="flex items-center justify-between px-3 py-2 hover:bg-gray-50">
          <div className="flex items-center space-x-2">
            <StatusIndicator status={device.status} />
            <span className="text-sm">{device.name}</span>
          </div>
          <VolumeIndicator volume={device.volume} />
        </div>
      ))}
    </div>
  );
};
```

#### Quick Volume Control
```typescript
const QuickVolumeControl: React.FC = () => {
  const [masterVolume, setMasterVolume] = useMasterVolume();
  
  return (
    <div className="px-3 py-2 border-t border-gray-200">
      <div className="flex items-center space-x-2 mb-2">
        <VolumeIcon className="w-4 h-4 text-gray-600" />
        <span className="text-sm font-medium">Master Volume</span>
        <span className="text-xs text-gray-500 ml-auto">{Math.round(masterVolume)}%</span>
      </div>
      
      <input
        type="range"
        min="0"
        max="100"
        value={masterVolume}
        onChange={(e) => setMasterVolume(parseInt(e.target.value))}
        className="w-full h-2 bg-gray-200 rounded-lg appearance-none cursor-pointer"
      />
      
      <div className="flex justify-between mt-1">
        <button 
          onClick={() => setMasterVolume(0)}
          className="text-xs text-gray-500 hover:text-gray-700"
        >
          Mute
        </button>
        <button 
          onClick={() => setMasterVolume(100)}
          className="text-xs text-gray-500 hover:text-gray-700"
        >
          Max
        </button>
      </div>
    </div>
  );
};
```

### 5.2. Main Control Window (Secondary Interface)

#### Device Grid Layout
```typescript
const DeviceGrid: React.FC = () => {
  const [devices] = useDevices();
  const [selectedDevices, setSelectedDevices] = useSelectedDevices();
  
  return (
    <div className="grid grid-cols-2 lg:grid-cols-3 gap-4 p-6">
      {devices.map(device => (
        <DeviceCard
          key={device.id}
          device={device}
          selected={selectedDevices.includes(device.id)}
          onSelect={(selected) => {
            if (selected) {
              setSelectedDevices([...selectedDevices, device.id]);
            } else {
              setSelectedDevices(selectedDevices.filter(id => id !== device.id));
            }
          }}
        />
      ))}
    </div>
  );
};

const DeviceCard: React.FC<DeviceCardProps> = ({ device, selected, onSelect }) => {
  return (
    <div className={`
      bg-white rounded-lg shadow-md p-4 transition-all duration-200
      ${selected ? 'ring-2 ring-blue-500 bg-blue-50' : 'hover:shadow-lg'}
    `}>
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center space-x-2">
          <input
            type="checkbox"
            checked={selected}
            onChange={(e) => onSelect(e.target.checked)}
            className="rounded border-gray-300"
          />
          <h3 className="font-medium text-gray-900">{device.name}</h3>
        </div>
        <StatusBadge status={device.status} />
      </div>
      
      <VolumeControl
        deviceId={device.id}
        currentVolume={device.volume * 100}
        onVolumeChange={(volume) => SetDeviceVolume(device.id, volume / 100)}
      />
      
      <div className="mt-3 flex justify-between items-center">
        <span className="text-sm text-gray-500">
          Preset: {device.currentPreset || 'None'}
        </span>
        <button
          onClick={() => toggleMute(device.id)}
          className={`
            px-2 py-1 rounded text-xs font-medium
            ${device.muted 
              ? 'bg-red-100 text-red-800' 
              : 'bg-gray-100 text-gray-800'
            }
          `}
        >
          {device.muted ? 'Unmute' : 'Mute'}
        </button>
      </div>
    </div>
  );
};
```

## 6. Advanced Features

### 6.1. Keyboard Shortcuts & Global Hotkeys

#### System-wide Hotkeys
```go
// Global hotkey registration
func (a *App) registerGlobalHotkeys() error {
    // Volume up/down
    hotkey.Register(hotkey.KeyF1, hotkey.ModCmd, func() {
        a.AdjustMasterVolume(0.1)
    })
    
    hotkey.Register(hotkey.KeyF2, hotkey.ModCmd, func() {
        a.AdjustMasterVolume(-0.1)
    })
    
    // Quick preset access
    for i := 1; i <= 9; i++ {
        presetIndex := i
        hotkey.Register(hotkey.Key(hotkey.Key1 + i - 1), hotkey.ModCmd|hotkey.ModShift, func() {
            a.RecallPresetAll(presetIndex)
        })
    }
    
    // Show/hide main window
    hotkey.Register(hotkey.KeySpace, hotkey.ModCmd|hotkey.ModAlt, func() {
        runtime.WindowToggle(a.ctx)
    })
    
    return nil
}
```

### 6.2. Notifications & Feedback

#### Native Notifications
```go
func (a *App) showNotification(title, message string, notificationType string) {
    notification := &notification.Notification{
        Title:   title,
        Message: message,
        Icon:    a.getNotificationIcon(notificationType),
        Sound:   notification.SoundDefault,
    }
    
    notification.Show()
}

func (a *App) notifyVolumeChange(deviceName string, volume float32) {
    a.showNotification(
        "Volume Changed",
        fmt.Sprintf("%s: %d%%", deviceName, int(volume*100)),
        "volume",
    )
}

func (a *App) notifyPresetRecall(presetIndex int, deviceCount int) {
    a.showNotification(
        "Preset Applied",
        fmt.Sprintf("Preset %d applied to %d device(s)", presetIndex, deviceCount),
        "preset",
    )
}
```

### 6.3. Settings & Configuration

#### Settings Management
```go
type AppSettings struct {
    AutoDiscovery     bool     `json:"autoDiscovery"`
    DiscoveryInterval int      `json:"discoveryInterval"` // seconds
    DefaultPort       int      `json:"defaultPort"`
    Interfaces        []string `json:"interfaces"`
    ShowNotifications bool     `json:"showNotifications"`
    StartMinimized    bool     `json:"startMinimized"`
    GlobalHotkeys     bool     `json:"globalHotkeys"`
    Theme             string   `json:"theme"` // "light", "dark", "auto"
}

func (a *App) LoadSettings() (*AppSettings, error) {
    configPath := a.getConfigPath()
    data, err := os.ReadFile(configPath)
    if err != nil {
        return a.getDefaultSettings(), nil // Return defaults if no config
    }
    
    var settings AppSettings
    err = json.Unmarshal(data, &settings)
    return &settings, err
}

func (a *App) SaveSettings(settings *AppSettings) error {
    configPath := a.getConfigPath()
    data, err := json.MarshalIndent(settings, "", "  ")
    if err != nil {
        return err
    }
    
    return os.WriteFile(configPath, data, 0644)
}
```

#### Settings UI
```typescript
const SettingsPanel: React.FC = () => {
  const [settings, setSettings] = useSettings();
  
  return (
    <div className="space-y-6 p-6">
      <div>
        <h3 className="text-lg font-medium mb-4">Discovery Settings</h3>
        <div className="space-y-3">
          <label className="flex items-center">
            <input
              type="checkbox"
              checked={settings.autoDiscovery}
              onChange={(e) => updateSetting('autoDiscovery', e.target.checked)}
              className="rounded border-gray-300"
            />
            <span className="ml-2">Enable automatic device discovery</span>
          </label>
          
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Discovery Interval (seconds)
            </label>
            <input
              type="number"
              min="5"
              max="60"
              value={settings.discoveryInterval}
              onChange={(e) => updateSetting('discoveryInterval', parseInt(e.target.value))}
              className="w-20 px-3 py-1 border border-gray-300 rounded-md"
            />
          </div>
        </div>
      </div>
      
      <div>
        <h3 className="text-lg font-medium mb-4">Interface Settings</h3>
        <div className="space-y-3">
          <label className="flex items-center">
            <input
              type="checkbox"
              checked={settings.showNotifications}
              onChange={(e) => updateSetting('showNotifications', e.target.checked)}
              className="rounded border-gray-300"
            />
            <span className="ml-2">Show desktop notifications</span>
          </label>
          
          <label className="flex items-center">
            <input
              type="checkbox"
              checked={settings.globalHotkeys}
              onChange={(e) => updateSetting('globalHotkeys', e.target.checked)}
              className="rounded border-gray-300"
            />
            <span className="ml-2">Enable global keyboard shortcuts</span>
          </label>
        </div>
      </div>
    </div>
  );
};
```

## 7. Implementation Phases

### Phase 1: Core Infrastructure (Week 1-2)
- [ ] Set up Wails v2 project structure
- [ ] Integrate existing PPA control system Go code
- [ ] Implement basic device discovery service
- [ ] Create fundamental Wails backend methods
- [ ] Set up React frontend with TypeScript

### Phase 2: Basic Menubar Interface (Week 3-4)
- [ ] Implement system tray/menubar integration
- [ ] Create compact device list component
- [ ] Add basic volume control functionality
- [ ] Implement preset recall buttons
- [ ] Add device connection/disconnection

### Phase 3: Advanced Controls (Week 5-6)
- [ ] Build detailed control window
- [ ] Implement device grid layout
- [ ] Add individual device controls
- [ ] Create preset management system
- [ ] Implement multi-device selection

### Phase 4: System Integration (Week 7-8)
- [ ] Add global keyboard shortcuts
- [ ] Implement native notifications
- [ ] Create settings management
- [ ] Add auto-start functionality
- [ ] Implement proper error handling

### Phase 5: Polish & Testing (Week 9-10)
- [ ] UI/UX refinements and animations
- [ ] Performance optimization
- [ ] Comprehensive testing with real devices
- [ ] Documentation and user guides
- [ ] Prepare for distribution

## 8. Technical Challenges & Solutions

### 8.1. Cross-Platform System Tray
**Challenge**: Different system tray implementations on macOS vs Windows
**Solution**: Use Wails built-in system tray support with platform-specific menu handling

### 8.2. Real-time State Synchronization
**Challenge**: Keeping frontend UI in sync with device state changes
**Solution**: Implement event-driven architecture with Wails context events

```go
// Backend event emission
func (a *App) emitDeviceUpdate(device *DeviceInfo) {
    runtime.EventsEmit(a.ctx, "device:updated", device)
}

// Frontend event listening
useEffect(() => {
  const unsubscribe = EventsOn("device:updated", (device: DeviceInfo) => {
    setDevices(prev => prev.map(d => d.id === device.id ? device : d));
  });
  
  return unsubscribe;
}, []);
```

### 8.3. UDP Network Access
**Challenge**: Ensuring UDP socket access works across different network configurations
**Solution**: Implement robust network interface detection and fallback mechanisms

### 8.4. Performance Optimization
**Challenge**: Maintaining responsive UI during network operations
**Solution**: Use Go goroutines for network operations and React concurrent features

## 9. Distribution & Deployment

### 9.1. Build Configuration
```json
{
  "name": "PPA Speaker Control",
  "outputfilename": "ppa-speaker-control",
  "frontend": {
    "dir": "./frontend",
    "install": "npm install",
    "build": "npm run build"
  },
  "backend": {
    "dir": "./backend"
  },
  "author": {
    "name": "Your Name",
    "email": "your.email@example.com"
  },
  "info": {
    "companyName": "Your Company",
    "productName": "PPA Speaker Control",
    "productVersion": "1.0.0",
    "copyright": "Copyright © 2024",
    "comments": "Professional DSP speaker control application"
  }
}
```

### 9.2. macOS Distribution
- **Code Signing**: Required for distribution outside App Store
- **Notarization**: Required for macOS Gatekeeper compatibility
- **DMG Creation**: Professional installer package
- **Auto-updater**: Built-in update mechanism

### 9.3. Windows Distribution
- **Code Signing**: Authenticode certificate for trust
- **MSI Installer**: Professional Windows installer
- **Auto-updater**: Integrated update system
- **Windows Store**: Optional distribution channel

## 10. Security & Privacy

### 10.1. Network Security
- Local network communication only (no internet required)
- UDP packet validation and sanitization
- Device authentication (if supported by hardware)
- Network interface restrictions

### 10.2. System Permissions
- **macOS**: Accessibility permissions for global hotkeys
- **Windows**: No special permissions required
- **Network**: Local network access only
- **File System**: Configuration files in user directory only

## 11. Future Enhancements

### 11.1. Advanced Features
- **Multi-room Audio**: Zone-based control and grouping
- **Scheduling**: Automated preset changes based on time
- **Remote Access**: Secure remote control via VPN
- **Plugin System**: Third-party integrations

### 11.2. Integration Possibilities
- **Home Automation**: Integration with HomeKit, Alexa, Google Home
- **Music Services**: Direct integration with Spotify, Apple Music
- **Professional Tools**: Integration with DAWs and mixing consoles
- **Mobile Companion**: Sync with mobile app for unified control

---

## Summary

This native desktop application will provide professional-grade control of PPA DSP speaker systems through an intuitive menubar/taskbar interface. By leveraging Wails v2, we combine the performance and network capabilities of Go with a modern, responsive web-based UI, creating a seamless native experience that feels at home on both macOS and Windows.

The application prioritizes quick access to essential controls while providing comprehensive management capabilities when needed. The modular architecture ensures maintainability and extensibility, while the robust error handling and offline capabilities make it suitable for professional audio environments where reliability is paramount.

Key differentiators:
- **Native Performance**: Direct UDP communication without browser limitations
- **System Integration**: True menubar/taskbar experience with global shortcuts
- **Professional Focus**: Designed for live audio environments and professional use
- **Cross-Platform**: Single codebase for macOS and Windows
- **Extensible**: Built on proven PPA control system architecture
``` 