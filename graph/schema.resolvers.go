package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dougedey/elsinore/devices"
	"github.com/dougedey/elsinore/graph/generated"
	"github.com/dougedey/elsinore/graph/model"
	"github.com/dougedey/elsinore/hardware"
	"github.com/dougedey/elsinore/system"
)

func (r *hysteriaSettingsResolver) ID(ctx context.Context, obj *devices.HysteriaSettings) (string, error) {
	return fmt.Sprint(obj.ID), nil
}

func (r *manualSettingsResolver) ID(ctx context.Context, obj *devices.ManualSettings) (string, error) {
	return fmt.Sprint(obj.ID), nil
}

func (r *mutationResolver) AssignProbe(ctx context.Context, name string, address string) (*devices.TemperatureController, error) {
	probe := hardware.GetTemperature(address)
	if probe != nil {
		probeDetails := devices.TempProbeDetail{FriendlyName: probe.PhysAddr, PhysAddr: probe.PhysAddr}
		probeDetails.UpdateReading()
		return devices.CreateTemperatureController(name, &probeDetails)
	}
	return nil, fmt.Errorf("could not find a probe for %v", address)
}

func (r *mutationResolver) RemoveProbeFromTemperatureController(ctx context.Context, address string) (*devices.TemperatureController, error) {
	controller := devices.FindTemperatureControllerForProbe(address)
	if controller == nil {
		return nil, fmt.Errorf("no controller could be found for %v", address)
	}
	error := controller.RemoveProbe(address)
	return controller, error
}

func (r *mutationResolver) UpdateTemperatureController(ctx context.Context, controllerSettings model.TemperatureControllerSettingsInput) (*devices.TemperatureController, error) {
	controller := devices.FindTemperatureControllerByID(controllerSettings.ID)
	if controller == nil {
		return nil, fmt.Errorf("no controller could be found for: %v", controllerSettings.ID)
	}

	err := controller.ApplySettings(controllerSettings)
	return controller, err
}

func (r *mutationResolver) DeleteTemperatureController(ctx context.Context, id string) (*model.DeleteTemperatureControllerReturnType, error) {
	probeList := devices.DeleteTemperatureControllerByID(id)
	if probeList == nil {
		return nil, fmt.Errorf("failed to find a controller to delete for: %v", id)
	}
	controllerReturn := model.DeleteTemperatureControllerReturnType{ID: id, TemperatureProbes: probeList}
	return &controllerReturn, nil
}

func (r *mutationResolver) UpdateSettings(ctx context.Context, settings model.SettingsInput) (*system.Settings, error) {
	if settings.BreweryName != nil {
		system.CurrentSettings().BreweryName = *settings.BreweryName
	}
	return system.CurrentSettings(), nil
}

func (r *mutationResolver) ModifySwitch(ctx context.Context, switchSettings model.SwitchSettingsInput) (*devices.Switch, error) {
	var curSwitch *devices.Switch
	if switchSettings.ID == nil {
		errors := []string{}
		if switchSettings.Gpio == nil || len(strings.TrimSpace(*switchSettings.Gpio)) == 0 {
			errors = append(errors, "GPIO is required when creating a new switch")
		}
		if switchSettings.Name == nil || len(strings.TrimSpace(*switchSettings.Name)) == 0 {
			errors = append(errors, "Name is required when creating a new switch")
		}
		if len(errors) > 0 {
			return nil, fmt.Errorf(strings.Join(errors, "\n"))
		}
		newSwitch, err := devices.CreateSwitch(*switchSettings.Gpio, *switchSettings.Name)
		if err != nil {
			return nil, err
		}
		curSwitch = newSwitch
	} else {
		curSwitch = devices.FindSwitchByID(*switchSettings.ID)
		if curSwitch == nil {
			return nil, fmt.Errorf("no switch with id: %v found", *switchSettings.ID)
		}
	}

	if switchSettings.Name != nil {
		curSwitch.Output.FriendlyName = *switchSettings.Name
	}
	if switchSettings.Gpio != nil {
		curSwitch.Output.Identifier = *switchSettings.Gpio
		curSwitch.Reset()
	}
	return curSwitch, nil
}

func (r *mutationResolver) ToggleSwitch(ctx context.Context, id string, mode model.SwitchMode) (*devices.Switch, error) {
	s := devices.FindSwitchByID(id)
	if strings.EqualFold("on", mode.String()) {
		s.On()
	} else {
		s.Off()
	}
	return s, nil
}

func (r *pidSettingsResolver) ID(ctx context.Context, obj *devices.PidSettings) (string, error) {
	return fmt.Sprint(obj.ID), nil
}

