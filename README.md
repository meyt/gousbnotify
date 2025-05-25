# gousbnotify

![AI Assisted Badge](https://img.shields.io/badge/AI-ASSISTED-purple)

USB connect/disconnect notifications for Linux.

- Show pop-up notification
- Play custom notification sound
- Portable, just single binary (default sounds embedded)
- No Distro or DE restriction

![image](https://github.com/user-attachments/assets/c2cf19c1-9b85-4426-8843-09e89b785dd8)

## Install

```bash
mkdir ~/.local/bin
cd ~/.local/bin
wget -O gousbnotify https://github.com/meyt/gousbnotify/releases/latest/download/gousbnotify-linux-amd64
chmod +x gousbnotify
gousbnotify -install
```

Also you can run in foreground without installing, just run `gousbnotify`.

## Parameters

All parameters can be used with or without `-install`

- `-nosound`: Disables the sound notification
- `-nonotif`: Disables the pop-up notification.
- `-connect-sound`: Sound file path for "Connect".
- `-disconnect-sound`: Sound file path for "Disconnect".

For example:

```bash
gousbnotify -install \
    -connect-sound="/mnt/drivec/Windows/Media/Windows Hardware Insert.wav"
    -disconnect-sound="/mnt/drivec/Windows/Media/Windows Hardware Remove.wav"
```

## Uninstall

```bash
cd ~/.local/bin
gousbnotify -uninstall
```

## Credits

- Default Sounds: https://cgit.freedesktop.org/sound-theme-freedesktop/
- Assisted AI agents: DeepSeek, Grok, Perplexity, Gemini
