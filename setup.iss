[Setup]
; 安装程序的基本信息
AppName=Tank Game
AppVersion=1.0
WizardStyle=modern
DefaultDirName={autopf}\Tank Game
DefaultGroupName=Tank Game
OutputDir=.
OutputBaseFilename=TankGame-Setup

[Files]
; 将可执行文件和其他资源文件添加到安装包中
Source: "TankGame.exe"; DestDir: "{app}"
Source: "Readme.txt"; DestDir: "{app}"; Flags: isreadme
Source: "fonts\STSONG.ttf"; DestDir: "{app}\fonts"

[Icons]
; 创建桌面和开始菜单快捷方式
Name: "{group}\Tank Game"; Filename: "{app}\TankGame.exe"
Name: "{commondesktop}\Tank Game"; Filename: "{app}\TankGame.exe"

[Run]
; 安装完成后运行程序
Filename: "{app}\TankGame.exe"; Description: "{cm:LaunchProgram,Tank Game}"; Flags: nowait postinstall skipifsilent