func (r *queryResolver) Probe(ctx context.Context, address *string) (*model.TemperatureProbe, error) {
	device := hardware.GetTemperature(*address)
	if device != nil {
		reading := device.Reading()
		return &model.TemperatureProbe{PhysAddr: &device.PhysAddr, Reading: &reading, Updated: &device.Updated}, nil
	}
	return nil, fmt.Errorf("no device found for address %v", *address)
}

func (r *queryResolver) ProbeList(ctx context.Context, available *bool) ([]*model.TemperatureProbe, error) {
	probeList := []*model.TemperatureProbe{}
	for _, device := range hardware.GetProbes() {
		if available != nil && *available && devices.FindTemperatureControllerForProbe(device.PhysAddr) != nil {
			continue
		}
		reading := device.Reading()
		probeList = append(probeList, &model.TemperatureProbe{PhysAddr: &device.PhysAddr, Reading: &reading, Updated: &device.Updated})
	}
	return probeList, nil
}

func (r *queryResolver) FetchProbes(ctx context.Context, addresses []*string) ([]*model.TemperatureProbe, error) {
	deviceList := []*model.TemperatureProbe{}
	missingAddresses := []string{}
	for _, address := range addresses {
		device := hardware.GetTemperature(*address)
		if device != nil {
			reading := device.Reading()
			deviceList = append(deviceList, &model.TemperatureProbe{PhysAddr: &device.PhysAddr, Reading: &reading, Updated: &device.Updated})
		} else {
			missingAddresses = append(missingAddresses, *address)
		}
	}

	missingError := (error)(nil)
	if len(missingAddresses) > 0 {
		missingError = fmt.Errorf("no device(s) found for address(es): %v", missingAddresses)
	}

	return deviceList, missingError
}

func (r *queryResolver) TemperatureControllers(ctx context.Context, name *string) ([]*devices.TemperatureController, error) {
	if name == nil {
		return devices.AllTemperatureControllers(), nil
	}
	controller := devices.FindTemperatureControllerByName(*name)
	if controller == nil {
		return nil, fmt.Errorf("no controller could be found for %v", *name)
	}
	return []*devices.TemperatureController{controller}, nil
}

func (r *queryResolver) Settings(ctx context.Context) (*system.Settings, error) {
	return system.CurrentSettings(), nil
}

func (r *queryResolver) Switches(ctx context.Context) ([]*devices.Switch, error) {
	return devices.AllSwitches(), nil
}

func (r *switchResolver) ID(ctx context.Context, obj *devices.Switch) (string, error) {
	return strconv.FormatUint(uint64(obj.ID), 10), nil
}

func (r *temperatureControllerResolver) ID(ctx context.Context, obj *devices.TemperatureController) (string, error) {
	return strconv.FormatUint(uint64(obj.ID), 10), nil
}

func (r *temperatureControllerResolver) TempProbeDetails(ctx context.Context, obj *devices.TemperatureController) ([]*model.TempProbeDetails, error) {
	probeList := []*model.TempProbeDetails{}
	for _, tempProbe := range obj.TempProbeDetails {
		reading := tempProbe.Reading()
		probeDetail := model.TempProbeDetails{ID: fmt.Sprint(tempProbe.ID), PhysAddr: &tempProbe.PhysAddr, Reading: &reading, Name: &tempProbe.FriendlyName, Updated: &tempProbe.Updated}
		probeList = append(probeList, &probeDetail)
	}
	return probeList, nil
}

// HysteriaSettings returns generated.HysteriaSettingsResolver implementation.
func (r *Resolver) HysteriaSettings() generated.HysteriaSettingsResolver {
	return &hysteriaSettingsResolver{r}
}

// ManualSettings returns generated.ManualSettingsResolver implementation.
func (r *Resolver) ManualSettings() generated.ManualSettingsResolver {
	return &manualSettingsResolver{r}
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// PidSettings returns generated.PidSettingsResolver implementation.
func (r *Resolver) PidSettings() generated.PidSettingsResolver { return &pidSettingsResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Switch returns generated.SwitchResolver implementation.
func (r *Resolver) Switch() generated.SwitchResolver { return &switchResolver{r} }

// TemperatureController returns generated.TemperatureControllerResolver implementation.
func (r *Resolver) TemperatureController() generated.TemperatureControllerResolver {
	return &temperatureControllerResolver{r}
}

type hysteriaSettingsResolver struct{ *Resolver }
type manualSettingsResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type pidSettingsResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type switchResolver struct{ *Resolver }
type temperatureControllerResolver struct{ *Resolver }
