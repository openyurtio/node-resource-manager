package err

import "fmt"

type ExistsFormatErr struct {
	FsType         string
	ExistingFormat string
	MountErr       error
}

func (e *ExistsFormatErr) Error() string {
	return fmt.Sprintf("Failed to mount the volume as :%v, volume has already contains %v, volume Mount error: %v", e.FsType, e.ExistingFormat, e.MountErr)

}

type DeviceNotExistsErr struct {
	Device string
}

func (e *DeviceNotExistsErr) Error() string {
	return fmt.Sprintf("Device [%s] not exists in current node", e.Device)
}
