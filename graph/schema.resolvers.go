package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dougedey/elsinore/devices"
	"github.com/dougedey/elsinore/graph/generated"
)

func (r *hysteriaSettingsResolver) ID(ctx context.Context, obj *devices.HysteriaSettings) (string, error) {
	return fmt.Sprint(obj.ID), nil
}

func (r *manualSettingsResolver) ID(ctx context.Context, obj *devices.ManualSettings) (string, error) {
	return fmt.Sprint(obj.ID), nil
}

func (r *mutationResolver) AssignProbe(ctx context.Context, name *string, address *string) (*devices.TemperatureController, error) {
	probe := devices.GetTemperature(*address)
	if probe != nil {
		return devices.CreateTemperatureController(*name, probe)
	}
	return nil, fmt.Errorf("could not find a probe for %v", *address)
}

func (r *mutationResolver) RemoveProbeFromController(ctx context.Context, address *string) (*devices.TemperatureController, error) {
	controller := devices.FindTemperatureControllerForProbe(*address)
	if controller == nil {
		return nil, fmt.Errorf("no controller could be found for %v", address)
	}
	error := controller.RemoveProbe(*address)
	return controller, error
}

func (r *pidSettingsResolver) ID(ctx context.Context, obj *devices.PidSettings) (string, error) {
	return fmt.Sprint(obj.ID), nil
}

func (r *queryResolver) Probe(ctx context.Context, address *string) (*devices.TemperatureProbe, error) {
	device := devices.GetTemperature(*address)
	if device != nil {
		return device, nil
	}
	return nil, fmt.Errorf("no device found for address %v", *address)
}

func (r *queryResolver) ProbeList(ctx context.Context) ([]*devices.TemperatureProbe, error) {
	return devices.GetProbes(), nil
}

func (r *queryResolver) FetchProbes(ctx context.Context, addresses []*string) ([]*devices.TemperatureProbe, error) {
	deviceList := []*devices.TemperatureProbe{}
	missingAddresses := []string{}
	for _, address := range addresses {
		device := devices.GetTemperature(*address)
		if device != nil {
			deviceList = append(deviceList, device)
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

func (r *temperatureControllerResolver) ID(ctx context.Context, obj *devices.TemperatureController) (string, error) {
	return strconv.FormatUint(uint64(obj.ID), 10), nil
}

func (r *temperatureProbeResolver) ID(ctx context.Context, obj *devices.TemperatureProbe) (string, error) {
	return fmt.Sprint(obj.ID), nil
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

// TemperatureController returns generated.TemperatureControllerResolver implementation.
func (r *Resolver) TemperatureController() generated.TemperatureControllerResolver {
	return &temperatureControllerResolver{r}
}

// TemperatureProbe returns generated.TemperatureProbeResolver implementation.
func (r *Resolver) TemperatureProbe() generated.TemperatureProbeResolver {
	return &temperatureProbeResolver{r}
}

type hysteriaSettingsResolver struct{ *Resolver }
type manualSettingsResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type pidSettingsResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type temperatureControllerResolver struct{ *Resolver }
type temperatureProbeResolver struct{ *Resolver }
