package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/jochenvg/go-udev"
)

type DeviceInfo struct {
	Identification struct {
		Serial       string
		Vendor       string
		VendorID     string
		Model        string
		ModelID      string
		Manufacturer string
	}
	USBDetails struct {
		Version  string
		Class    string
		Subclass string
		Protocol string
		Driver   string
		Speed    string
	}
	DevicePaths struct {
		DevPath string
		DevNode string
	}
	StorageInfo struct {
		Size       string
		DiskSize   string
		Filesystem string
	}
	Event struct {
		Action string
		Type   string
	}
}

func NewDeviceInfo(properties map[string]string) *DeviceInfo {
	d := &DeviceInfo{}

	// Helper function to get property if it exists
	get := func(key string) string {
		if val, exists := properties[key]; exists && val != "" {
			return val
		}
		return ""
	}

	// Identification
	d.Identification.Serial = get("ID_SERIAL_SHORT")
	d.Identification.Vendor = get("ID_VENDOR_FROM_DATABASE")
	d.Identification.VendorID = get("ID_VENDOR_ID")
	d.Identification.Model = get("ID_MODEL_FROM_DATABASE")
	d.Identification.ModelID = get("ID_MODEL_ID")
	d.Identification.Manufacturer = get("ID_MANUFACTURER")

	// USB Details
	d.USBDetails.Version = get("ID_USB_VERSION")
	d.USBDetails.Class = get("ID_USB_CLASS_FROM_DATABASE")
	d.USBDetails.Subclass = get("ID_USB_SUBCLASS_FROM_DATABASE")
	d.USBDetails.Protocol = get("ID_USB_PROTOCOL_FROM_DATABASE")
	d.USBDetails.Driver = get("DRIVER")
	d.USBDetails.Speed = get("SPEED")

	// Device Paths
	d.DevicePaths.DevPath = get("DEVPATH")
	d.DevicePaths.DevNode = get("DEVNAME")

	// Storage Info
	d.StorageInfo.Size = get("ID_PART_ENTRY_SIZE")
	d.StorageInfo.DiskSize = get("ID_PART_ENTRY_DISK")
	d.StorageInfo.Filesystem = get("ID_FS_TYPE")

	// Event Info
	d.Event.Action = get("ACTION")
	d.Event.Type = get("DEVTYPE")

	return d
}

func (d *DeviceInfo) FriendlyName() string {
	var parts []string

	// Prefer manufacturer name if available
	if d.Identification.Manufacturer != "" {
		parts = append(parts, d.Identification.Manufacturer)
	} else if d.Identification.Vendor != "" {
		parts = append(parts, d.Identification.Vendor)
	}

	// Add model if available
	if d.Identification.Model != "" {
		parts = append(parts, d.Identification.Model)
	}

	// Fallback to vendor ID if no other names available
	if len(parts) == 0 && d.Identification.VendorID != "" {
		parts = append(parts, "Vendor "+d.Identification.VendorID)
	}

	// Fallback to just "USB Device" if we have nothing else
	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, ", ")
}

func (d *DeviceInfo) FriendlyMessage() string {
	if d.Event.Action == "remove" {
		return fmt.Sprintf("USB Device disconnected")
	}
	deviceName := d.FriendlyName()
	return fmt.Sprintf("USB Device connected: %s", deviceName)
}

func (d *DeviceInfo) String() string {
	var sb strings.Builder
	w := tabwriter.NewWriter(&sb, 0, 0, 2, ' ', 0)

	fmt.Fprintf(w, "===== %s =====\n", d.FriendlyMessage())
	fmt.Fprintf(w, "Event Type:\t%s\n", d.Event.Type)

	if d.Identification.Serial != "" {
		fmt.Fprintln(w, "\n[Identification]")
		if d.Identification.Manufacturer != "" {
			fmt.Fprintf(w, "Manufacturer:\t%s\n", d.Identification.Manufacturer)
		}
		if d.Identification.Vendor != "" {
			fmt.Fprintf(w, "Vendor:\t%s\n", d.Identification.Vendor)
		}
		if d.Identification.Model != "" {
			fmt.Fprintf(w, "Model:\t%s\n", d.Identification.Model)
		}
		fmt.Fprintf(w, "Serial:\t%s\n", d.Identification.Serial)
	}

	if d.USBDetails.Version != "" || d.USBDetails.Class != "" {
		fmt.Fprintln(w, "\n[USB Details]")
		if d.USBDetails.Version != "" {
			fmt.Fprintf(w, "Version:\t%s\n", d.USBDetails.Version)
		}
		if d.USBDetails.Class != "" {
			fmt.Fprintf(w, "Class/Subclass/Protocol:\t%s/%s/%s\n",
				d.USBDetails.Class, d.USBDetails.Subclass, d.USBDetails.Protocol)
		}
		if d.USBDetails.Driver != "" {
			fmt.Fprintf(w, "Driver:\t%s\n", d.USBDetails.Driver)
		}
	}

	if d.DevicePaths.DevPath != "" {
		fmt.Fprintln(w, "\n[Device Paths]")
		fmt.Fprintf(w, "System Path:\t%s\n", d.DevicePaths.DevPath)
		if d.DevicePaths.DevNode != "" {
			fmt.Fprintf(w, "Device Node:\t%s\n", d.DevicePaths.DevNode)
		}
	}

	if d.Event.Type == "partition" || d.StorageInfo.Filesystem != "" {
		fmt.Fprintln(w, "\n[Storage Information]")
		if d.StorageInfo.Size != "" {
			fmt.Fprintf(w, "Size:\t%s sectors\n", d.StorageInfo.Size)
		}
		if d.StorageInfo.DiskSize != "" {
			fmt.Fprintf(w, "Disk Size:\t%s\n", d.StorageInfo.DiskSize)
		}
		if d.StorageInfo.Filesystem != "" {
			fmt.Fprintf(w, "Filesystem:\t%s\n", d.StorageInfo.Filesystem)
		}
	}

	w.Flush()
	return sb.String()
}

func monitorUSBEvents(ctx context.Context) (<-chan *DeviceInfo, error) {
	u := udev.Udev{}
	m := u.NewMonitorFromNetlink("udev")
	_ = m.FilterAddMatchSubsystem("usb")
	_ = m.FilterAddMatchSubsystem("block")
	_ = m.FilterAddMatchTag("seat")

	devChan, errChan, err := m.DeviceChan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create device channel: %v", err)
	}

	outputChan := make(chan *DeviceInfo)
	go func() {
		defer close(outputChan)
		for {
			select {
			case <-ctx.Done():
				return
			case err := <-errChan:
				fmt.Fprintf(os.Stderr, "Monitor error: %v\n", err)
				return
			case udevDev := <-devChan:
				outputChan <- NewDeviceInfo(udevDev.Properties())
			}
		}
	}()

	return outputChan, nil
}
